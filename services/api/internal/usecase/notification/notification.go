// Package notification implements the MVP-polling in-app notification
// center (bell in both shells). No websocket/push transport exists yet --
// the frontend polls GET /notifications/unread-count every 30-60s and
// refetches the list when the panel opens. Swapping in a push transport
// later only touches the frontend polling loop; this package's shape
// (poll-friendly list + unread-count + mark-read) does not need to change.
package notification

import (
	"context"
	"errors"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/platform/authtoken"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

var ErrUnknownType = errors.New("unknown notification type")

// Known notification types -- the single set both the backend (creation +
// preference enforcement) and frontend (preference toggles) agree on. Add
// new types here as new triggers are wired up.
const (
	TypeLeaveStatusChange = "leave_status_change"
)

var KnownTypes = []string{TypeLeaveStatusChange}

type Service struct {
	notifications repo.NotificationRepo
	preferences   repo.NotificationPreferenceRepo
}

func NewService(notifications repo.NotificationRepo, preferences repo.NotificationPreferenceRepo) Service {
	return Service{notifications: notifications, preferences: preferences}
}

// Notify creates a notification unless the recipient has opted out of this
// type via preferences (opt-out model, D-15-adjacent: notifying about a
// status change should not be forced on someone who disabled that type).
func (s Service) Notify(ctx context.Context, aud authtoken.Audience, recipientID int64, notifType, title, body, link string) error {
	enabled, err := s.preferences.IsEnabled(ctx, string(aud), recipientID, notifType)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	n := &domain.Notification{
		RecipientAudience: string(aud),
		RecipientID:       recipientID,
		Type:              notifType,
		Title:             title,
	}
	if body != "" {
		n.Body = &body
	}
	if link != "" {
		n.Link = &link
	}
	return s.notifications.Create(ctx, n)
}

func (s Service) List(ctx context.Context, aud authtoken.Audience, id int64, limit, offset int) ([]domain.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.notifications.ListForRecipient(ctx, string(aud), id, limit, offset)
}

func (s Service) UnreadCount(ctx context.Context, aud authtoken.Audience, id int64) (int64, error) {
	return s.notifications.UnreadCount(ctx, string(aud), id)
}

func (s Service) MarkRead(ctx context.Context, aud authtoken.Audience, id, notificationID int64) error {
	return s.notifications.MarkRead(ctx, string(aud), id, notificationID)
}

func (s Service) MarkAllRead(ctx context.Context, aud authtoken.Audience, id int64) error {
	return s.notifications.MarkAllRead(ctx, string(aud), id)
}

// Preferences merges saved rows against KnownTypes so the frontend always
// gets one toggle per known type, defaulted to enabled when no row exists.
func (s Service) Preferences(ctx context.Context, aud authtoken.Audience, id int64) (map[string]bool, error) {
	saved, err := s.preferences.List(ctx, string(aud), id)
	if err != nil {
		return nil, err
	}
	byType := make(map[string]bool, len(saved))
	for _, p := range saved {
		byType[p.Type] = p.Enabled
	}
	out := make(map[string]bool, len(KnownTypes))
	for _, t := range KnownTypes {
		if v, ok := byType[t]; ok {
			out[t] = v
		} else {
			out[t] = true
		}
	}
	return out, nil
}

func (s Service) SetPreference(ctx context.Context, aud authtoken.Audience, id int64, notifType string, enabled bool) error {
	known := false
	for _, t := range KnownTypes {
		if t == notifType {
			known = true
			break
		}
	}
	if !known {
		return ErrUnknownType
	}
	return s.preferences.Set(ctx, string(aud), id, notifType, enabled)
}
