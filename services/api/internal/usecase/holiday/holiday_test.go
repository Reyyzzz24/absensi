package holiday_test

import (
	"context"
	"testing"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/eprisi/absensi-next/services/api/internal/domain"
	"github.com/eprisi/absensi-next/services/api/internal/repo"
	"github.com/eprisi/absensi-next/services/api/internal/testutil"
	"github.com/eprisi/absensi-next/services/api/internal/usecase/holiday"
)

type fakeFetcher struct {
	events []holiday.Event
	err    error
}

func (f fakeFetcher) FetchYear(ctx context.Context, year int) ([]holiday.Event, error) {
	return f.events, f.err
}

func newService(t *testing.T, fetcher holiday.Fetcher) (holiday.Service, *gorm.DB) {
	t.Helper()
	gdb := testutil.NewDB(t)
	company := repo.NewCompanySettingsRepo(gdb)
	national := repo.NewNationalHolidayRepo(gdb)
	manual := repo.NewCompanyHolidayRepo(gdb)
	return holiday.NewService(company, national, manual, fetcher), gdb
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// mustCreateAdminUser relies on the "superadmin"/"admin" roles that
// migration 000001 seeds into every fresh database -- no need to create one
// here (same pattern as attendance_test.go/leave_test.go).
func mustCreateAdminUser(t *testing.T, gdb *gorm.DB) int64 {
	t.Helper()
	var role domain.Role
	if err := gdb.Where("name = ?", "superadmin").First(&role).Error; err != nil {
		t.Fatalf("find seeded superadmin role: %v", err)
	}
	u := domain.User{Name: "Admin", Username: "admin_holiday_test", PasswordHash: "x", RoleID: role.ID}
	if err := gdb.Create(&u).Error; err != nil {
		t.Fatalf("create admin user: %v", err)
	}
	return u.ID
}

func TestResolveDayStatus_Weekday_IsWorkday(t *testing.T) {
	s, _ := newService(t, fakeFetcher{})
	// 2026-08-27 is a Thursday -- default working_weekdays is Mon-Fri.
	status, err := s.ResolveDayStatus(context.Background(), date(2026, 8, 27))
	if err != nil {
		t.Fatalf("ResolveDayStatus: %v", err)
	}
	if status.IsHoliday {
		t.Fatalf("expected a weekday to not be a holiday, got %+v", status)
	}
}

func TestResolveDayStatus_DefaultWeekend_SaturdaySunday(t *testing.T) {
	s, _ := newService(t, fakeFetcher{})
	for _, d := range []time.Time{date(2026, 8, 29), date(2026, 8, 30)} { // Sat, Sun
		status, err := s.ResolveDayStatus(context.Background(), d)
		if err != nil {
			t.Fatalf("ResolveDayStatus: %v", err)
		}
		if !status.IsHoliday || status.Source != holiday.SourceWeekend {
			t.Fatalf("expected %s to resolve as weekend, got %+v", d, status)
		}
	}
}

func TestResolveDayStatus_ConfigurableWorkingWeekdays_SixDayWeek(t *testing.T) {
	s, gdb := newService(t, fakeFetcher{})
	// Only Sunday off (6-day work week) -- must not hardcode Sat/Sun.
	if err := gdb.Model(&domain.CompanySettings{}).Where("id = 1").
		Update("working_weekdays", pq.Int64Array{1, 2, 3, 4, 5, 6}).Error; err != nil {
		t.Fatalf("update working_weekdays: %v", err)
	}

	saturday := date(2026, 8, 29)
	status, err := s.ResolveDayStatus(context.Background(), saturday)
	if err != nil {
		t.Fatalf("ResolveDayStatus: %v", err)
	}
	if status.IsHoliday {
		t.Fatalf("expected Saturday to be a workday under a 6-day week, got %+v", status)
	}

	sunday := date(2026, 8, 30)
	status, err = s.ResolveDayStatus(context.Background(), sunday)
	if err != nil {
		t.Fatalf("ResolveDayStatus: %v", err)
	}
	if !status.IsHoliday || status.Source != holiday.SourceWeekend {
		t.Fatalf("expected Sunday to be the only weekend day, got %+v", status)
	}
}

func TestResolveDayStatus_NationalHoliday(t *testing.T) {
	s, gdb := newService(t, fakeFetcher{})
	independence := date(2026, 8, 17) // a Monday in 2026, would otherwise be a workday
	if err := gdb.Create(&domain.NationalHoliday{
		HolidayDate: independence, Name: "Hari Kemerdekaan RI", Year: 2026, Source: domain.HolidaySourceSync,
	}).Error; err != nil {
		t.Fatalf("seed national holiday: %v", err)
	}

	status, err := s.ResolveDayStatus(context.Background(), independence)
	if err != nil {
		t.Fatalf("ResolveDayStatus: %v", err)
	}
	if !status.IsHoliday || status.Source != holiday.SourceNational || status.Label != "Hari Kemerdekaan RI" {
		t.Fatalf("expected national holiday, got %+v", status)
	}
}

func TestResolveDayStatus_ManualCompanyHoliday_SingleAndRange(t *testing.T) {
	s, gdb := newService(t, fakeFetcher{})
	userID := mustCreateAdminUser(t, gdb)

	single := date(2026, 8, 19) // a Wednesday
	if err := gdb.Create(&domain.CompanyHoliday{
		StartDate: single, EndDate: single, Name: "Cuti bersama perusahaan", Type: domain.CompanyHolidayTypeCutiBersama, CreatedBy: userID,
	}).Error; err != nil {
		t.Fatalf("seed company holiday: %v", err)
	}

	status, err := s.ResolveDayStatus(context.Background(), single)
	if err != nil {
		t.Fatalf("ResolveDayStatus: %v", err)
	}
	if !status.IsHoliday || status.Source != holiday.SourceCompany || !status.IsCutiBersama {
		t.Fatalf("expected manual company holiday marked cuti bersama, got %+v", status)
	}

	// A day just outside the single-day holiday must be unaffected.
	before := date(2026, 8, 18)
	status, err = s.ResolveDayStatus(context.Background(), before)
	if err != nil {
		t.Fatalf("ResolveDayStatus: %v", err)
	}
	if status.IsHoliday {
		t.Fatalf("expected day before the holiday to be a workday, got %+v", status)
	}

	// A multi-day range: every day inside [start,end] must resolve holiday.
	rangeStart := date(2026, 9, 1)
	rangeEnd := date(2026, 9, 3)
	if err := gdb.Create(&domain.CompanyHoliday{
		StartDate: rangeStart, EndDate: rangeEnd, Name: "Libur ulang tahun perusahaan", Type: domain.CompanyHolidayTypeLibur, CreatedBy: userID,
	}).Error; err != nil {
		t.Fatalf("seed company holiday range: %v", err)
	}
	for _, d := range []time.Time{rangeStart, date(2026, 9, 2), rangeEnd} {
		status, err := s.ResolveDayStatus(context.Background(), d)
		if err != nil {
			t.Fatalf("ResolveDayStatus: %v", err)
		}
		if !status.IsHoliday || status.Source != holiday.SourceCompany {
			t.Fatalf("expected %s (inside range) to be a company holiday, got %+v", d, status)
		}
	}
	dayAfterRange := date(2026, 9, 4)
	status, err = s.ResolveDayStatus(context.Background(), dayAfterRange)
	if err != nil {
		t.Fatalf("ResolveDayStatus: %v", err)
	}
	if status.IsHoliday {
		t.Fatalf("expected day after the range to be a workday, got %+v", status)
	}
}

// D-25 precedence: national > company > weekend when multiple sources hit
// the same date.
func TestResolveDayStatus_Precedence_NationalWinsOverCompanyAndWeekend(t *testing.T) {
	s, gdb := newService(t, fakeFetcher{})
	userID := mustCreateAdminUser(t, gdb)

	saturday := date(2026, 8, 29) // already a weekend by default config
	if err := gdb.Create(&domain.NationalHoliday{
		HolidayDate: saturday, Name: "Libur Nasional Bertepatan Weekend", Year: 2026, Source: domain.HolidaySourceSync,
	}).Error; err != nil {
		t.Fatalf("seed national: %v", err)
	}
	if err := gdb.Create(&domain.CompanyHoliday{
		StartDate: saturday, EndDate: saturday, Name: "Entri manual di tanggal sama", Type: domain.CompanyHolidayTypeLibur, CreatedBy: userID,
	}).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}

	status, err := s.ResolveDayStatus(context.Background(), saturday)
	if err != nil {
		t.Fatalf("ResolveDayStatus: %v", err)
	}
	if status.Source != holiday.SourceNational || status.Label != "Libur Nasional Bertepatan Weekend" {
		t.Fatalf("expected national to win precedence, got %+v", status)
	}
}

func TestResolveRange_MatchesPerDateResolution(t *testing.T) {
	s, gdb := newService(t, fakeFetcher{})
	userID := mustCreateAdminUser(t, gdb)

	national := date(2026, 8, 17)
	if err := gdb.Create(&domain.NationalHoliday{HolidayDate: national, Name: "Hari Kemerdekaan RI", Year: 2026, Source: domain.HolidaySourceSync}).Error; err != nil {
		t.Fatalf("seed national: %v", err)
	}
	manualRange := date(2026, 8, 20)
	if err := gdb.Create(&domain.CompanyHoliday{
		StartDate: manualRange, EndDate: manualRange, Name: "Libur perusahaan", Type: domain.CompanyHolidayTypeLibur, CreatedBy: userID,
	}).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}

	start := date(2026, 8, 15)
	end := date(2026, 8, 31)
	rangeResult, err := s.ResolveRange(context.Background(), start, end)
	if err != nil {
		t.Fatalf("ResolveRange: %v", err)
	}

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		perDate, err := s.ResolveDayStatus(context.Background(), d)
		if err != nil {
			t.Fatalf("ResolveDayStatus(%s): %v", d, err)
		}
		key := d.Format("2006-01-02")
		got, ok := rangeResult[key]
		if !ok {
			t.Fatalf("ResolveRange missing date %s", key)
		}
		if got != perDate {
			t.Fatalf("ResolveRange(%s) = %+v, want %+v (from ResolveDayStatus)", key, got, perDate)
		}
	}
}

// D-14/B5: overnight shifts are keyed by the shift's START date (civil date
// at check-in), never the wall-clock date the checkout physically happens
// on. A Friday-night shift crossing into Saturday must resolve holiday
// status using Friday's date, not Saturday's -- otherwise an overnight
// worker would be wrongly flagged/exempted based on the wrong day. This
// resolver takes whatever date the caller passes, so the "edge case" is
// entirely about caller discipline (attendance.go already always passes
// work_date, the shift-start civil date -- see attendance.go civilDate/
// CheckInOrOut) -- verified here by confirming Friday and the following
// Saturday resolve independently/correctly, so a caller passing the wrong
// one would visibly get the wrong answer.
func TestResolveDayStatus_OvernightEdge_FridayVsSaturdayAreIndependent(t *testing.T) {
	s, _ := newService(t, fakeFetcher{})
	friday := date(2026, 8, 28)   // Friday -- default workday
	saturday := date(2026, 8, 29) // Saturday -- default weekend

	fridayStatus, err := s.ResolveDayStatus(context.Background(), friday)
	if err != nil {
		t.Fatalf("ResolveDayStatus(friday): %v", err)
	}
	if fridayStatus.IsHoliday {
		t.Fatalf("expected Friday (shift start date) to be a workday, got %+v", fridayStatus)
	}

	saturdayStatus, err := s.ResolveDayStatus(context.Background(), saturday)
	if err != nil {
		t.Fatalf("ResolveDayStatus(saturday): %v", err)
	}
	if !saturdayStatus.IsHoliday {
		t.Fatalf("expected Saturday to be a holiday, got %+v", saturdayStatus)
	}
	// The point: an overnight shift starting Friday must be evaluated
	// against fridayStatus (not a holiday), even though the checkout event
	// physically happens after midnight on what is calendar-Saturday.
}

func TestSyncNational_UpsertIsIdempotentAndPreservesManualOverride(t *testing.T) {
	s, gdb := newService(t, fakeFetcher{events: []holiday.Event{
		{Date: date(2026, 1, 1), Name: "Tahun Baru Masehi", IsCutiBersama: false},
		{Date: date(2026, 3, 20), Name: "Cuti Bersama Hari Suci Nyepi", IsCutiBersama: true},
	}})

	if n, err := s.SyncNational(context.Background(), 2026); err != nil || n != 2 {
		t.Fatalf("SyncNational: n=%d err=%v", n, err)
	}
	// Running sync again must not duplicate rows (unique constraint on
	// holiday_date would error on a naive insert).
	if n, err := s.SyncNational(context.Background(), 2026); err != nil || n != 2 {
		t.Fatalf("second SyncNational: n=%d err=%v", n, err)
	}

	rows, err := s.ListNationalHolidays(context.Background(), 2026)
	if err != nil {
		t.Fatalf("ListNationalHolidays: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected exactly 2 rows after two syncs, got %d", len(rows))
	}

	// Admin hand-edits one row -> flips to source=manual.
	var edited domain.NationalHoliday
	if err := gdb.Where("holiday_date = ?", date(2026, 1, 1)).First(&edited).Error; err != nil {
		t.Fatalf("find edited row: %v", err)
	}
	edited.Name = "Tahun Baru (dikoreksi admin)"
	if err := gdb.Save(&edited).Error; err != nil {
		t.Fatalf("save manual edit: %v", err)
	}
	if err := gdb.Model(&domain.NationalHoliday{}).Where("id = ?", edited.ID).Update("source", "manual").Error; err != nil {
		t.Fatalf("flip source to manual: %v", err)
	}

	// Sync again with the ORIGINAL upstream name -- must NOT clobber the
	// manual correction.
	if _, err := s.SyncNational(context.Background(), 2026); err != nil {
		t.Fatalf("third SyncNational: %v", err)
	}
	var after domain.NationalHoliday
	if err := gdb.Where("holiday_date = ?", date(2026, 1, 1)).First(&after).Error; err != nil {
		t.Fatalf("find after resync: %v", err)
	}
	if after.Name != "Tahun Baru (dikoreksi admin)" {
		t.Fatalf("expected manual override preserved, got name=%q", after.Name)
	}
}

func TestSyncNational_SourceFailure_LeavesCacheUntouched(t *testing.T) {
	// Deliberately share ONE db/repo set across two Service instances (one
	// per fetcher) -- newService drops+recreates the whole test database on
	// every call, so calling it twice in one test would wipe out the first
	// service's seeded data before the failure assertion below.
	gdb := testutil.NewDB(t)
	company := repo.NewCompanySettingsRepo(gdb)
	national := repo.NewNationalHolidayRepo(gdb)
	manual := repo.NewCompanyHolidayRepo(gdb)

	working := holiday.NewService(company, national, manual, fakeFetcher{events: []holiday.Event{
		{Date: date(2026, 1, 1), Name: "Tahun Baru Masehi"},
	}})
	if _, err := working.SyncNational(context.Background(), 2026); err != nil {
		t.Fatalf("seed sync: %v", err)
	}

	failing := holiday.NewService(company, national, manual, fakeFetcher{err: context.DeadlineExceeded})
	if _, err := failing.SyncNational(context.Background(), 2026); err == nil {
		t.Fatalf("expected sync to surface the fetcher error")
	}

	rows, err := working.ListNationalHolidays(context.Background(), 2026)
	if err != nil {
		t.Fatalf("ListNationalHolidays: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected the previously-cached row to survive a failed sync, got %d rows", len(rows))
	}
}
