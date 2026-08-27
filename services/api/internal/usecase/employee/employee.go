// Package employee implements admin CRUD for employee accounts (legacy
// `karyawan` table / KaryawanController).
//
// Deliberately does NOT replicate the legacy bug where editing any field on
// an employee record silently reset their password to a hardcoded "12345"
// (KaryawanController::update, LOGIC_SPEC.md §9 / D-3). Update only rehashes
// the password when the admin explicitly supplies a new one; every other
// field edit leaves the existing hash untouched.
package employee

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

var ErrNIKTaken = errors.New("nik already registered")
var ErrNotFound = errors.New("employee not found")

type Service struct {
	repo repo.EmployeeRepo
}

func NewService(r repo.EmployeeRepo) Service {
	return Service{repo: r}
}

type CreateInput struct {
	NIK          string
	FullName     string
	Password     string
	DepartmentID *int64
	Position     string
	Phone        string
}

func (s Service) List(ctx context.Context, q string) ([]domain.Employee, error) {
	return s.repo.List(ctx, q)
}

func (s Service) Create(ctx context.Context, in CreateInput) (*domain.Employee, error) {
	existing, err := s.repo.FindByNIK(ctx, in.NIK)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrNIKTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	e := &domain.Employee{
		NIK:          in.NIK,
		FullName:     in.FullName,
		PasswordHash: string(hash),
		DepartmentID: in.DepartmentID,
		IsActive:     true,
	}
	if in.Position != "" {
		e.Position = &in.Position
	}
	if in.Phone != "" {
		e.Phone = &in.Phone
	}

	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

type UpdateInput struct {
	FullName     string
	Password     string // optional; only rehashed when non-empty
	DepartmentID *int64
	Position     string
	Phone        string
	IsActive     bool
}

func (s Service) Update(ctx context.Context, id int64, in UpdateInput) (*domain.Employee, error) {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, ErrNotFound
	}

	e.FullName = in.FullName
	e.DepartmentID = in.DepartmentID
	e.IsActive = in.IsActive
	if in.Position != "" {
		e.Position = &in.Position
	} else {
		e.Position = nil
	}
	if in.Phone != "" {
		e.Phone = &in.Phone
	} else {
		e.Phone = nil
	}

	if in.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		e.PasswordHash = string(hash)
	}

	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}
