package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	appmiddleware "github.com/eprisi/absensi-next/services/api/internal/middleware"
	"github.com/eprisi/absensi-next/services/api/internal/platform/authtoken"
	"github.com/eprisi/absensi-next/services/api/internal/platform/storage"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/profile"
)

func authAudience(s string) authtoken.Audience { return authtoken.Audience(s) }

const maxAvatarBytes = 2 * 1024 * 1024 // 2MB decoded

type ProfileHandler struct {
	service profile.Service
}

func NewProfileHandler(service profile.Service) ProfileHandler {
	return ProfileHandler{service: service}
}

// identity reads the caller's own (audience, id) from the JWT already
// verified by RequireAuth -- every profile.Service call below is scoped to
// exactly this pair, never anything from the request body/query (D-5).
func identity(r *http.Request) (aud string, id int64, ok bool) {
	claims, has := appmiddleware.ClaimsFromContext(r.Context())
	if !has {
		return "", 0, false
	}
	parsedID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return "", 0, false
	}
	return string(claims.Audience), parsedID, true
}

func (h ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	aud, id, ok := identity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	p, err := h.service.Get(r.Context(), authAudience(aud), id)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

type updateProfileRequest struct {
	Phone string `json:"phone"`
}

func (h ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	aud, id, ok := identity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p, err := h.service.UpdatePhone(r.Context(), authAudience(aud), id, req.Phone)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

type uploadAvatarRequest struct {
	Photo string `json:"photo"` // base64, same convention as check-in photos
}

func (h ProfileHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	aud, id, ok := identity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req uploadAvatarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	photoBytes, err := base64.StdEncoding.DecodeString(req.Photo)
	if err != nil || len(photoBytes) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "photo must be valid base64 image data")
		return
	}
	if len(photoBytes) > maxAvatarBytes {
		writeError(w, http.StatusUnprocessableEntity, "photo exceeds 2MB limit")
		return
	}
	contentType := http.DetectContentType(photoBytes)
	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/webp" {
		writeError(w, http.StatusUnprocessableEntity, "photo must be PNG, JPEG, or WebP")
		return
	}

	p, err := h.service.UpdateAvatar(r.Context(), authAudience(aud), id, photoBytes)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// Avatar streams the caller's own avatar image -- same "proxy through a
// same-origin route because <img> can't send an Authorization header"
// pattern as monitoring_handler.Photo, mirrored on the frontend.
func (h ProfileHandler) Avatar(w http.ResponseWriter, r *http.Request) {
	aud, id, ok := identity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	relPath, err := h.service.AvatarPath(r.Context(), authAudience(aud), id)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	if relPath == "" {
		writeError(w, http.StatusNotFound, "no avatar set")
		return
	}
	bytes, err := h.service.ReadAvatar(relPath)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidPath) {
			writeError(w, http.StatusInternalServerError, "invalid photo path")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load avatar")
		return
	}
	w.Header().Set("Content-Type", http.DetectContentType(bytes))
	w.Write(bytes)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	aud, id, ok := identity(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CurrentPassword == "" || len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "current_password is required and new_password must be at least 8 characters")
		return
	}
	err := h.service.ChangePassword(r.Context(), authAudience(aud), id, req.CurrentPassword, req.NewPassword)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, profile.ErrNotFound):
		writeError(w, http.StatusNotFound, "account not found")
	case errors.Is(err, profile.ErrWrongPassword):
		writeError(w, http.StatusUnprocessableEntity, "current password is incorrect")
	default:
		writeError(w, http.StatusInternalServerError, "request failed")
	}
}
