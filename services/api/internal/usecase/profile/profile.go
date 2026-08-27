// Package profile implements the self-service "who am I" endpoints shared
// by both audiences (GET/PATCH /api/me, avatar upload, change password).
// Every method is scoped by the (audience, id) pair taken from the caller's
// own JWT claims -- there is deliberately no id parameter accepted from the
// request body/query for any of these, so there is no IDOR surface here
// (D-5 pattern): a user can only ever act on their own record.
package profile

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/platform/authtoken"
	"github.com/eprisi/absensi-next/services/api/internal/platform/storage"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

var ErrNotFound = errors.New("account not found")
var ErrWrongPassword = errors.New("current password is incorrect")

type Profile struct {
	ID         int64   `json:"id"`
	Audience   string  `json:"audience"` // "employee" | "admin"
	Name       string  `json:"name"`
	Identifier string  `json:"identifier"` // NIK for employee, username for admin
	Email      *string `json:"email,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Department *string `json:"department,omitempty"`
	Position   *string `json:"position,omitempty"`
	Role       *string `json:"role,omitempty"`
	PhotoPath  *string `json:"photo_path,omitempty"`
}

type Service struct {
	employees repo.EmployeeRepo
	users     repo.UserRepo
	avatars   storage.LocalStore
}

func NewService(employees repo.EmployeeRepo, users repo.UserRepo, avatars storage.LocalStore) Service {
	return Service{employees: employees, users: users, avatars: avatars}
}

func (s Service) Get(ctx context.Context, aud authtoken.Audience, id int64) (*Profile, error) {
	if aud == authtoken.AudienceEmployee {
		e, err := s.employees.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if e == nil {
			return nil, ErrNotFound
		}
		return employeeToProfile(e), nil
	}

	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return userToProfile(u), nil
}

// UpdatePhone is the only field either audience can PATCH about themselves
// directly (approved scope: foto, no. HP, password -- nama/email/NIP/
// departemen/jabatan are read-only self-service, admin-managed elsewhere).
func (s Service) UpdatePhone(ctx context.Context, aud authtoken.Audience, id int64, phone string) (*Profile, error) {
	var phonePtr *string
	if phone != "" {
		phonePtr = &phone
	}

	if aud == authtoken.AudienceEmployee {
		e, err := s.employees.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if e == nil {
			return nil, ErrNotFound
		}
		e.Phone = phonePtr
		if err := s.employees.Update(ctx, e); err != nil {
			return nil, err
		}
		return employeeToProfile(e), nil
	}

	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	u.Phone = phonePtr
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return userToProfile(u), nil
}

func (s Service) UpdateAvatar(ctx context.Context, aud authtoken.Audience, id int64, photoBytes []byte) (*Profile, error) {
	relPath, err := s.avatars.SaveAvatar(string(aud), id, photoBytes)
	if err != nil {
		return nil, err
	}

	if aud == authtoken.AudienceEmployee {
		e, err := s.employees.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if e == nil {
			return nil, ErrNotFound
		}
		e.PhotoPath = &relPath
		if err := s.employees.Update(ctx, e); err != nil {
			return nil, err
		}
		return employeeToProfile(e), nil
	}

	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	u.PhotoPath = &relPath
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return userToProfile(u), nil
}

// ReadAvatar reads back a previously-saved avatar file by its relative
// path (as returned by Get/AvatarPath) -- thin passthrough kept here so the
// handler never imports storage.LocalStore directly.
func (s Service) ReadAvatar(relPath string) ([]byte, error) {
	return s.avatars.ReadPhoto(relPath)
}

// AvatarPath returns the currently stored relative photo path, used by the
// handler that streams the image back (mirrors monitoring_handler's photo
// pattern). Returns "" (not an error) when no avatar has been set yet.
func (s Service) AvatarPath(ctx context.Context, aud authtoken.Audience, id int64) (string, error) {
	p, err := s.Get(ctx, aud, id)
	if err != nil {
		return "", err
	}
	if p.PhotoPath == nil {
		return "", nil
	}
	return *p.PhotoPath, nil
}

// ChangePassword requires the current password to match before rehashing --
// never allow silently swapping a password without proving you hold the old
// one (this is the exact legacy bug class D-3 already fixed for admin-edits-
// employee; here it's the same principle applied to self-service).
func (s Service) ChangePassword(ctx context.Context, aud authtoken.Audience, id int64, currentPassword, newPassword string) error {
	if aud == authtoken.AudienceEmployee {
		e, err := s.employees.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if e == nil {
			return ErrNotFound
		}
		if bcrypt.CompareHashAndPassword([]byte(e.PasswordHash), []byte(currentPassword)) != nil {
			return ErrWrongPassword
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		e.PasswordHash = string(hash)
		return s.employees.Update(ctx, e)
	}

	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrNotFound
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(currentPassword)) != nil {
		return ErrWrongPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return s.users.Update(ctx, u)
}

func employeeToProfile(e *domain.Employee) *Profile {
	p := &Profile{
		ID:         e.ID,
		Audience:   "employee",
		Name:       e.FullName,
		Identifier: e.NIK,
		Phone:      e.Phone,
		Position:   e.Position,
		PhotoPath:  e.PhotoPath,
	}
	if e.Department != nil {
		p.Department = &e.Department.Name
	}
	return p
}

func userToProfile(u *domain.User) *Profile {
	role := u.Role.Name
	return &Profile{
		ID:         u.ID,
		Audience:   "admin",
		Name:       u.Name,
		Identifier: u.Username,
		Email:      u.Email,
		Phone:      u.Phone,
		Role:       &role,
		PhotoPath:  u.PhotoPath,
	}
}
