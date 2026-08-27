// Package storage saves check-in/out photos to local disk. Filenames are
// unique per event (unix nanoseconds), fixing the legacy collision bug where
// a same-day repeat check-in silently overwrote the previous photo -- D-11.
//
// Local disk is a dev/Phase-2 placeholder; object storage (S3/MinIO) is a
// Phase 5 production concern, not decided yet.
package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrInvalidPath = errors.New("invalid photo path")

type LocalStore struct {
	root string
}

func NewLocalStore(root string) LocalStore {
	return LocalStore{root: root}
}

// SaveCheckInPhoto writes photoBytes under {root}/checkin/{employeeID}/ and
// returns the path relative to root (what gets persisted in
// attendances.check_in_photo_path / check_out_photo_path).
func (s LocalStore) SaveCheckInPhoto(employeeID int64, photoBytes []byte) (string, error) {
	dir := filepath.Join(s.root, "checkin", fmt.Sprintf("%d", employeeID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create storage dir: %w", err)
	}

	filename := fmt.Sprintf("%d.png", time.Now().UnixNano())
	fullPath := filepath.Join(dir, filename)

	if err := os.WriteFile(fullPath, photoBytes, 0o644); err != nil {
		return "", fmt.Errorf("write photo: %w", err)
	}

	relPath, err := filepath.Rel(s.root, fullPath)
	if err != nil {
		return "", err
	}
	return relPath, nil
}

// SaveAvatar writes photoBytes under {root}/avatar/{audience}/{id}/ and
// returns the path relative to root. Old avatar files for the same
// audience/id are not cleaned up (same unique-filename-per-event pattern as
// SaveCheckInPhoto, D-11) -- acceptable in this dev-local-disk phase, a
// production object-storage migration (Phase 5) would add lifecycle rules.
func (s LocalStore) SaveAvatar(audience string, id int64, photoBytes []byte) (string, error) {
	dir := filepath.Join(s.root, "avatar", audience, fmt.Sprintf("%d", id))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create storage dir: %w", err)
	}

	filename := fmt.Sprintf("%d.png", time.Now().UnixNano())
	fullPath := filepath.Join(dir, filename)

	if err := os.WriteFile(fullPath, photoBytes, 0o644); err != nil {
		return "", fmt.Errorf("write avatar: %w", err)
	}

	relPath, err := filepath.Rel(s.root, fullPath)
	if err != nil {
		return "", err
	}
	return relPath, nil
}

// ReadPhoto reads a previously-saved photo back, given the relative path
// stored in attendances.check_in_photo_path / check_out_photo_path. relPath
// always originates from our own SaveCheckInPhoto output, never directly
// from client input, but the containment check is kept anyway as a
// defense-in-depth guard against path traversal.
func (s LocalStore) ReadPhoto(relPath string) ([]byte, error) {
	fullPath := filepath.Join(s.root, relPath)

	absRoot, err := filepath.Abs(s.root)
	if err != nil {
		return nil, err
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(absPath, absRoot+string(filepath.Separator)) {
		return nil, ErrInvalidPath
	}

	return os.ReadFile(fullPath)
}
