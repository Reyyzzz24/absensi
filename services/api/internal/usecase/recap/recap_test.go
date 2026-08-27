package recap_test

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
	"github.com/eprisi/absensi-next/services/api/internal/testutil"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/holiday"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/recap"
)

func newService(t *testing.T) (recap.Service, *gorm.DB) {
	t.Helper()
	gdb := testutil.NewDB(t)
	holidays := holiday.NewService(repo.NewCompanySettingsRepo(gdb), repo.NewNationalHolidayRepo(gdb), repo.NewCompanyHolidayRepo(gdb), holiday.NewGoogleCalendarFetcher())
	svc := recap.NewService(
		repo.NewEmployeeRepo(gdb),
		repo.NewAttendanceRepo(gdb),
		repo.NewLeaveRequestRepo(gdb),
		repo.NewShiftScheduleRepo(gdb),
		holidays,
	)
	return svc, gdb
}

func mustCreateEmployee(t *testing.T, gdb *gorm.DB, nik string) domain.Employee {
	t.Helper()
	e := domain.Employee{NIK: nik, FullName: "Test " + nik, PasswordHash: "x", IsActive: true}
	if err := gdb.Create(&e).Error; err != nil {
		t.Fatalf("create employee: %v", err)
	}
	return e
}

// Parity matrix #14: legacy izin/sakit had no effect on the attendance
// recap at all (present/absent columns ignored approved leave). The new
// recap must classify an approved-leave day as "izin"/"sakit", never
// "alpha" (absent) -- D-15/D-16.
func TestGenerate_ApprovedLeaveClassifiedAsIzinNotAlpha(t *testing.T) {
	svc, gdb := newService(t)
	ctx := context.Background()
	emp := mustCreateEmployee(t, gdb, "0001")

	year, month := 2026, 8
	leaveDate := time.Date(year, time.Month(month), 15, 0, 0, 0, 0, time.UTC)

	admin := mustCreateAdminUserForRecap(t, gdb)
	lr := domain.LeaveRequest{
		EmployeeID: emp.ID,
		Type:       domain.LeaveTypeIzin,
		StartDate:  leaveDate,
		EndDate:    leaveDate,
		Status:     domain.LeaveStatusApproved,
		ReviewedBy: &admin.ID,
	}
	if err := gdb.Create(&lr).Error; err != nil {
		t.Fatalf("create leave request: %v", err)
	}

	result, err := svc.Generate(ctx, year, month, emp.ID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(result.Employees) != 1 {
		t.Fatalf("expected 1 employee in recap, got %d", len(result.Employees))
	}
	day := result.Employees[0].Days[14] // day 15 = index 14
	if day.Status != recap.DayStatusIzin {
		t.Fatalf("expected day 15 to be classified as izin, got %s", day.Status)
	}
	if day.Status == recap.DayStatusAlpha {
		t.Fatal("approved leave must never be classified as alpha (absent)")
	}
}

// A "pending" (not yet approved) leave request must NOT suppress "alpha" --
// only approved leave excuses an absence, matching D-15's whole point
// (unlike legacy, which had no approval gate to begin with).
func TestGenerate_PendingLeaveDoesNotSuppressAlpha(t *testing.T) {
	svc, gdb := newService(t)
	ctx := context.Background()
	emp := mustCreateEmployee(t, gdb, "0001")

	year, month := 2026, 8
	// The 15th is a Saturday in Aug 2026 -- would now correctly resolve as
	// "libur" via the D-25 holiday resolver regardless of the leave
	// request, which defeats this test's actual point (a still-PENDING
	// leave must not suppress alpha). Use the 17th (a Monday, an ordinary
	// workday under the default Mon-Fri config) so the two concerns don't
	// collide.
	leaveDate := time.Date(year, time.Month(month), 17, 0, 0, 0, 0, time.UTC)
	lr := domain.LeaveRequest{
		EmployeeID: emp.ID,
		Type:       domain.LeaveTypeIzin,
		StartDate:  leaveDate,
		EndDate:    leaveDate,
		Status:     domain.LeaveStatusPending,
	}
	if err := gdb.Create(&lr).Error; err != nil {
		t.Fatalf("create leave request: %v", err)
	}

	result, err := svc.Generate(ctx, year, month, emp.ID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	day := result.Employees[0].Days[16] // day 17 = index 16
	if day.Status != recap.DayStatusAlpha {
		t.Fatalf("expected a still-pending leave day to remain alpha, got %s", day.Status)
	}
}

func mustCreateAdminUserForRecap(t *testing.T, gdb *gorm.DB) domain.User {
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
