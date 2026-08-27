// Shared TypeScript types mirroring the JSON shapes documented in
// docs/openapi.yaml. Extracted from apps/web/lib/types.ts (Phase 3) so a
// future mobile app (Phase 6) can depend on the same source of truth
// instead of re-declaring these shapes against the same API.

export type AttendanceStatus = "open" | "closed" | "flagged_no_checkout";

export type Attendance = {
  id: number;
  employee_id: number;
  employee?: { id: number; nik: string; full_name: string };
  work_date: string;
  cycle_number: number;
  shift_id?: number;
  is_wfh: boolean;
  office_location_id?: number;
  check_in_at?: string;
  check_in_lat?: number;
  check_in_lng?: number;
  check_in_distance_m?: number;
  check_in_photo_path?: string;
  is_late?: boolean;
  check_out_at?: string;
  check_out_lat?: number;
  check_out_lng?: number;
  check_out_distance_m?: number;
  check_out_photo_path?: string;
  is_early_leave?: boolean;
  status: AttendanceStatus;
};

export type LeaveType = "izin" | "sakit";
export type LeaveStatus = "pending" | "approved" | "rejected";

export type LeaveRequest = {
  id: number;
  employee_id: number;
  employee?: { id: number; nik: string; full_name: string };
  type: LeaveType;
  start_date: string;
  end_date: string;
  reason?: string;
  status: LeaveStatus;
  reviewed_by?: number;
  reviewed_at?: string;
};

export type Task = {
  id: number;
  employee_id: number;
  title: string;
  detail?: string;
  starts_at: string;
  ends_at?: string;
  status: string;
};

export type OfficeLocation = {
  id: number;
  name: string;
  latitude: number;
  longitude: number;
  radius_meters: number;
  is_active: boolean;
};

export type Shift = {
  id: number;
  code: string;
  name: string;
  is_day_off: boolean;
  start_time?: string;
  end_time?: string;
  is_overnight: boolean;
  late_grace_minutes: number;
};

export type Department = {
  id: number;
  code: string;
  name: string;
};

export type Employee = {
  id: number;
  nik: string;
  full_name: string;
  department_id?: number;
  department?: Department;
  position?: string;
  phone?: string;
  is_active: boolean;
};

export type RecapDayStatus = "hadir" | "izin" | "sakit" | "libur" | "alpha";

export type RecapDay = {
  date: string;
  status: RecapDayStatus;
  is_late?: boolean;
  // Set even when status="hadir" -- e.g. voluntary/overtime work on a
  // libur day (D-25 holiday resolver).
  is_holiday?: boolean;
};

export type EmployeeRecap = {
  employee_id: number;
  nik: string;
  full_name: string;
  days: RecapDay[];
  summary: Record<string, number>;
};

export type MonthRecap = {
  year: number;
  month: number;
  days_in_month: number;
  employees: EmployeeRecap[];
};

export type FieldAssignment = {
  id: number;
  employee_id: number;
  work_date: string;
  note?: string;
  approved_by: number;
};

// Shared self-service types (GET/PATCH /api/me, notifications) -- same
// shape regardless of whether the caller is an employee or admin token.
export type Profile = {
  id: number;
  audience: "employee" | "admin";
  name: string;
  identifier: string; // NIK for employee, username for admin
  email?: string;
  phone?: string;
  department?: string;
  position?: string;
  role?: string;
  photo_path?: string;
};

export type Notification = {
  id: number;
  recipient_audience: "employee" | "admin";
  recipient_id: number;
  type: string;
  title: string;
  body?: string;
  link?: string;
  read_at?: string;
  created_at: string;
};

export type NotificationPreferences = Record<string, boolean>;

export type CompanySettings = {
  id: number;
  name: string;
  logo_path?: string;
  // ISO weekday numbers (1=Monday..7=Sunday) considered work days -- D-25
  // holiday resolver. Default {1,2,3,4,5}; a 6-day work week is just a
  // different array, not hardcoded Sat/Sun.
  working_weekdays: number[];
};

// --- Holiday management (D-25) ---
// Three sources combined by one resolver: configurable weekend, nationally
// synced holidays (incl. cuti bersama), and manual company holidays.
// Precedence when sources overlap: national > company > weekend.

export type HolidaySource = "none" | "weekend" | "national" | "company";

export type HolidayDayStatus = {
  is_holiday: boolean;
  source: HolidaySource;
  label?: string;
  is_cuti_bersama?: boolean;
};

// GET /holidays/calendar and /admin/holidays/calendar response item.
export type HolidayCalendarDay = {
  date: string; // YYYY-MM-DD
} & HolidayDayStatus;

export type NationalHoliday = {
  id: number;
  holiday_date: string; // YYYY-MM-DD
  name: string;
  year: number;
  is_cuti_bersama: boolean;
  source: "sync" | "manual";
  created_at: string;
  updated_at: string;
};

export type CompanyHolidayType = "libur" | "cuti_bersama";

export type CompanyHoliday = {
  id: number;
  start_date: string; // YYYY-MM-DD
  end_date: string; // YYYY-MM-DD
  name: string;
  type: CompanyHolidayType;
  note?: string;
  created_by: number;
  created_at: string;
  updated_at: string;
};
