package leave_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
	"github.com/eprisi/absensi-next/services/api/internal/testutil"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/leave"
)

// reviewerID is a valid users.id (FK target for leave_requests.reviewed_by)
// shared by every test in this file via newService.
func newService(t *testing.T) (svc leave.Service, employeeID, reviewerID int64) {
	t.Helper()
	gdb := testutil.NewDB(t)
	emp := mustCreateEmployee(t, gdb)
	admin := mustCreateAdminUser(t, gdb)
	return leave.NewService(repo.NewLeaveRequestRepo(gdb)), emp.ID, admin.ID
}

func mustCreateEmployee(t *testing.T, gdb *gorm.DB) domain.Employee {
	t.Helper()
	e := domain.Employee{NIK: "0001", FullName: "Test", PasswordHash: "x", IsActive: true}
	if err := gdb.Create(&e).Error; err != nil {
		t.Fatalf("create employee: %v", err)
	}
	return e
}

// mustCreateAdminUser relies on the "superadmin"/"admin" roles that
// migration 000001 seeds into every fresh database.
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

// D-15: legacy izin/sakit requests were inserted directly with no approval
// step at all. The new flow must default to "pending", never auto-approve.
func TestSubmit_DefaultsToPending(t *testing.T) {
	svc, employeeID, _ := newService(t)
	lr, err := svc.Submit(context.Background(), leave.SubmitInput{
		EmployeeID: employeeID,
		Type:       domain.LeaveTypeIzin,
		StartDate:  time.Now(),
		EndDate:    time.Now().AddDate(0, 0, 1),
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if lr.Status != domain.LeaveStatusPending {
		t.Fatalf("expected pending status, got %s", lr.Status)
	}
}

func TestSubmit_RejectsEndDateBeforeStartDate(t *testing.T) {
	svc, employeeID, _ := newService(t)
	_, err := svc.Submit(context.Background(), leave.SubmitInput{
		EmployeeID: employeeID,
		Type:       domain.LeaveTypeSakit,
		StartDate:  time.Now(),
		EndDate:    time.Now().AddDate(0, 0, -1),
	})
	if !errors.Is(err, leave.ErrInvalidDateRange) {
		t.Fatalf("expected ErrInvalidDateRange, got %v", err)
	}
}

// D-15: an admin decision actually changes state and is recorded.
func TestReview_ApprovesPendingRequest(t *testing.T) {
	svc, employeeID, reviewerID := newService(t)
	ctx := context.Background()
	lr, err := svc.Submit(ctx, leave.SubmitInput{
		EmployeeID: employeeID, Type: domain.LeaveTypeIzin,
		StartDate: time.Now(), EndDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	reviewed, err := svc.Review(ctx, lr.ID, reviewerID, "approved")
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	if reviewed.Status != domain.LeaveStatusApproved {
		t.Fatalf("expected approved, got %s", reviewed.Status)
	}
	if reviewed.ReviewedBy == nil || *reviewed.ReviewedBy != reviewerID {
		t.Fatalf("expected reviewed_by=%d, got %v", reviewerID, reviewed.ReviewedBy)
	}
}

func TestReview_RejectsDoubleReview(t *testing.T) {
	svc, employeeID, reviewerID := newService(t)
	ctx := context.Background()
	lr, err := svc.Submit(ctx, leave.SubmitInput{
		EmployeeID: employeeID, Type: domain.LeaveTypeIzin,
		StartDate: time.Now(), EndDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := svc.Review(ctx, lr.ID, reviewerID, "approved"); err != nil {
		t.Fatalf("first review: %v", err)
	}

	_, err = svc.Review(ctx, lr.ID, reviewerID, "rejected")
	if !errors.Is(err, leave.ErrAlreadyReviewed) {
		t.Fatalf("expected ErrAlreadyReviewed on a second review, got %v", err)
	}
}

func TestReview_RejectsInvalidDecision(t *testing.T) {
	svc, employeeID, reviewerID := newService(t)
	ctx := context.Background()
	lr, err := svc.Submit(ctx, leave.SubmitInput{
		EmployeeID: employeeID, Type: domain.LeaveTypeIzin,
		StartDate: time.Now(), EndDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}

	_, err = svc.Review(ctx, lr.ID, reviewerID, "maybe")
	if !errors.Is(err, leave.ErrInvalidDecision) {
		t.Fatalf("expected ErrInvalidDecision, got %v", err)
	}
}
