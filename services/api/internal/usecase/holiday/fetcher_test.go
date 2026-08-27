package holiday

import (
	"strings"
	"testing"
	"time"
)

const sampleICS = `BEGIN:VCALENDAR
PRODID:-//Google Inc//Google Calendar 70.9054//EN
VERSION:2.0
BEGIN:VEVENT
DTSTART;VALUE=DATE:20260101
DTEND;VALUE=DATE:20260102
SUMMARY:Tahun Baru Masehi
STATUS:CONFIRMED
END:VEVENT
BEGIN:VEVENT
DTSTART;VALUE=DATE:20260320
DTEND;VALUE=DATE:20260321
SUMMARY:Cuti Bersama Hari Suci Nyepi (Tahun Baru Saka)
STATUS:CONFIRMED
END:VEVENT
BEGIN:VEVENT
DTSTART;VALUE=DATE:20250817
DTEND;VALUE=DATE:20250818
SUMMARY:Hari Kemerdekaan RI
STATUS:CONFIRMED
END:VEVENT
END:VCALENDAR
`

func TestParseICS_ExtractsMatchingYearOnly(t *testing.T) {
	events, err := parseICS(strings.NewReader(sampleICS), 2026)
	if err != nil {
		t.Fatalf("parseICS: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events for 2026 (2025 event filtered out), got %d: %+v", len(events), events)
	}
}

func TestParseICS_DetectsCutiBersamaPrefix(t *testing.T) {
	events, err := parseICS(strings.NewReader(sampleICS), 2026)
	if err != nil {
		t.Fatalf("parseICS: %v", err)
	}
	var newYear, nyepi *Event
	for i := range events {
		switch events[i].Name {
		case "Tahun Baru Masehi":
			newYear = &events[i]
		case "Cuti Bersama Hari Suci Nyepi (Tahun Baru Saka)":
			nyepi = &events[i]
		}
	}
	if newYear == nil || nyepi == nil {
		t.Fatalf("expected both events present, got %+v", events)
	}
	if newYear.IsCutiBersama {
		t.Fatalf("New Year is a fixed public holiday, not cuti bersama")
	}
	if !nyepi.IsCutiBersama {
		t.Fatalf("expected 'Cuti Bersama ...' event to be flagged is_cuti_bersama")
	}
	if !nyepi.Date.Equal(time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected date parsed: %v", nyepi.Date)
	}
}
