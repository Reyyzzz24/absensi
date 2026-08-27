package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eprisi/absensi-next/services/api/internal/platform/authtoken"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequireAuth_RejectsMissingToken(t *testing.T) {
	issuer := authtoken.NewIssuer("secret")
	h := RequireAuth(issuer, authtoken.AudienceAdmin)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/admin/whatever", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", rec.Code)
	}
}

func TestRequireAuth_RejectsWrongAudience(t *testing.T) {
	issuer := authtoken.NewIssuer("secret")
	// Employee-audience token presented to an admin-only route.
	token, err := issuer.Generate("1", authtoken.AudienceEmployee, "", time.Minute)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	h := RequireAuth(issuer, authtoken.AudienceAdmin)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/admin/whatever", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong audience, got %d", rec.Code)
	}
}

func TestRequireAuth_AcceptsValidToken(t *testing.T) {
	issuer := authtoken.NewIssuer("secret")
	token, err := issuer.Generate("1", authtoken.AudienceAdmin, "admin", time.Minute)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	h := RequireAuth(issuer, authtoken.AudienceAdmin)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/admin/whatever", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid token, got %d", rec.Code)
	}
}

// D-7: legacy admin panel had no per-role restriction at all -- any
// authenticated admin could hit every admin route. RequireRole is the fix;
// regressing it back to "any admin passes" would be a real security bug.

func TestRequireRole_RejectsInsufficientRole(t *testing.T) {
	issuer := authtoken.NewIssuer("secret")
	token, err := issuer.Generate("1", authtoken.AudienceAdmin, "admin", time.Minute)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	h := RequireAuth(issuer, authtoken.AudienceAdmin)(RequireRole("superadmin")(okHandler()))

	req := httptest.NewRequest(http.MethodPost, "/admin/config/shifts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for plain 'admin' hitting a superadmin-only route, got %d", rec.Code)
	}
}

func TestRequireRole_AllowsMatchingRole(t *testing.T) {
	issuer := authtoken.NewIssuer("secret")
	token, err := issuer.Generate("1", authtoken.AudienceAdmin, "superadmin", time.Minute)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	h := RequireAuth(issuer, authtoken.AudienceAdmin)(RequireRole("superadmin")(okHandler()))

	req := httptest.NewRequest(http.MethodPost, "/admin/config/shifts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for superadmin hitting a superadmin-only route, got %d", rec.Code)
	}
}
