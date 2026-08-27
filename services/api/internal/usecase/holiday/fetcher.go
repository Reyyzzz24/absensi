package holiday

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Event is one calendar-source holiday occurrence, source-agnostic --
// Service.SyncNational doesn't care where it came from.
type Event struct {
	Date          time.Time
	Name          string
	IsCutiBersama bool
}

// Fetcher abstracts the external calendar source so Service (and its
// tests) never depend on a live network call -- SyncNational is the only
// caller, and it's admin-triggered, never on the attendance/recap request
// path (D-25: "sistem tidak hard-depend ke API saat runtime").
type Fetcher interface {
	FetchYear(ctx context.Context, year int) ([]Event, error)
}

// GoogleCalendarFetcher parses the public "Hari libur di Indonesia" ICS
// feed Google maintains. Chosen over community REST APIs
// (api-harilibur/dayoffapi -- both returned 402 Payment Required when
// checked) and over date.nager.at (reliable but doesn't distinguish cuti
// bersama at all) -- this feed explicitly labels cuti bersama events
// ("Cuti Bersama ..." in SUMMARY), which every other free option lacked.
type GoogleCalendarFetcher struct {
	URL    string
	Client *http.Client
}

const defaultGoogleCalendarICSURL = "https://calendar.google.com/calendar/ical/id.indonesian%23holiday%40group.v.calendar.google.com/public/basic.ics"

func NewGoogleCalendarFetcher() GoogleCalendarFetcher {
	return GoogleCalendarFetcher{
		URL:    defaultGoogleCalendarICSURL,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (f GoogleCalendarFetcher) FetchYear(ctx context.Context, year int) ([]Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch holiday calendar: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch holiday calendar: unexpected status %d", resp.StatusCode)
	}
	return parseICS(resp.Body, year)
}

// parseICS extracts VEVENT DTSTART;VALUE=DATE + SUMMARY pairs for the given
// year. Deliberately line-oriented rather than a full ICS parser -- this
// feed's structure is simple and stable (all-day date events only), and a
// dependency-free ~40 line parser is easier to audit than pulling in a full
// RFC 5545 library for one field pair.
func parseICS(r io.Reader, year int) ([]Event, error) {
	scanner := bufio.NewScanner(r)
	// Some calendar lines can be long (folded lines aren't unfolded here
	// since DTSTART/SUMMARY for this feed are always single-line, but give
	// the scanner generous headroom rather than erroring on a stray long line).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var events []Event
	var curDate time.Time
	var curSummary string
	inEvent := false
	yearPrefix := fmt.Sprintf("%d", year)

	flush := func() {
		if inEvent && !curDate.IsZero() && curSummary != "" {
			events = append(events, Event{
				Date:          curDate,
				Name:          curSummary,
				IsCutiBersama: strings.HasPrefix(curSummary, "Cuti Bersama"),
			})
		}
	}

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		switch {
		case line == "BEGIN:VEVENT":
			inEvent = true
			curDate = time.Time{}
			curSummary = ""
		case line == "END:VEVENT":
			flush()
			inEvent = false
		case inEvent && strings.HasPrefix(line, "DTSTART;VALUE=DATE:"):
			raw := strings.TrimPrefix(line, "DTSTART;VALUE=DATE:")
			if !strings.HasPrefix(raw, yearPrefix) {
				continue // cheap pre-filter before parsing every date
			}
			t, err := time.Parse("20060102", raw)
			if err == nil && t.Year() == year {
				curDate = t
			}
		case inEvent && strings.HasPrefix(line, "SUMMARY:"):
			curSummary = strings.TrimPrefix(line, "SUMMARY:")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse holiday calendar: %w", err)
	}
	return events, nil
}
