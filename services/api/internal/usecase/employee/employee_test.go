package employee_test

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/eprisi/absensi-next/services/api/internal/repo"
	"github.com/eprisi/absensi-next/services/api/internal/testutil"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/employee"
)

// LOGIC_SPEC.md §9 / D-3: legacy KaryawanController::update silently reset
// an employee's password to a hardcoded "12345" on ANY field edit. Update
// must leave the existing password hash untouched when no new password is
// supplied.
func TestUpdate_DoesNotResetPasswordWhenOmitted(t *testing.T) {
	gdb := testutil.NewDB(t)
	svc := employee.NewService(repo.NewEmployeeRepo(gdb))
	ctx := context.Background()

	created, err := svc.Create(ctx, employee.CreateInput{
		NIK: "0001", FullName: "Original Name", Password: "original-password",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	originalHash := created.PasswordHash

	updated, err := svc.Update(ctx, created.ID, employee.UpdateInput{
		FullName: "Edited Name", // only the name changes; Password left empty
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if updated.PasswordHash != originalHash {
		t.Fatal("editing a field other than password must not change the password hash (D-3 regression)")
	}
	// The original password must still work.
	if err := bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("original-password")); err != nil {
		t.Fatalf("original password should still verify after an unrelated edit: %v", err)
	}
}

func TestUpdate_RehashesWhenNewPasswordExplicitlyProvided(t *testing.T) {
	gdb := testutil.NewDB(t)
	svc := employee.NewService(repo.NewEmployeeRepo(gdb))
	ctx := context.Background()

	created, err := svc.Create(ctx, employee.CreateInput{
		NIK: "0001", FullName: "Name", Password: "old-password",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := svc.Update(ctx, created.ID, employee.UpdateInput{
		FullName: "Name", Password: "new-password", IsActive: true,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("new-password")); err != nil {
		t.Fatal("explicitly-provided new password should verify after update")
	}
	if bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("old-password")) == nil {
		t.Fatal("old password should no longer verify after an explicit password change")
	}
}

// Deactivation must be a real, enforceable state change -- reflected
// immediately in what Update returns, and (covered separately in the auth
// package) actually blocks login.
func TestUpdate_CanDeactivateAndReactivate(t *testing.T) {
	gdb := testutil.NewDB(t)
	svc := employee.NewService(repo.NewEmployeeRepo(gdb))
	ctx := context.Background()

	created, err := svc.Create(ctx, employee.CreateInput{NIK: "0001", FullName: "Name", Password: "password123"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !created.IsActive {
		t.Fatal("newly created employee should be active by default")
	}

	deactivated, err := svc.Update(ctx, created.ID, employee.UpdateInput{FullName: "Name", IsActive: false})
	if err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if deactivated.IsActive {
		t.Fatal("expected employee to be inactive after update")
	}

	reactivated, err := svc.Update(ctx, created.ID, employee.UpdateInput{FullName: "Name", IsActive: true})
	if err != nil {
		t.Fatalf("reactivate: %v", err)
	}
	if !reactivated.IsActive {
		t.Fatal("expected employee to be active again after re-enabling")
	}
}

func TestUpdate_UnknownEmployeeReturnsNotFound(t *testing.T) {
	gdb := testutil.NewDB(t)
	svc := employee.NewService(repo.NewEmployeeRepo(gdb))

	_, err := svc.Update(context.Background(), 99999, employee.UpdateInput{FullName: "X", IsActive: true})
	if !errors.Is(err, employee.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for a nonexistent employee id, got %v", err)
	}
}

func TestCreate_RejectsDuplicateNIK(t *testing.T) {
	gdb := testutil.NewDB(t)
	svc := employee.NewService(repo.NewEmployeeRepo(gdb))
	ctx := context.Background()

	if _, err := svc.Create(ctx, employee.CreateInput{NIK: "0001", FullName: "A", Password: "password123"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := svc.Create(ctx, employee.CreateInput{NIK: "0001", FullName: "B", Password: "password123"})
	if !errors.Is(err, employee.ErrNIKTaken) {
		t.Fatalf("expected ErrNIKTaken for duplicate NIK, got %v", err)
	}
}
