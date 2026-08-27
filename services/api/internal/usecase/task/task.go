// Package task implements employee task CRUD. Ownership is enforced on
// every read/write -- the legacy edit_task/update_task endpoints let any
// authenticated employee view or edit another employee's task by changing
// the {id} in the URL (IDOR, LOGIC_SPEC.md §7, D-5).
package task

import (
	"context"
	"errors"
	"time"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

var ErrNotFound = errors.New("task not found")

type Service struct {
	repo repo.TaskRepo
}

func NewService(r repo.TaskRepo) Service {
	return Service{repo: r}
}

type TaskInput struct {
	Title    string
	Detail   string
	StartsAt time.Time
	EndsAt   *time.Time
	Status   string
}

func (s Service) Create(ctx context.Context, employeeID int64, in TaskInput) (*domain.Task, error) {
	t := &domain.Task{
		EmployeeID: employeeID,
		Title:      in.Title,
		StartsAt:   in.StartsAt,
		EndsAt:     in.EndsAt,
		Status:     in.Status,
	}
	if in.Detail != "" {
		t.Detail = &in.Detail
	}
	if t.Status == "" {
		t.Status = "planned"
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s Service) ListOwn(ctx context.Context, employeeID int64) ([]domain.Task, error) {
	return s.repo.ListByEmployee(ctx, employeeID)
}

// Update fails with ErrNotFound if the task doesn't exist OR doesn't belong
// to employeeID -- ownership scoping, D-5.
func (s Service) Update(ctx context.Context, id, employeeID int64, in TaskInput) (*domain.Task, error) {
	t, err := s.repo.FindOwnedByID(ctx, id, employeeID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrNotFound
	}

	t.Title = in.Title
	t.StartsAt = in.StartsAt
	t.EndsAt = in.EndsAt
	if in.Status != "" {
		t.Status = in.Status
	}
	if in.Detail != "" {
		t.Detail = &in.Detail
	}

	if err := s.repo.Save(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}
