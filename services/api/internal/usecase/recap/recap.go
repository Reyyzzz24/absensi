// Package recap computes the monthly per-employee attendance grid (legacy
// cetakrekap). Deliberately NOT a SQL pivot -- the legacy query hardcoded
// 31 `MAX(IF(DAY(tgl_presensi)=N, ...))` columns per month, silently
// mishandled short months, and was unmaintainable (LOGIC_SPEC.md §12).
// Instead this loads each employee's attendance rows and approved leave
// requests for the month once, then walks day-by-day in Go to classify
// each day. One extra query per employee for shift resolution on days with
// neither attendance nor leave (to tell "libur" apart from "alpha") -- fine
// at current scale, revisit with a batch query if the employee count grows
// large enough to matter.
package recap

import (
	"context"
	"time"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

type DayStatus string

const (
	DayStatusHadir DayStatus = "hadir"
	DayStatusIzin  DayStatus = "izin"
	DayStatusSakit DayStatus = "sakit"
	DayStatusLibur DayStatus = "libur"
	DayStatusAlpha DayStatus = "alpha"
)

type Day struct {
	Date   string    `json:"date"`
	Status DayStatus `json:"status"`
	IsLate bool      `json:"is_late,omitempty"`
}

type EmployeeRecap struct {
	EmployeeID int64          `json:"employee_id"`
	NIK        string         `json:"nik"`
	FullName   string         `json:"full_name"`
	Days       []Day          `json:"days"`
	Summary    map[string]int `json:"summary"`
}

type MonthRecap struct {
	Year        int             `json:"year"`
	Month       int             `json:"month"`
	DaysInMonth int             `json:"days_in_month"`
	Employees   []EmployeeRecap `json:"employees"`
}

type Service struct {
	employees  repo.EmployeeRepo
	attendance repo.AttendanceRepo
	leave      repo.LeaveRequestRepo
	schedules  repo.ShiftScheduleRepo
}

func NewService(employees repo.EmployeeRepo, attendance repo.AttendanceRepo, leave repo.LeaveRequestRepo, schedules repo.ShiftScheduleRepo) Service {
	return Service{employees: employees, attendance: attendance, leave: leave, schedules: schedules}
}

// Generate builds the recap grid for the given year/month. If employeeID is
// non-zero, the grid is limited to that one employee.
func (s Service) Generate(ctx context.Context, year, month int, employeeID int64) (*MonthRecap, error) {
	loc := time.UTC
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, -1)
	daysInMonth := end.Day()

	var employees []domain.Employee
	if employeeID != 0 {
		e, err := s.employees.FindByID(ctx, employeeID)
		if err != nil {
			return nil, err
		}
		if e != nil {
			employees = []domain.Employee{*e}
		}
	} else {
		all, err := s.employees.List(ctx)
		if err != nil {
			return nil, err
		}
		employees = all
	}

	result := &MonthRecap{Year: year, Month: month, DaysInMonth: daysInMonth}

	for _, emp := range employees {
		empRecap, err := s.generateForEmployee(ctx, emp, start, end, daysInMonth)
		if err != nil {
			return nil, err
		}
		result.Employees = append(result.Employees, empRecap)
	}

	return result, nil
}

func (s Service) generateForEmployee(ctx context.Context, emp domain.Employee, start, end time.Time, daysInMonth int) (EmployeeRecap, error) {
	attendances, err := s.attendance.ListByEmployeeAndDateRange(ctx, emp.ID, start, end)
	if err != nil {
		return EmployeeRecap{}, err
	}
	leaves, err := s.leave.ListApprovedOverlapping(ctx, emp.ID, start, end)
	if err != nil {
		return EmployeeRecap{}, err
	}

	type attendanceInfo struct {
		present bool
		late    bool
	}
	byDate := make(map[string]attendanceInfo, len(attendances))
	for _, a := range attendances {
		key := a.WorkDate.Format("2006-01-02")
		info := byDate[key]
		info.present = true
		if a.IsLate != nil && *a.IsLate {
			info.late = true
		}
		byDate[key] = info
	}

	leaveByDate := make(map[string]domain.LeaveType, daysInMonth)
	for _, lr := range leaves {
		for d := lr.StartDate; !d.After(lr.EndDate); d = d.AddDate(0, 0, 1) {
			leaveByDate[d.Format("2006-01-02")] = lr.Type
		}
	}

	summary := map[string]int{
		string(DayStatusHadir): 0,
		string(DayStatusIzin):  0,
		string(DayStatusSakit): 0,
		string(DayStatusLibur): 0,
		string(DayStatusAlpha): 0,
	}
	summary["telat"] = 0

	days := make([]Day, 0, daysInMonth)
	for d := 0; d < daysInMonth; d++ {
		date := start.AddDate(0, 0, d)
		key := date.Format("2006-01-02")

		day := Day{Date: key}

		switch {
		case byDate[key].present:
			day.Status = DayStatusHadir
			day.IsLate = byDate[key].late
			if day.IsLate {
				summary["telat"]++
			}
		case leaveByDate[key] == domain.LeaveTypeIzin:
			day.Status = DayStatusIzin
		case leaveByDate[key] == domain.LeaveTypeSakit:
			day.Status = DayStatusSakit
		default:
			shift, err := s.schedules.ResolveShift(ctx, emp.ID, date)
			if err != nil {
				return EmployeeRecap{}, err
			}
			if shift != nil && shift.IsDayOff {
				day.Status = DayStatusLibur
			} else {
				day.Status = DayStatusAlpha
			}
		}

		summary[string(day.Status)]++
		days = append(days, day)
	}

	return EmployeeRecap{
		EmployeeID: emp.ID,
		NIK:        emp.NIK,
		FullName:   emp.FullName,
		Days:       days,
		Summary:    summary,
	}, nil
}
