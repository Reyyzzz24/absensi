package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/xuri/excelize/v2"

	"github.com/eprisi/absensi-next/services/api/internal/usecase/recap"
)

var indonesianMonths = [...]string{
	"", "januari", "februari", "maret", "april", "mei", "juni",
	"juli", "agustus", "september", "oktober", "november", "desember",
}

// statusLetter mirrors RecapTable.tsx's STATUS_LETTER exactly -- same
// source of truth (recap.Service.Generate), same display convention, so
// the exported file matches what's on screen instead of drifting.
var statusLetter = map[recap.DayStatus]string{
	recap.DayStatusHadir: "H",
	recap.DayStatusIzin:  "I",
	recap.DayStatusSakit: "S",
	recap.DayStatusLibur: "L",
	recap.DayStatusAlpha: "A",
}

// statusFill mirrors RecapTable.tsx's STATUS_STYLE colors (lightened for a
// spreadsheet background rather than a UI badge). Purely cosmetic -- the
// letter itself is what carries the meaning, this just matches the app.
var statusFill = map[recap.DayStatus]string{
	recap.DayStatusHadir: "#DCFCE7",
	recap.DayStatusIzin:  "#DBEAFE",
	recap.DayStatusSakit: "#F3E8FF",
	recap.DayStatusLibur: "#F1F5F9",
	recap.DayStatusAlpha: "#FEE2E2",
}

// Export streams the same recap grid rendered on screen as an .xlsx file --
// deliberately calls the identical recap.Service.Generate used by Get, so
// the exported numbers can never drift from what's on the page (single
// source of truth). Admin-only (D-7), same RBAC tier as the on-screen recap.
func (h RecapHandler) Export(w http.ResponseWriter, r *http.Request) {
	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "year is required (YYYY)")
		return
	}
	month, err := strconv.Atoi(r.URL.Query().Get("month"))
	if err != nil || month < 1 || month > 12 {
		writeError(w, http.StatusBadRequest, "month is required (1-12)")
		return
	}

	var employeeID int64
	if v := r.URL.Query().Get("employee_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid employee_id")
			return
		}
		employeeID = id
	}

	result, err := h.service.Generate(r.Context(), year, month, employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate recap")
		return
	}

	buf, err := buildRecapWorkbook(result)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build export file")
		return
	}

	monthName := "bulan"
	if month >= 1 && month <= 12 {
		monthName = indonesianMonths[month]
	}
	filename := fmt.Sprintf("rekap-absensi-%s-%d.xlsx", monthName, year)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Write(buf.Bytes())
}

func buildRecapWorkbook(result *recap.MonthRecap) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Rekap"
	f.SetSheetName("Sheet1", sheet)

	// --- styles ---
	titleStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#334155"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F1F5F9"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#CBD5E1", Style: 1},
		},
	})
	nameStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	dayStyles := make(map[recap.DayStatus]int, len(statusFill))
	for status, color := range statusFill {
		id, _ := f.NewStyle(&excelize.Style{
			Fill:      excelize.Fill{Type: "pattern", Color: []string{color}, Pattern: 1},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		})
		dayStyles[status] = id
	}
	lateStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#FEF3C7"}, Pattern: 1},
		Font:      &excelize.Font{Bold: true, Color: "#B45309"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	// --- title rows ---
	monthName := "-"
	if result.Month >= 1 && result.Month <= 12 {
		monthName = indonesianMonths[result.Month]
	}
	f.SetCellValue(sheet, "A1", "Rekap Absensi Bulanan")
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	f.SetCellValue(sheet, "A2", fmt.Sprintf("Periode: %s %d", capitalize(monthName), result.Year))

	// --- header row (row 4) ---
	const headerRow = 4
	col := 1
	setHeader := func(v string) {
		cell, _ := excelize.CoordinatesToCellName(col, headerRow)
		f.SetCellValue(sheet, cell, v)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
		col++
	}
	setHeader("Karyawan")
	setHeader("NIK")
	for d := 1; d <= result.DaysInMonth; d++ {
		setHeader(strconv.Itoa(d))
	}
	setHeader("H")
	setHeader("Telat")
	setHeader("I")
	setHeader("S")
	setHeader("A")

	// --- data rows ---
	row := headerRow + 1
	for _, emp := range result.Employees {
		c := 1
		nameCell, _ := excelize.CoordinatesToCellName(c, row)
		f.SetCellValue(sheet, nameCell, emp.FullName)
		f.SetCellStyle(sheet, nameCell, nameCell, nameStyle)
		c++

		nikCell, _ := excelize.CoordinatesToCellName(c, row)
		f.SetCellValue(sheet, nikCell, emp.NIK)
		f.SetCellStyle(sheet, nikCell, nikCell, centerStyle)
		c++

		for _, day := range emp.Days {
			cell, _ := excelize.CoordinatesToCellName(c, row)
			if day.IsLate {
				f.SetCellValue(sheet, cell, "*")
				f.SetCellStyle(sheet, cell, cell, lateStyle)
			} else {
				f.SetCellValue(sheet, cell, statusLetter[day.Status])
				if styleID, ok := dayStyles[day.Status]; ok {
					f.SetCellStyle(sheet, cell, cell, styleID)
				}
			}
			c++
		}

		summaryCols := []string{"hadir", "telat", "izin", "sakit", "alpha"}
		for _, key := range summaryCols {
			cell, _ := excelize.CoordinatesToCellName(c, row)
			f.SetCellValue(sheet, cell, emp.Summary[key])
			f.SetCellStyle(sheet, cell, cell, centerStyle)
			c++
		}
		row++
	}

	// --- column widths ---
	f.SetColWidth(sheet, "A", "A", 26)
	f.SetColWidth(sheet, "B", "B", 12)
	firstDayCol, _ := excelize.ColumnNumberToName(3)
	lastDayCol, _ := excelize.ColumnNumberToName(2 + result.DaysInMonth)
	f.SetColWidth(sheet, firstDayCol, lastDayCol, 4)
	firstSummaryCol, _ := excelize.ColumnNumberToName(3 + result.DaysInMonth)
	lastSummaryCol, _ := excelize.ColumnNumberToName(7 + result.DaysInMonth)
	f.SetColWidth(sheet, firstSummaryCol, lastSummaryCol, 8)

	// --- freeze header row + name/NIK columns (matches the sticky
	// header/first-column behavior in the on-screen table) ---
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      2,
		YSplit:      headerRow,
		TopLeftCell: "C5",
		ActivePane:  "bottomRight",
	})

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}
