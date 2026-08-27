package task_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
	"github.com/eprisi/absensi-next/services/api/internal/testutil"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/task"
)

// A5/D-5: legacy edit_task/update_task let any authenticated employee edit
// another employee's task by changing the {id} in the URL. Update must
// return ErrNotFound (not the other employee's task, and not a different
// "forbidden" error that would leak the task's existence) when the caller
// doesn't own it.

func TestUpdate_RejectsAnotherEmployeesTask(t *testing.T) {
	gdb := testutil.NewDB(t)
	svc := task.NewService(repo.NewTaskRepo(gdb))
	ctx := context.Background()

	owner := domain.Employee{NIK: "0001", FullName: "Owner", PasswordHash: "x", IsActive: true}
	attacker := domain.Employee{NIK: "0002", FullName: "Attacker", PasswordHash: "x", IsActive: true}
	if err := gdb.Create(&owner).Error; err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if err := gdb.Create(&attacker).Error; err != nil {
		t.Fatalf("create attacker: %v", err)
	}

	created, err := svc.Create(ctx, owner.ID, task.TaskInput{Title: "Owner's task", StartsAt: time.Now()})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	_, err = svc.Update(ctx, created.ID, attacker.ID, task.TaskInput{Title: "hijacked", StartsAt: time.Now()})
	if !errors.Is(err, task.ErrNotFound) {
		t.Fatalf("expected ErrNotFound when a different employee tries to update the task, got %v", err)
	}

	// Confirm the task was genuinely untouched, not just that the call errored.
	var reloaded domain.Task
	if err := gdb.First(&reloaded, created.ID).Error; err != nil {
		t.Fatalf("reload task: %v", err)
	}
	if reloaded.Title != "Owner's task" {
		t.Fatalf("task must not be modified by a non-owner, got title %q", reloaded.Title)
	}
}

func TestUpdate_AllowsOwner(t *testing.T) {
	gdb := testutil.NewDB(t)
	svc := task.NewService(repo.NewTaskRepo(gdb))
	ctx := context.Background()

	owner := domain.Employee{NIK: "0001", FullName: "Owner", PasswordHash: "x", IsActive: true}
	if err := gdb.Create(&owner).Error; err != nil {
		t.Fatalf("create owner: %v", err)
	}

	created, err := svc.Create(ctx, owner.ID, task.TaskInput{Title: "Original", StartsAt: time.Now()})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	updated, err := svc.Update(ctx, created.ID, owner.ID, task.TaskInput{Title: "Updated", StartsAt: time.Now()})
	if err != nil {
		t.Fatalf("owner should be able to update their own task, got %v", err)
	}
	if updated.Title != "Updated" {
		t.Fatalf("expected title to be updated, got %q", updated.Title)
	}
}

func TestListOwn_DoesNotLeakOtherEmployeesTasks(t *testing.T) {
	gdb := testutil.NewDB(t)
	svc := task.NewService(repo.NewTaskRepo(gdb))
	ctx := context.Background()

	a := domain.Employee{NIK: "0001", FullName: "A", PasswordHash: "x", IsActive: true}
	b := domain.Employee{NIK: "0002", FullName: "B", PasswordHash: "x", IsActive: true}
	if err := gdb.Create(&a).Error; err != nil {
		t.Fatalf("create a: %v", err)
	}
	if err := gdb.Create(&b).Error; err != nil {
		t.Fatalf("create b: %v", err)
	}

	if _, err := svc.Create(ctx, a.ID, task.TaskInput{Title: "A's task", StartsAt: time.Now()}); err != nil {
		t.Fatalf("create task for a: %v", err)
	}
	if _, err := svc.Create(ctx, b.ID, task.TaskInput{Title: "B's task", StartsAt: time.Now()}); err != nil {
		t.Fatalf("create task for b: %v", err)
	}

	tasksForA, err := svc.ListOwn(ctx, a.ID)
	if err != nil {
		t.Fatalf("list own for a: %v", err)
	}
	if len(tasksForA) != 1 || tasksForA[0].Title != "A's task" {
		t.Fatalf("expected exactly A's own task, got %+v", tasksForA)
	}
}
