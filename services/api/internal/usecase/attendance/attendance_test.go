package attendance_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
	"github.com/eprisi/absensi-next/services/api/internal/testutil"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/attendance"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/holiday"
)

// newService wires a real attendance.Service against a fresh test database.
func newService(t *testing.T) (attendance.Service, *gorm.DB) {
	t.Helper()
	gdb := testutil.NewDB(t)
	locations := repo.NewOfficeLocationRepo(gdb)
	assignments := repo.NewFieldAssignmentRepo(gdb)
	schedules := repo.NewShiftScheduleRepo(gdb)
	holidays := holiday.NewService(repo.NewCompanySettingsRepo(gdb), repo.NewNationalHolidayRepo(gdb), repo.NewCompanyHolidayRepo(gdb), holiday.NewGoogleCalendarFetcher())
	return attendance.NewService(gdb, locations, assignments, schedules, holidays), gdb
}

func mustCreateEmployee(t *testing.T, gdb *gorm.DB, nik string) domain.Employee {
	t.Helper()
	e := domain.Employee{NIK: nik, FullName: "Test " + nik, PasswordHash: "x", IsActive: true}
	if err := gdb.Create(&e).Error; err != nil {
		t.Fatalf("create employee: %v", err)
	}
	return e
}

func mustCreateOfficeLocation(t *testing.T, gdb *gorm.DB, lat, lng float64, radius int) domain.OfficeLocation {
	t.Helper()
	loc := domain.OfficeLocation{Name: "Kantor Pusat", Latitude: lat, Longitude: lng, RadiusMeters: radius, IsActive: true}
	if err := gdb.Create(&loc).Error; err != nil {
		t.Fatalf("create office location: %v", err)
	}
	return loc
}

// mustCreateShift creates a shift and assigns it to the employee for the
// given work date via a per-date WorkSchedule row.
func mustCreateShift(t *testing.T, gdb *gorm.DB, code, start, end string, graceMinutes int) domain.Shift {
	t.Helper()
	s := domain.Shift{Code: code, Name: code, StartTime: &start, EndTime: &end, LateGraceMinutes: graceMinutes}
	if err := gdb.Create(&s).Error; err != nil {
		t.Fatalf("create shift: %v", err)
	}
	return s
}

func mustAssignShift(t *testing.T, gdb *gorm.DB, employeeID int64, workDate time.Time, shiftID int64) {
	t.Helper()
	ws := domain.WorkSchedule{EmployeeID: employeeID, WorkDate: workDate, ShiftID: shiftID}
	if err := gdb.Create(&ws).Error; err != nil {
		t.Fatalf("assign shift: %v", err)
	}
}

// --- A1: geofencing must actually be enforced (legacy computed distance but
// never rejected anything -- AUDIT_FINDINGS.md A1 / D-1). ---

func TestCheckIn_OutsideRadius_Rejected(t *testing.T) {
	svc, gdb := newService(t)
	ctx := context.Background()
	emp := mustCreateEmployee(t, gdb, "0001")
	mustCreateOfficeLocation(t, gdb, -6.2, 106.8, 100)

	// ~11km away -- far outside a 100m radius.
	_, err := svc.CheckInOrOut(ctx, attendance.CheckInInput{
		EmployeeID: emp.ID,
		Latitude:   -6.3,
		Longitude:  106.8,
		PhotoPath:  "x.png",
	})
	if !errors.Is(err, attendance.ErrOutsideGeofence) {
		t.Fatalf("expected ErrOutsideGeofence, got %v", err)
	}
}

func TestCheckIn_InsideRadius_Accepted(t *testing.T) {
	svc, gdb := newService(t)
	ctx := context.Background()
	emp := mustCreateEmployee(t, gdb, "0001")
	mustCreateOfficeLocation(t, gdb, -6.2, 106.8, 100)

	row, err := svc.CheckInOrOut(ctx, attendance.CheckInInput{
		EmployeeID: emp.ID,
		Latitude:   -6.2,
		Longitude:  106.8,
		PhotoPath:  "x.png",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if row.Status != domain.AttendanceStatusOpen {
		t.Fatalf("expected status open after check-in, got %s", row.Status)
	}
	if row.CheckInDistanceM == nil || *row.CheckInDistanceM > 100 {
		t.Fatalf("expected recorded distance <= 100m, got %v", row.CheckInDistanceM)
	}
}

// --- D-1/D-21: WFH and approved field assignment bypass the radius check. ---

func TestCheckIn_WFH_BypassesGeofence(t *testing.T) {
	svc, gdb := newService(t)
	ctx := context.Background()
	emp := mustCreateEmployee(t, gdb, "0001")
	mustCreateOfficeLocation(t, gdb, -6.2, 106.8, 100)

	_, err := svc.CheckInOrOut(ctx, attendance.CheckInInput{
		EmployeeID: emp.ID,
		Latitude:   -6.3, // far away
		Longitude:  106.8,
		PhotoPath:  "x.png",
		IsWFH:      true,
	})
	if err != nil {
		t.Fatalf("WFH should bypass geofence, got %v", err)
	}
}

func TestCheckIn_FieldAssignment_BypassesGeofence(t *testing.T) {
	svc, gdb := newService(t)
	ctx := context.Background()
	emp := mustCreateEmployee(t, gdb, "0001")
	mustCreateOfficeLocation(t, gdb, -6.2, 106.8, 100)

	today := time.Now().In(mustLoadJakarta())
	civilToday := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	admin := mustCreateAdminUser(t, gdb)
	if err := gdb.Create(&domain.FieldAssignment{EmployeeID: emp.ID, WorkDate: civilToday, ApprovedBy: admin.ID}).Error; err != nil {
		t.Fatalf("create field assignment: %v", err)
	}

	_, err := svc.CheckInOrOut(ctx, attendance.CheckInInput{
		EmployeeID: emp.ID,
		Latitude:   -6.3, // far away, but pre-approved off-site today
		Longitude:  106.8,
		PhotoPath:  "x.png",
	})
	if err != nil {
		t.Fatalf("approved field assignment should bypass geofence, got %v", err)
	}
}

// --- D-9/D-23: idempotency -- a checkout followed immediately by another
// check-in attempt must be rejected until the cooldown passes, not silently
// create a duplicate/racy row. ---

func TestCheckInOrOut_CooldownAfterCheckout(t *testing.T) {
	svc, gdb := newService(t)
	ctx := context.Background()
	emp := mustCreateEmployee(t, gdb, "0001")
	mustCreateOfficeLocation(t, gdb, -6.2, 106.8, 100)

	in := attendance.CheckInInput{EmployeeID: emp.ID, Latitude: -6.2, Longitude: 106.8, PhotoPath: "x.png"}

	// Cycle 1: check-in.
	row, err := svc.CheckInOrOut(ctx, in)
	if err != nil {
		t.Fatalf("check-in failed: %v", err)
	}
	if row.Status != domain.AttendanceStatusOpen {
		t.Fatalf("expected open after first call, got %s", row.Status)
	}

	// Cycle 1: checkout (second call same day -> closes the open row).
	row, err = svc.CheckInOrOut(ctx, in)
	if err != nil {
		t.Fatalf("checkout failed: %v", err)
	}
	if row.Status != domain.AttendanceStatusClosed {
		t.Fatalf("expected closed after second call, got %s", row.Status)
	}

	// Immediately after checkout: a new cycle should be rejected by cooldown.
	_, err = svc.CheckInOrOut(ctx, in)
	if !errors.Is(err, attendance.ErrCooldownActive) {
		t.Fatalf("expected ErrCooldownActive immediately after checkout, got %v", err)
	}
}

// --- D-14: overnight shift -- a check-in the next calendar day should close
// yesterday's still-open row instead of starting a fresh one. ---

func TestCheckInOrOut_OvernightShift_ClosesYesterdaysOpenRow(t *testing.T) {
	svc, gdb := newService(t)
	ctx := context.Background()
	emp := mustCreateEmployee(t, gdb, "0001")
	mustCreateOfficeLocation(t, gdb, -6.2, 106.8, 100)

	// 22:00 -> 06:00 is overnight (end < start), per migration 000004's
	// generated column.
	shift := mustCreateShift(t, gdb, "SH-NIGHT", "22:00:00", "06:00:00", 15)

	today := time.Now().In(mustLoadJakarta())
	civilToday := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	civilYesterday := civilToday.AddDate(0, 0, -1)
	mustAssignShift(t, gdb, emp.ID, civilYesterday, shift.ID)

	// Manually insert an open attendance row dated yesterday, as if the
	// employee checked in last night and hasn't checked out yet.
	checkInTime := civilYesterday.Add(23 * time.Hour) // ~23:00 yesterday
	open := domain.Attendance{
		EmployeeID:  emp.ID,
		WorkDate:    civilYesterday,
		CycleNumber: 1,
		ShiftID:     &shift.ID,
		CheckInAt:   &checkInTime,
		Status:      domain.AttendanceStatusOpen,
	}
	if err := gdb.Create(&open).Error; err != nil {
		t.Fatalf("seed open overnight row: %v", err)
	}

	// A check-in attempt "this morning" should close yesterday's row, not
	// open a new one for today.
	row, err := svc.CheckInOrOut(ctx, attendance.CheckInInput{
		EmployeeID: emp.ID,
		Latitude:   -6.2,
		Longitude:  106.8,
		PhotoPath:  "out.png",
	})
	if err != nil {
		t.Fatalf("expected checkout of yesterday's row to succeed, got %v", err)
	}
	if row.Status != domain.AttendanceStatusClosed {
		t.Fatalf("expected yesterday's row to be closed, got %s", row.Status)
	}
	// WorkDate is a DATE column -- Postgres/GORM round-trips it as UTC
	// midnight regardless of the zone it was written with, so compare the
	// calendar date components rather than full instant equality.
	if row.WorkDate.Format("2006-01-02") != civilYesterday.Format("2006-01-02") {
		t.Fatalf("expected the closed row to be dated yesterday (%s), got %s",
			civilYesterday.Format("2006-01-02"), row.WorkDate.Format("2006-01-02"))
	}

	var count int64
	gdb.Model(&domain.Attendance{}).Where("employee_id = ? AND work_date = ?", emp.ID, civilToday).Count(&count)
	if count != 0 {
		t.Fatalf("expected no new row for today, found %d", count)
	}
}

// --- D-10: lateness is graded against the shift's own grace period, and
// only on the day's first cycle. ---

func TestCheckIn_LateAfterGracePeriod(t *testing.T) {
	svc, gdb := newService(t)
	ctx := context.Background()
	emp := mustCreateEmployee(t, gdb, "0001")
	mustCreateOfficeLocation(t, gdb, -6.2, 106.8, 100)

	// Shift starting at a time definitely in the past today, with a 0-minute
	// grace period, so "now" is always later than start+grace.
	shift := mustCreateShift(t, gdb, "SH-EARLY", "00:00:01", "00:00:02", 0)
	today := time.Now().In(mustLoadJakarta())
	civilToday := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	mustAssignShift(t, gdb, emp.ID, civilToday, shift.ID)

	row, err := svc.CheckInOrOut(ctx, attendance.CheckInInput{
		EmployeeID: emp.ID,
		Latitude:   -6.2,
		Longitude:  106.8,
		PhotoPath:  "x.png",
	})
	if err != nil {
		t.Fatalf("check-in failed: %v", err)
	}
	if row.IsLate == nil || !*row.IsLate {
		t.Fatalf("expected is_late=true, got %v", row.IsLate)
	}
}

// D-25: check-in on a resolved holiday is allowed (default policy) but is
// never "late" and is tagged is_holiday -- there's no shift obligation to
// be late against on a libur day, even if the employee has a shift
// assigned and shows up "late" by the clock.
func TestCheckIn_OnCompanyHoliday_AllowedNotLateAndTagged(t *testing.T) {
	svc, gdb := newService(t)
	ctx := context.Background()
	emp := mustCreateEmployee(t, gdb, "0001")
	mustCreateOfficeLocation(t, gdb, -6.2, 106.8, 100)

	shift := mustCreateShift(t, gdb, "SH-EARLY", "00:00:01", "00:00:02", 0)
	today := time.Now().In(mustLoadJakarta())
	civilToday := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	mustAssignShift(t, gdb, emp.ID, civilToday, shift.ID)

	admin := mustCreateAdminUser(t, gdb)
	if err := gdb.Create(&domain.CompanyHoliday{
		StartDate: civilToday, EndDate: civilToday, Name: "Libur uji coba", Type: domain.CompanyHolidayTypeLibur, CreatedBy: admin.ID,
	}).Error; err != nil {
		t.Fatalf("seed company holiday: %v", err)
	}

	row, err := svc.CheckInOrOut(ctx, attendance.CheckInInput{
		EmployeeID: emp.ID,
		Latitude:   -6.2,
		Longitude:  106.8,
		PhotoPath:  "x.png",
	})
	if err != nil {
		t.Fatalf("expected check-in to be allowed on a holiday, got error: %v", err)
	}
	if row.IsLate != nil {
		t.Fatalf("expected is_late to stay unset on a holiday check-in, got %v", *row.IsLate)
	}
	if !row.IsHoliday {
		t.Fatalf("expected the attendance row to be tagged is_holiday")
	}
}

// mustCreateAdminUser relies on the "superadmin"/"admin" roles that migration
// 000001 seeds into every fresh database -- no need to create one here.
func mustCreateAdminUser(t *testing.T, gdb *gorm.DB) domain.User {
	t.Helper()
	var role domain.Role
	if err := gdb.Where("name = ?", "superadmin").First(&role).Error; err != nil {
		t.Fatalf("find seeded superadmin role: %v", err)
	}
	u := domain.User{Name: "Admin", Username: "admin", PasswordHash: "x", RoleID: role.ID}
	if err := gdb.Create(&u).Error; err != nil {
		t.Fatalf("create admin user: %v", err)
	}
	return u
}

// --- B4: legacy only ever flagged lateness -- no early-leave/overtime
// concept at all. New system flags "pulang cepat" (early leave) on checkout. ---

func TestCheckOut_FlaggedEarlyLeave(t *testing.T) {
	svc, gdb := newService(t)
	ctx := context.Background()
	emp := mustCreateEmployee(t, gdb, "0001")
	mustCreateOfficeLocation(t, gdb, -6.2, 106.8, 100)

	// Shift ends at 23:59:59 today -- "now" (whenever the test runs) is
	// always before that, so checkout should be flagged as early leave.
	shift := mustCreateShift(t, gdb, "SH-LATEEND", "00:00:01", "23:59:59", 0)
	today := time.Now().In(mustLoadJakarta())
	civilToday := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	mustAssignShift(t, gdb, emp.ID, civilToday, shift.ID)

	in := attendance.CheckInInput{EmployeeID: emp.ID, Latitude: -6.2, Longitude: 106.8, PhotoPath: "x.png"}
	if _, err := svc.CheckInOrOut(ctx, in); err != nil {
		t.Fatalf("check-in failed: %v", err)
	}

	row, err := svc.CheckInOrOut(ctx, in) // checkout
	if err != nil {
		t.Fatalf("checkout failed: %v", err)
	}
	if row.IsEarlyLeave == nil || !*row.IsEarlyLeave {
		t.Fatalf("expected is_early_leave=true, got %v", row.IsEarlyLeave)
	}
}

func mustLoadJakarta() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return loc
}
