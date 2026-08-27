// Package department implements admin CRUD for departments (legacy
// `departemen` table / DepartemenController).
package department

import (
	"context"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

type Service struct {
	repo repo.DepartmentRepo
}

func NewService(r repo.DepartmentRepo) Service {
	return Service{repo: r}
}

type DepartmentInput struct {
	Code string
	Name string
}

func (s Service) List(ctx context.Context) ([]domain.Department, error) {
	return s.repo.List(ctx)
}

func (s Service) Create(ctx context.Context, in DepartmentInput) (*domain.Department, error) {
	d := &domain.Department{Code: in.Code, Name: in.Name}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}
