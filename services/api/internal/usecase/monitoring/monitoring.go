// Package monitoring implements the admin-only "view any employee's
// attendance" screen -- legacy tampilkanpeta/tampilkanpeta2
// (PresensiController), cross-employee visibility by design since it's
// admin-gated (D-7), unlike every self-service endpoint which scopes to
// the calling employee.
package monitoring

import (
	"context"
	"errors"
	"time"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/platform/storage"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

var (
	ErrNotFound    = errors.New("attendance record not found")
	ErrNoPhoto     = errors.New("no photo recorded for this event")
	ErrInvalidKind = errors.New("type must be 'in' or 'out'")
)

type Service struct {
	repo   repo.AttendanceRepo
	photos storage.LocalStore
}

func NewService(r repo.AttendanceRepo, photos storage.LocalStore) Service {
	return Service{repo: r, photos: photos}
}

func (s Service) ListByDate(ctx context.Context, date time.Time) ([]domain.Attendance, error) {
	return s.repo.ListByDate(ctx, date)
}

// Photo returns the raw image bytes for an attendance record's check-in or
// check-out photo (kind = "in" | "out").
func (s Service) Photo(ctx context.Context, id int64, kind string) ([]byte, error) {
	if kind != "in" && kind != "out" {
		return nil, ErrInvalidKind
	}

	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, ErrNotFound
	}

	path := a.CheckInPhotoPath
	if kind == "out" {
		path = a.CheckOutPhotoPath
	}
	if path == nil {
		return nil, ErrNoPhoto
	}

	return s.photos.ReadPhoto(*path)
}
