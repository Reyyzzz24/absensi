// Package attendance implements the single, unified check-in/out flow that
// replaces the legacy PresensiController (office + WFH) and
// AbsensiController/EOS flows, which ran in parallel against the same table
// with divergent cooldown/overnight rules -- docs/DECISIONS.md D-8.
package attendance

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/platform/geo"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

var (
	ErrOutsideGeofence   = errors.New("outside allowed check-in radius")
	ErrAlreadyCheckedOut = errors.New("already checked out for this shift")
	ErrCooldownActive    = errors.New("too soon after previous checkout to start a new cycle")
)

// minCycleGap is the minimum time between a checkout and the next check-in
// cycle for the same employee+day, to absorb accidental double-taps/retries
// while still allowing legitimate same-day multi-cycle attendance (e.g. the
// legacy GA department's security-guard rotations, D-23). Legacy used
// inconsistent values (30min in PresensiController, 5min in AbsensiController
// -- LOGIC_SPEC.md §5); 5 minutes is chosen here as a safe engineering
// default, not a re-litigated business decision -- easy to make
// admin-configurable later if a stricter/looser value is needed.
const minCycleGap = 5 * time.Minute

var jakarta = mustLoadLocation("Asia/Jakarta")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		// Falls back to fixed UTC+7 if the tzdata isn't available in the
		// runtime environment, rather than silently using server-local time
		// (legacy relied on config/app.php timezone -- LOGIC_SPEC.md §1).
		return time.FixedZone("WIB", 7*60*60)
	}
	return loc
}

type Service struct {
	db          *gorm.DB
	locations   repo.OfficeLocationRepo
	assignments repo.FieldAssignmentRepo
	schedules   repo.ShiftScheduleRepo
}

func NewService(db *gorm.DB, locations repo.OfficeLocationRepo, assignments repo.FieldAssignmentRepo, schedules repo.ShiftScheduleRepo) Service {
	return Service{db: db, locations: locations, assignments: assignments, schedules: schedules}
}

type CheckInInput struct {
	EmployeeID int64
	Latitude   float64
	Longitude  float64
	PhotoPath  string
	IsWFH      bool
}

// CheckInOrOut is the single entry point for both check-in and check-out.
// Which one happens is derived server-side from existing attendance state,
// never from client input:
//  1. If yesterday's row is still open on an overnight shift (D-14), this
//     call closes that row (checkout) rather than starting a new one.
//  2. Otherwise, the latest cycle for today is looked up: none yet ->
//     check-in (cycle 1); open -> checkout; closed and past the cooldown ->
//     a new cycle's check-in (cycle N+1, supports multi-cycle days per D-23);
//     closed and still within the cooldown -> ErrCooldownActive.
//
// Concurrent submissions are serialized with SELECT ... FOR UPDATE inside a
// transaction, backed by the UNIQUE(employee_id, work_date, cycle_number)
// constraint on attendances -- replacing the legacy check-then-act race
// condition (D-9).
func (s Service) CheckInOrOut(ctx context.Context, in CheckInInput) (*domain.Attendance, error) {
	now := time.Now().In(jakarta)
	today := civilDate(now)
	yesterday := today.AddDate(0, 0, -1)

	var result domain.Attendance
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var openOvernight domain.Attendance
		err := tx.
			Joins("JOIN shifts ON shifts.id = attendances.shift_id").
			Where("attendances.employee_id = ? AND attendances.work_date = ? AND attendances.status = ? AND shifts.is_overnight = true",
				in.EmployeeID, yesterday, domain.AttendanceStatusOpen).
			Order("attendances.cycle_number DESC").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&openOvernight).Error
		switch {
		case err == nil:
			if cerr := s.checkOut(ctx, tx, &openOvernight, now, in); cerr != nil {
				return cerr
			}
			result = openOvernight
			return nil
		case errors.Is(err, gorm.ErrRecordNotFound):
			// fall through to today's latest cycle
		default:
			return err
		}

		var latest domain.Attendance
		err = tx.
			Where("employee_id = ? AND work_date = ?", in.EmployeeID, today).
			Order("cycle_number DESC").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&latest).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			row, cerr := s.checkIn(ctx, tx, today, 1, now, in)
			if cerr != nil {
				return cerr
			}
			result = *row
			return nil
		case err != nil:
			return err
		}

		if latest.CheckOutAt == nil {
			if cerr := s.checkOut(ctx, tx, &latest, now, in); cerr != nil {
				return cerr
			}
			result = latest
			return nil
		}

		if now.Sub(*latest.CheckOutAt) < minCycleGap {
			return ErrCooldownActive
		}

		row, cerr := s.checkIn(ctx, tx, today, latest.CycleNumber+1, now, in)
		if cerr != nil {
			return cerr
		}
		result = *row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s Service) checkIn(ctx context.Context, tx *gorm.DB, workDate time.Time, cycleNumber int, now time.Time, in CheckInInput) (*domain.Attendance, error) {
	shift, err := s.schedules.ResolveShift(ctx, in.EmployeeID, workDate)
	if err != nil {
		return nil, err
	}

	locationID, distance, err := s.enforceGeofence(ctx, in, workDate)
	if err != nil {
		return nil, err
	}

	var isLate *bool
	// Only the day's first cycle is evaluated against the shift start time;
	// later cycles (multi-cycle days, D-23) aren't "late" in the same sense.
	if cycleNumber == 1 && shift != nil && shift.StartTime != nil {
		late := isAfterGrace(now, *shift.StartTime, shift.LateGraceMinutes)
		isLate = &late
	}

	var shiftID *int64
	if shift != nil {
		shiftID = &shift.ID
	}

	row := domain.Attendance{
		EmployeeID:       in.EmployeeID,
		WorkDate:         workDate,
		CycleNumber:      cycleNumber,
		ShiftID:          shiftID,
		IsWFH:            in.IsWFH,
		OfficeLocationID: locationID,
		CheckInAt:        &now,
		CheckInLat:       &in.Latitude,
		CheckInLng:       &in.Longitude,
		CheckInDistanceM: distance,
		CheckInPhotoPath: &in.PhotoPath,
		IsLate:           isLate,
		Status:           domain.AttendanceStatusOpen,
	}

	if err := tx.Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s Service) checkOut(ctx context.Context, tx *gorm.DB, row *domain.Attendance, now time.Time, in CheckInInput) error {
	_, distance, err := s.enforceGeofence(ctx, in, row.WorkDate)
	if err != nil {
		return err
	}

	if row.ShiftID != nil {
		var shift domain.Shift
		if err := tx.First(&shift, *row.ShiftID).Error; err == nil && shift.EndTime != nil {
			early := isBeforeTimeOfDay(now, *shift.EndTime)
			row.IsEarlyLeave = &early
		}
	}

	row.CheckOutAt = &now
	row.CheckOutLat = &in.Latitude
	row.CheckOutLng = &in.Longitude
	row.CheckOutDistanceM = distance
	row.CheckOutPhotoPath = &in.PhotoPath
	row.Status = domain.AttendanceStatusClosed

	return tx.Save(row).Error
}

// enforceGeofence returns the matched office_location ID and distance in
// meters, or ErrOutsideGeofence if the point is outside every active
// location's radius. Bypassed for WFH or an approved field assignment
// (dinas luar, D-21) -- distance is still computed/recorded when possible for
// audit purposes, but not enforced in that case.
func (s Service) enforceGeofence(ctx context.Context, in CheckInInput, workDate time.Time) (*int64, *int, error) {
	bypass := in.IsWFH
	if !bypass {
		assigned, err := s.assignments.Exists(ctx, in.EmployeeID, workDate)
		if err != nil {
			return nil, nil, err
		}
		bypass = assigned
	}

	locations, err := s.locations.ActiveLocations(ctx)
	if err != nil {
		return nil, nil, err
	}

	var closestID *int64
	var closestDistance *int
	for i := range locations {
		loc := locations[i]
		d := int(geo.DistanceMeters(in.Latitude, in.Longitude, loc.Latitude, loc.Longitude))
		if closestDistance == nil || d < *closestDistance {
			id := loc.ID
			closestID = &id
			closestDistance = &d
		}
		if d <= loc.RadiusMeters {
			return &loc.ID, &d, nil
		}
	}

	if bypass {
		return closestID, closestDistance, nil
	}
	return nil, nil, ErrOutsideGeofence
}

// ActiveGeofences exposes the active office locations (id, name,
// coordinates, radius) to the CHECKING-IN EMPLOYEE, not just admins --
// needed so the check-in page can draw the geofence circle on a map and
// show a live "inside/outside radius" hint before submitting (display-only;
// the actual enforcement in enforceGeofence above is what actually counts,
// D-1). Deliberately the same read as admin's OfficeLocationRepo.List
// (active-only), just reachable from the employee-audience route too --
// there is nothing admin-sensitive in an office's public coordinates.
func (s Service) ActiveGeofences(ctx context.Context) ([]domain.OfficeLocation, error) {
	return s.locations.ActiveLocations(ctx)
}

func civilDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// isAfterGrace reports whether `now`'s time-of-day is later than
// startTime + graceMinutes. startTime is "HH:MM:SS".
func isAfterGrace(now time.Time, startTime string, graceMinutes int) bool {
	start, err := parseTimeOfDay(now, startTime)
	if err != nil {
		return false
	}
	return now.After(start.Add(time.Duration(graceMinutes) * time.Minute))
}

func isBeforeTimeOfDay(now time.Time, hhmmss string) bool {
	t, err := parseTimeOfDay(now, hhmmss)
	if err != nil {
		return false
	}
	return now.Before(t)
}

func parseTimeOfDay(base time.Time, hhmmss string) (time.Time, error) {
	parsed, err := time.ParseInLocation("15:04:05", hhmmss, base.Location())
	if err != nil {
		return time.Time{}, err
	}
	y, m, d := base.Date()
	return time.Date(y, m, d, parsed.Hour(), parsed.Minute(), parsed.Second(), 0, base.Location()), nil
}
