// Package config implements admin-facing configuration for shifts, office
// locations, shift schedules, and field assignments ("dinas luar", D-21).
// Legacy equivalent: KonfigurasiController.
package config

import (
	"context"
	"errors"
	"time"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

var ErrNotFound = errors.New("not found")

type Service struct {
	shifts       repo.ShiftRepo
	weeklyShifts repo.WeeklyShiftDefaultRepo
	workSchedule repo.WorkScheduleRepo
	locations    repo.OfficeLocationRepo
	assignments  repo.FieldAssignmentRepo
	company      repo.CompanySettingsRepo
}

func NewService(
	shifts repo.ShiftRepo,
	weeklyShifts repo.WeeklyShiftDefaultRepo,
	workSchedule repo.WorkScheduleRepo,
	locations repo.OfficeLocationRepo,
	assignments repo.FieldAssignmentRepo,
	company repo.CompanySettingsRepo,
) Service {
	return Service{
		shifts:       shifts,
		weeklyShifts: weeklyShifts,
		workSchedule: workSchedule,
		locations:    locations,
		assignments:  assignments,
		company:      company,
	}
}

func (s Service) GetCompanySettings(ctx context.Context) (*domain.CompanySettings, error) {
	return s.company.Get(ctx)
}

type CompanySettingsInput struct {
	Name     string
	LogoPath *string // nil = leave unchanged
}

func (s Service) UpdateCompanySettings(ctx context.Context, in CompanySettingsInput) (*domain.CompanySettings, error) {
	existing, err := s.company.Get(ctx)
	if err != nil {
		return nil, err
	}
	existing.Name = in.Name
	if in.LogoPath != nil {
		existing.LogoPath = in.LogoPath
	}
	if err := s.company.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// UpdateWorkingWeekdays sets the company's work week (D-25 holiday
// resolver) -- ISO weekday numbers (1=Monday..7=Sunday) considered work
// days. Deliberately not hardcoded Sat/Sun so a 6-day work week is just a
// different set, not a code change.
func (s Service) UpdateWorkingWeekdays(ctx context.Context, weekdays []int64) (*domain.CompanySettings, error) {
	existing, err := s.company.Get(ctx)
	if err != nil {
		return nil, err
	}
	existing.WorkingWeekdays = weekdays
	if err := s.company.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// --- Shifts (legacy jam_kerja) ---

type ShiftInput struct {
	Code             string
	Name             string
	IsDayOff         bool
	StartTime        string // "HH:MM:SS", empty for day-off shifts
	EndTime          string
	LateGraceMinutes int
}

func (s Service) ListShifts(ctx context.Context) ([]domain.Shift, error) {
	return s.shifts.List(ctx)
}

func (s Service) CreateShift(ctx context.Context, in ShiftInput) (*domain.Shift, error) {
	shift := &domain.Shift{
		Code:             in.Code,
		Name:             in.Name,
		IsDayOff:         in.IsDayOff,
		LateGraceMinutes: in.LateGraceMinutes,
	}
	if in.StartTime != "" {
		shift.StartTime = &in.StartTime
	}
	if in.EndTime != "" {
		shift.EndTime = &in.EndTime
	}
	if err := s.shifts.Create(ctx, shift); err != nil {
		return nil, err
	}
	return shift, nil
}

func (s Service) UpdateShift(ctx context.Context, id int64, in ShiftInput) (*domain.Shift, error) {
	existing, err := s.shifts.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	existing.Name = in.Name
	existing.IsDayOff = in.IsDayOff
	existing.LateGraceMinutes = in.LateGraceMinutes
	if in.StartTime != "" {
		existing.StartTime = &in.StartTime
	} else {
		existing.StartTime = nil
	}
	if in.EndTime != "" {
		existing.EndTime = &in.EndTime
	} else {
		existing.EndTime = nil
	}

	if err := s.shifts.Update(ctx, existing); err != nil {
		return nil, err
	}
	return s.shifts.FindByID(ctx, id)
}

// --- Office locations (legacy konfigurasi_lokasi) ---

type OfficeLocationInput struct {
	Name         string
	Latitude     float64
	Longitude    float64
	RadiusMeters int
	IsActive     bool
}

func (s Service) ListOfficeLocations(ctx context.Context) ([]domain.OfficeLocation, error) {
	return s.locations.List(ctx)
}

// CreateOfficeLocation defaults RadiusMeters to 100 (D-1) when not supplied.
func (s Service) CreateOfficeLocation(ctx context.Context, in OfficeLocationInput) (*domain.OfficeLocation, error) {
	radius := in.RadiusMeters
	if radius <= 0 {
		radius = 100
	}
	loc := &domain.OfficeLocation{
		Name:         in.Name,
		Latitude:     in.Latitude,
		Longitude:    in.Longitude,
		RadiusMeters: radius,
		IsActive:     true,
	}
	if err := s.locations.Create(ctx, loc); err != nil {
		return nil, err
	}
	return loc, nil
}

// UpdateOfficeLocation lets an admin move/resize an existing geofence (drag
// the pin, change the radius) or toggle it active/inactive. Unlike Create,
// this does not force RadiusMeters to a default -- callers already have an
// existing value to fall back on if they don't mean to change it.
func (s Service) UpdateOfficeLocation(ctx context.Context, id int64, in OfficeLocationInput) (*domain.OfficeLocation, error) {
	existing, err := s.locations.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	existing.Name = in.Name
	existing.Latitude = in.Latitude
	existing.Longitude = in.Longitude
	if in.RadiusMeters > 0 {
		existing.RadiusMeters = in.RadiusMeters
	}
	existing.IsActive = in.IsActive

	if err := s.locations.Update(ctx, existing); err != nil {
		return nil, err
	}
	return s.locations.FindByID(ctx, id)
}

// --- Shift schedules (jadwal_kerja / konfigurasi_jamkerja, D-18) ---

func (s Service) SetWorkSchedule(ctx context.Context, employeeID int64, workDate time.Time, shiftID int64) error {
	return s.workSchedule.Upsert(ctx, employeeID, workDate, shiftID)
}

func (s Service) SetWeeklyShiftDefault(ctx context.Context, employeeID int64, dayOfWeek int16, shiftID int64) error {
	return s.weeklyShifts.Upsert(ctx, employeeID, dayOfWeek, shiftID)
}

// --- Field assignments ("dinas luar", D-21) ---

type FieldAssignmentInput struct {
	EmployeeID int64
	WorkDate   time.Time
	Note       string
	ApprovedBy int64
}

func (s Service) CreateFieldAssignment(ctx context.Context, in FieldAssignmentInput) (*domain.FieldAssignment, error) {
	fa := &domain.FieldAssignment{
		EmployeeID: in.EmployeeID,
		WorkDate:   in.WorkDate,
		ApprovedBy: in.ApprovedBy,
	}
	if in.Note != "" {
		fa.Note = &in.Note
	}
	if err := s.assignments.Create(ctx, fa); err != nil {
		return nil, err
	}
	return fa, nil
}

func (s Service) ListFieldAssignments(ctx context.Context, employeeID int64) ([]domain.FieldAssignment, error) {
	return s.assignments.ListByEmployee(ctx, employeeID)
}
