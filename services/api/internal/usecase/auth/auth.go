// Package auth implements employee and admin login. Password hashing stays
// bcrypt (golang.org/x/crypto/bcrypt), matching Laravel's default -- legacy
// bcrypt hashes verify correctly here with no rehash/migration needed
// (docs/DECISIONS.md D-12 / CLAUDE.md §8).
package auth

import (
	"context"
	"errors"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/platform/authtoken"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type Service struct {
	employees     repo.EmployeeRepo
	users         repo.UserRepo
	issuer        authtoken.Issuer
	revokedTokens repo.RevokedTokenRepo
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewService(employees repo.EmployeeRepo, users repo.UserRepo, issuer authtoken.Issuer, revokedTokens repo.RevokedTokenRepo, accessTTL, refreshTTL time.Duration) Service {
	return Service{employees: employees, users: users, issuer: issuer, revokedTokens: revokedTokens, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// LoginEmployee authenticates against the employees table (legacy `karyawan`
// / guard `karyawan`). Ownership/self-service endpoints derive the employee
// ID from this token's Subject, never from client-supplied input.
func (s Service) LoginEmployee(ctx context.Context, nik, password string) (Tokens, *domain.Employee, error) {
	emp, err := s.employees.FindByNIK(ctx, nik)
	if err != nil {
		return Tokens{}, nil, err
	}
	if emp == nil {
		return Tokens{}, nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(emp.PasswordHash), []byte(password)); err != nil {
		return Tokens{}, nil, ErrInvalidCredentials
	}

	subject := strconv.FormatInt(emp.ID, 10)
	access, err := s.issuer.Generate(subject, authtoken.AudienceEmployee, "", s.accessTTL)
	if err != nil {
		return Tokens{}, nil, err
	}
	refresh, err := s.issuer.Generate(subject, authtoken.AudienceEmployee, "", s.refreshTTL)
	if err != nil {
		return Tokens{}, nil, err
	}

	return Tokens{AccessToken: access, RefreshToken: refresh, ExpiresIn: int(s.accessTTL.Seconds())}, emp, nil
}

// LoginAdmin authenticates against the users table (legacy `users` / guard
// `user`). Role is embedded in the token for RBAC enforcement -- D-7 --
// unlike the legacy system where any authenticated admin could hit every
// admin route regardless of role_id.
func (s Service) LoginAdmin(ctx context.Context, username, password string) (Tokens, *domain.User, error) {
	u, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		return Tokens{}, nil, err
	}
	if u == nil {
		return Tokens{}, nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return Tokens{}, nil, ErrInvalidCredentials
	}

	subject := strconv.FormatInt(u.ID, 10)
	access, err := s.issuer.Generate(subject, authtoken.AudienceAdmin, u.Role.Name, s.accessTTL)
	if err != nil {
		return Tokens{}, nil, err
	}
	refresh, err := s.issuer.Generate(subject, authtoken.AudienceAdmin, u.Role.Name, s.refreshTTL)
	if err != nil {
		return Tokens{}, nil, err
	}

	return Tokens{AccessToken: access, RefreshToken: refresh, ExpiresIn: int(s.accessTTL.Seconds())}, u, nil
}

// Refresh mints a new access token from a still-valid, non-revoked refresh
// token. Revocation (D-24) closes the earlier Phase 2 gap where /auth/logout
// was a no-op against stateless JWTs -- a denylist of revoked jti's is
// checked here instead of maintaining a full session store.
func (s Service) Refresh(ctx context.Context, refreshToken string) (Tokens, error) {
	claims, err := s.issuer.Parse(refreshToken)
	if err != nil {
		return Tokens{}, ErrInvalidCredentials
	}
	revoked, err := s.revokedTokens.IsRevoked(ctx, claims.ID)
	if err != nil {
		return Tokens{}, err
	}
	if revoked {
		return Tokens{}, ErrInvalidCredentials
	}
	access, err := s.issuer.Generate(claims.Subject, claims.Audience, claims.Role, s.accessTTL)
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{AccessToken: access, RefreshToken: refreshToken, ExpiresIn: int(s.accessTTL.Seconds())}, nil
}

// Logout revokes the given refresh token so it can no longer be used with
// Refresh, even though it remains cryptographically valid until it expires
// naturally. Idempotent and tolerant of already-expired/garbage tokens --
// logout must never fail the client just because their session was already
// gone.
func (s Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.issuer.Parse(refreshToken)
	if err != nil {
		return nil
	}
	return s.revokedTokens.Revoke(ctx, claims.ID, claims.ExpiresAt.Time)
}
