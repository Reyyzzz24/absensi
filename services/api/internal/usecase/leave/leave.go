// Package leave implements leave/sick requests with an approval workflow.
// Merges the legacy separate izin/sakit forms into one (the "buat sakit"
// form had no store handler at all -- LOGIC_SPEC.md §6) and adds admin
// review (legacy auto-accepted with zero oversight) -- D-15/D-16.
package leave

import (
	"context"
	"errors"
	"time"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

var (
	ErrInvalidDateRange = errors.New("end_date must not be before start_date")
	ErrNotFound         = errors.New("leave request not found")
	ErrAlreadyReviewed  = errors.New("leave request already reviewed")
	ErrInvalidDecision  = errors.New("decision must be approved or rejected")
)

type Service struct {
	repo repo.LeaveRequestRepo
}

func NewService(r repo.LeaveRequestRepo) Service {
	return Service{repo: r}
}

type SubmitInput struct {
	EmployeeID int64
	Type       domain.LeaveType
	StartDate  time.Time
	EndDate    time.Time
	Reason     string
}

func (s Service) Submit(ctx context.Context, in SubmitInput) (*domain.LeaveRequest, error) {
	if in.EndDate.Before(in.StartDate) {
		return nil, ErrInvalidDateRange
	}

	lr := &domain.LeaveRequest{
		EmployeeID: in.EmployeeID,
		Type:       in.Type,
		StartDate:  in.StartDate,
		EndDate:    in.EndDate,
		Status:     domain.LeaveStatusPending, // D-15: no longer auto-accepted
	}
	if in.Reason != "" {
		lr.Reason = &in.Reason
	}

	if err := s.repo.Create(ctx, lr); err != nil {
		return nil, err
	}
	return lr, nil
}

func (s Service) ListOwn(ctx context.Context, employeeID int64) ([]domain.LeaveRequest, error) {
	return s.repo.ListByEmployee(ctx, employeeID)
}

func (s Service) ListForAdmin(ctx context.Context, status domain.LeaveStatus) ([]domain.LeaveRequest, error) {
	return s.repo.ListByStatus(ctx, status)
}

// Review lets an admin approve or reject a pending request. Approved leave
// is excluded from "absent" in attendance recap reports -- enforced in the
// reports query logic (not yet implemented), not here.
func (s Service) Review(ctx context.Context, id int64, reviewerUserID int64, decision string) (*domain.LeaveRequest, error) {
	var newStatus domain.LeaveStatus
	switch decision {
	case "approved":
		newStatus = domain.LeaveStatusApproved
	case "rejected":
		newStatus = domain.LeaveStatusRejected
	default:
		return nil, ErrInvalidDecision
	}

	lr, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if lr == nil {
		return nil, ErrNotFound
	}
	if lr.Status != domain.LeaveStatusPending {
		return nil, ErrAlreadyReviewed
	}

	now := time.Now()
	lr.Status = newStatus
	lr.ReviewedBy = &reviewerUserID
	lr.ReviewedAt = &now

	if err := s.repo.Save(ctx, lr); err != nil {
		return nil, err
	}
	return lr, nil
}
