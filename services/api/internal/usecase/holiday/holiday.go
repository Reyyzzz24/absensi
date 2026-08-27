// Package holiday implements the single holiday resolver combining the
// three sources agreed in docs/DECISIONS.md D-25: configurable weekend,
// nationally-synced holidays (incl. cuti bersama), and manual
// company-specific holidays. Deliberately does NOT generate any attendance
// rows for holidays -- callers (recap, check-in) resolve status on demand,
// so a later policy change (weekend redefined, an emergency cuti bersama)
// never leaves stale generated data to clean up.
package holiday

import (
	"context"
	"errors"
	"time"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
)

var ErrNotFound = errors.New("not found")

type Source string

const (
	SourceNone     Source = "none"
	SourceWeekend  Source = "weekend"
	SourceNational Source = "national"
	SourceCompany  Source = "company"
)

// DayStatus is the resolver's single output shape. Precedence when more
// than one source matches the same date (D-25): national > company >
// weekend -- national is the most authoritative/specific label, company
// holidays exist mainly to ADD holidays the national sync doesn't have
// (e.g. a cuti bersama not yet synced, or a company-only closure), and
// weekend is the least specific fallback.
type DayStatus struct {
	IsHoliday     bool   `json:"is_holiday"`
	Source        Source `json:"source"`
	Label         string `json:"label,omitempty"`
	IsCutiBersama bool   `json:"is_cuti_bersama,omitempty"`
}

var workday = DayStatus{IsHoliday: false, Source: SourceNone}

type Service struct {
	company  repo.CompanySettingsRepo
	national repo.NationalHolidayRepo
	manual   repo.CompanyHolidayRepo
	fetcher  Fetcher
}

func NewService(company repo.CompanySettingsRepo, national repo.NationalHolidayRepo, manual repo.CompanyHolidayRepo, fetcher Fetcher) Service {
	return Service{company: company, national: national, manual: manual, fetcher: fetcher}
}

// ResolveDayStatus resolves a single date. Prefer ResolveRange when
// resolving many dates at once (recap grids, calendar views) -- this does
// up to 3 queries per call, one of which (company_settings) never changes
// per date and is wasteful to repeat in a loop.
func (s Service) ResolveDayStatus(ctx context.Context, date time.Time) (DayStatus, error) {
	settings, err := s.company.Get(ctx)
	if err != nil {
		return DayStatus{}, err
	}
	national, err := s.national.FindByDate(ctx, date)
	if err != nil {
		return DayStatus{}, err
	}
	companyHolidays, err := s.manual.ListOverlapping(ctx, date, date)
	if err != nil {
		return DayStatus{}, err
	}
	return resolve(date, settings.WorkingWeekdays, national, companyHolidays), nil
}

// ResolveRange resolves every date in [start, end] (inclusive) in three
// queries total, regardless of range length -- the shape recap/calendar
// endpoints need, since all three sources are company-wide (not
// per-employee), so this only needs to run ONCE per report, not once per
// employee. Keyed by "2006-01-02".
func (s Service) ResolveRange(ctx context.Context, start, end time.Time) (map[string]DayStatus, error) {
	settings, err := s.company.Get(ctx)
	if err != nil {
		return nil, err
	}
	nationalRows, err := s.national.ListByDateRange(ctx, start, end)
	if err != nil {
		return nil, err
	}
	nationalByDate := make(map[string]domain.NationalHoliday, len(nationalRows))
	for _, n := range nationalRows {
		nationalByDate[n.HolidayDate.Format("2006-01-02")] = n
	}

	companyRows, err := s.manual.ListOverlapping(ctx, start, end)
	if err != nil {
		return nil, err
	}

	out := make(map[string]DayStatus)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		var national *domain.NationalHoliday
		if n, ok := nationalByDate[key]; ok {
			national = &n
		}
		var overlapping []domain.CompanyHoliday
		for _, c := range companyRows {
			if !d.Before(c.StartDate) && !d.After(c.EndDate) {
				overlapping = append(overlapping, c)
			}
		}
		out[key] = resolve(d, settings.WorkingWeekdays, national, overlapping)
	}
	return out, nil
}

func resolve(date time.Time, workingWeekdays []int64, national *domain.NationalHoliday, companyHolidays []domain.CompanyHoliday) DayStatus {
	if national != nil {
		return DayStatus{IsHoliday: true, Source: SourceNational, Label: national.Name, IsCutiBersama: national.IsCutiBersama}
	}
	if len(companyHolidays) > 0 {
		ch := companyHolidays[0]
		return DayStatus{IsHoliday: true, Source: SourceCompany, Label: ch.Name, IsCutiBersama: ch.Type == domain.CompanyHolidayTypeCutiBersama}
	}
	if isWeekend(date, workingWeekdays) {
		return DayStatus{IsHoliday: true, Source: SourceWeekend, Label: "Akhir pekan"}
	}
	return workday
}

// isWeekend reports whether date's ISO weekday (1=Monday..7=Sunday) is
// absent from the configured working days.
func isWeekend(date time.Time, workingWeekdays []int64) bool {
	iso := isoWeekday(date)
	for _, d := range workingWeekdays {
		if int(d) == iso {
			return false
		}
	}
	return true
}

func isoWeekday(date time.Time) int {
	wd := int(date.Weekday()) // time.Sunday == 0
	if wd == 0 {
		return 7
	}
	return wd
}

// --- Company manual holiday CRUD (thin passthrough, admin-only route) ---

func (s Service) ListCompanyHolidays(ctx context.Context) ([]domain.CompanyHoliday, error) {
	return s.manual.List(ctx)
}

type CompanyHolidayInput struct {
	StartDate time.Time
	EndDate   time.Time
	Name      string
	Type      domain.CompanyHolidayType
	Note      string
	CreatedBy int64
}

func (s Service) CreateCompanyHoliday(ctx context.Context, in CompanyHolidayInput) (*domain.CompanyHoliday, error) {
	h := &domain.CompanyHoliday{
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
		Name:      in.Name,
		Type:      in.Type,
		CreatedBy: in.CreatedBy,
	}
	if in.Note != "" {
		h.Note = &in.Note
	}
	if err := s.manual.Create(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

func (s Service) UpdateCompanyHoliday(ctx context.Context, id int64, in CompanyHolidayInput) (*domain.CompanyHoliday, error) {
	existing, err := s.manual.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	existing.StartDate = in.StartDate
	existing.EndDate = in.EndDate
	existing.Name = in.Name
	existing.Type = in.Type
	if in.Note != "" {
		existing.Note = &in.Note
	} else {
		existing.Note = nil
	}
	if err := s.manual.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s Service) DeleteCompanyHoliday(ctx context.Context, id int64) error {
	return s.manual.Delete(ctx, id)
}

// --- National holiday sync (admin-triggered, never on the request path) ---

func (s Service) SyncNational(ctx context.Context, year int) (synced int, err error) {
	events, err := s.fetcher.FetchYear(ctx, year)
	if err != nil {
		// Graceful degradation: the sync call fails, but every previously
		// cached row is untouched -- callers keep using the existing cache
		// (ResolveDayStatus/ResolveRange never call the fetcher).
		return 0, err
	}
	for _, ev := range events {
		h := &domain.NationalHoliday{
			HolidayDate:   ev.Date,
			Name:          ev.Name,
			Year:          int16(year),
			IsCutiBersama: ev.IsCutiBersama,
		}
		if err := s.national.UpsertSynced(ctx, h); err != nil {
			return synced, err
		}
		synced++
	}
	return synced, nil
}

func (s Service) ListNationalHolidays(ctx context.Context, year int) ([]domain.NationalHoliday, error) {
	return s.national.ListByYear(ctx, year)
}
