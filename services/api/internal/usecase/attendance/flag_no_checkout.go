package attendance

import (
	"context"
	"time"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
)

// FlagStaleOpenAttendances marks attendance rows still `open` (no checkout)
// well past their shift as `flagged_no_checkout` for admin review -- D-17.
// It never fabricates a checkout time; legacy left these rows dangling
// forever with no visibility at all (LOGIC_SPEC.md §14).
//
// Intended to run once daily via a scheduler (Phase 5 DevOps concern -- no
// cron wiring exists yet, per CLAUDE.md §1.5 this is scaffolded but not
// activated until that phase).
func (s Service) FlagStaleOpenAttendances(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().In(jakarta).Add(-olderThan)

	result := s.db.WithContext(ctx).
		Model(&domain.Attendance{}).
		Where("status = ? AND check_in_at IS NOT NULL AND check_in_at < ?", domain.AttendanceStatusOpen, cutoff).
		Update("status", domain.AttendanceStatusFlaggedNoCheckout)

	return result.RowsAffected, result.Error
}
