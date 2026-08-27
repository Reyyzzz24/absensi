package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eprisi/absensi-next/services/api/internal/usecase/auth"
)

type AuthHandler struct {
	service auth.Service
}

func NewAuthHandler(service auth.Service) AuthHandler {
	return AuthHandler{service: service}
}

type loginRequest struct {
	NIK      string `json:"nik"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (h AuthHandler) EmployeeLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NIK == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "nik and password are required")
		return
	}

	tokens, _, err := h.service.LoginEmployee(r.Context(), req.NIK, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	writeJSON(w, http.StatusOK, tokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	})
}

func (h AuthHandler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	tokens, _, err := h.service.LoginAdmin(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	writeJSON(w, http.StatusOK, tokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	tokens, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	writeJSON(w, http.StatusOK, tokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	})
}

// Logout revokes the given refresh token (D-24) so it can't be used with
// /auth/refresh again. Body is optional -- a client that already lost its
// refresh token (or never had one) still gets a clean 204, since the goal
// is "end this session," which is trivially true if there's nothing to end.
func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.RefreshToken != "" {
		if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
			writeError(w, http.StatusInternalServerError, "logout failed")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
