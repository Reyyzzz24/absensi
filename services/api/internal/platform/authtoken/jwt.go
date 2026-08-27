// Package authtoken issues and verifies the JWTs used by both the employee
// (karyawan) and admin (user) guards -- replacing Laravel's session-based
// auth:karyawan / auth:user guards (config/auth.php) with stateless tokens.
package authtoken

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Audience string

const (
	AudienceEmployee Audience = "employee"
	AudienceAdmin    Audience = "admin"
)

type Claims struct {
	jwt.RegisteredClaims
	Audience Audience `json:"aud_type"`
	Role     string   `json:"role,omitempty"` // populated for admin tokens, RBAC per D-7
}

type Issuer struct {
	secret []byte
}

func NewIssuer(secret string) Issuer {
	return Issuer{secret: []byte(secret)}
}

func (i Issuer) Generate(subject string, aud Audience, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	jti, err := newJTI()
	if err != nil {
		return "", err
	}
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		Audience: aud,
		Role:     role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(i.secret)
}

// newJTI generates the random ID embedded as the JWT's "jti" claim -- used
// to identify a specific refresh token for revocation on logout (D-24),
// since the token itself is never persisted server-side.
func newJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (i Issuer) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
