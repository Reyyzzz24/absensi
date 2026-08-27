// Package domain holds the GORM entities mapped onto the schema defined in
// services/api/migrations. Schema is managed exclusively via golang-migrate;
// these structs are for querying only (no gorm.AutoMigrate calls anywhere).
//
// json tags follow the snake_case field names already documented in
// docs/openapi.yaml -- these structs are returned directly by several
// handlers (writeJSON), so the tag is the actual wire contract, not just
// decoration.
package domain

import "time"

type Role struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Email        *string   `gorm:"uniqueIndex" json:"email,omitempty"`
	Username     string    `gorm:"uniqueIndex" json:"username"`
	PasswordHash string    `json:"-"`
	RoleID       int64     `json:"role_id"`
	Role         Role      `gorm:"foreignKey:RoleID" json:"role"`
	Phone        *string   `json:"phone,omitempty"` // added for self-service profile editing parity with Employee
	PhotoPath    *string   `json:"photo_path,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Department struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"uniqueIndex" json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Employee struct {
	ID           int64       `gorm:"primaryKey" json:"id"`
	NIK          string      `gorm:"column:nik;uniqueIndex" json:"nik"`
	FullName     string      `json:"full_name"`
	PasswordHash string      `json:"-"`
	DepartmentID *int64      `json:"department_id,omitempty"`
	Department   *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Position     *string     `json:"position,omitempty"`
	Phone        *string     `json:"phone,omitempty"`
	PhotoPath    *string     `json:"photo_path,omitempty"`
	IsActive     bool        `json:"is_active"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// Shift maps legacy jam_kerja. IsOvernight is a DB-generated column
// (end_time < start_time) -- see migration 000004 and DECISIONS.md D-14.
type Shift struct {
	ID               int64     `gorm:"primaryKey" json:"id"`
	Code             string    `gorm:"uniqueIndex" json:"code"`
	Name             string    `json:"name"`
	IsDayOff         bool      `json:"is_day_off"`
	StartTime        *string   `json:"start_time,omitempty"` // TIME as HH:MM:SS string; kept simple, no civil-time lib dependency yet
	EndTime          *string   `json:"end_time,omitempty"`
	IsOvernight      bool      `gorm:"->" json:"is_overnight"` // read-only, DB-generated
	LateGraceMinutes int       `json:"late_grace_minutes"`     // admin/superadmin-configurable, D-10
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type WeeklyShiftDefault struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	EmployeeID int64     `json:"employee_id"`
	DayOfWeek  int16     `json:"day_of_week"` // 0=Sunday
	ShiftID    int64     `json:"shift_id"`
	Shift      Shift     `gorm:"foreignKey:ShiftID" json:"shift"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// WorkSchedule maps legacy jadwal_kerja: per-date shift assignment. Takes
// precedence over WeeklyShiftDefault for the same employee+date -- D-18.
type WorkSchedule struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	EmployeeID int64     `json:"employee_id"`
	WorkDate   time.Time `gorm:"type:date" json:"work_date"`
	ShiftID    int64     `json:"shift_id"`
	Shift      Shift     `gorm:"foreignKey:ShiftID" json:"shift"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type OfficeLocation struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	RadiusMeters int       `json:"radius_meters"` // enforced server-side now, D-1 (legacy computed but never enforced)
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FieldAssignment is the "dinas luar" pre-approval -- D-21. When one exists
// for an employee+date, the check-in geofence check is bypassed for that day.
type FieldAssignment struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	EmployeeID int64     `json:"employee_id"`
	WorkDate   time.Time `gorm:"type:date" json:"work_date"`
	Note       *string   `json:"note,omitempty"`
	ApprovedBy int64     `json:"approved_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// RevokedRefreshToken is a denylist entry for a refresh JWT's jti, recorded
// on /auth/logout so that specific refresh token can no longer be used with
// /auth/refresh before it naturally expires (D-24 follow-up to the Phase 2
// stateless-refresh-token simplification).
type RevokedRefreshToken struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	JTI       string    `gorm:"column:jti;uniqueIndex" json:"jti"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type AttendanceStatus string

const (
	AttendanceStatusOpen              AttendanceStatus = "open"
	AttendanceStatusClosed            AttendanceStatus = "closed"
	AttendanceStatusFlaggedNoCheckout AttendanceStatus = "flagged_no_checkout"
)

// Attendance maps legacy `presensi`, unifying the two legacy parallel flows
// (PresensiController + AbsensiController/EOS) -- D-8. WorkDate is the
// shift's start date; for overnight shifts a checkout after midnight still
// updates this same row instead of creating a new one -- D-14.
type Attendance struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	EmployeeID  int64     `json:"employee_id"`
	Employee    *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	WorkDate    time.Time `gorm:"type:date" json:"work_date"`
	CycleNumber int       `json:"cycle_number"` // supports multiple in/out cycles per day, D-23 (legacy GA department)
	ShiftID     *int64    `json:"shift_id,omitempty"`
	IsWFH       bool      `json:"is_wfh"`

	OfficeLocationID *int64 `json:"office_location_id,omitempty"`

	CheckInAt        *time.Time `json:"check_in_at,omitempty"`
	CheckInLat       *float64   `json:"check_in_lat,omitempty"`
	CheckInLng       *float64   `json:"check_in_lng,omitempty"`
	CheckInDistanceM *int       `json:"check_in_distance_m,omitempty"`
	CheckInPhotoPath *string    `json:"check_in_photo_path,omitempty"`
	IsLate           *bool      `json:"is_late,omitempty"`

	CheckOutAt        *time.Time `json:"check_out_at,omitempty"`
	CheckOutLat       *float64   `json:"check_out_lat,omitempty"`
	CheckOutLng       *float64   `json:"check_out_lng,omitempty"`
	CheckOutDistanceM *int       `json:"check_out_distance_m,omitempty"`
	CheckOutPhotoPath *string    `json:"check_out_photo_path,omitempty"`
	IsEarlyLeave      *bool      `json:"is_early_leave,omitempty"`

	Status AttendanceStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LeaveType string

const (
	LeaveTypeIzin  LeaveType = "izin"
	LeaveTypeSakit LeaveType = "sakit"
)

type LeaveStatus string

const (
	LeaveStatusPending  LeaveStatus = "pending"
	LeaveStatusApproved LeaveStatus = "approved"
	LeaveStatusRejected LeaveStatus = "rejected"
)

// LeaveRequest maps legacy pengajuan_izin. Merges the legacy separate
// izin/sakit forms into one (D-16) and adds an approval workflow the legacy
// system never had (D-15). Approved requests are excluded from "absent" in
// attendance recap reports (enforced in report query logic, not here).
type LeaveRequest struct {
	ID         int64       `gorm:"primaryKey" json:"id"`
	EmployeeID int64       `json:"employee_id"`
	Employee   *Employee   `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Type       LeaveType   `json:"type"`
	StartDate  time.Time   `gorm:"type:date" json:"start_date"`
	EndDate    time.Time   `gorm:"type:date" json:"end_date"`
	Reason     *string     `json:"reason,omitempty"`
	Status     LeaveStatus `json:"status"`
	ReviewedBy *int64      `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time  `json:"reviewed_at,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// Task maps the legacy `tasks` table. Ownership must always be checked at
// the handler/usecase layer (employee_id = current user) -- legacy had an
// IDOR here (edit_task/update_task with no ownership check) -- D-5.
type Task struct {
	ID         int64      `gorm:"primaryKey" json:"id"`
	EmployeeID int64      `json:"employee_id"`
	Title      string     `json:"title"`
	Detail     *string    `json:"detail,omitempty"`
	StartsAt   time.Time  `json:"starts_at"`
	EndsAt     *time.Time `json:"ends_at,omitempty"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Notification is a generic in-app notification for either audience.
// RecipientAudience + RecipientID together identify the owner the same way
// JWT claims do (employee id vs admin user id) -- every query MUST filter by
// both, never RecipientID alone, to stay IDOR-safe (D-5 pattern: an employee
// id and an admin user id can collide numerically since they're separate
// tables/sequences).
type Notification struct {
	ID                int64      `gorm:"primaryKey" json:"id"`
	RecipientAudience string     `json:"recipient_audience"` // "employee" | "admin"
	RecipientID       int64      `json:"recipient_id"`
	Type              string     `json:"type"`
	Title             string     `json:"title"`
	Body              *string    `json:"body,omitempty"`
	Link              *string    `json:"link,omitempty"`
	ReadAt            *time.Time `json:"read_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// CompanySettings is a deliberate single-row table (id is always 1) -- this
// app has no multi-tenant concept (see admin-shell's static workspace card),
// so there is exactly one company profile, not a list of them.
type CompanySettings struct {
	ID       int64   `gorm:"primaryKey" json:"id"`
	Name     string  `json:"name"`
	LogoPath *string `json:"logo_path,omitempty"`
}

// NotificationPreference is opt-out per type: a missing row means enabled.
type NotificationPreference struct {
	RecipientAudience string `gorm:"primaryKey" json:"recipient_audience"`
	RecipientID       int64  `gorm:"primaryKey" json:"recipient_id"`
	Type              string `gorm:"primaryKey" json:"type"`
	Enabled           bool   `json:"enabled"`
}
