import "server-only";
import { adminApi } from "@/lib/authedApi";
import type { Attendance, Employee, LeaveRequest, MonthRecap } from "@/lib/types";

function isoDate(d: Date) {
  return d.toISOString().slice(0, 10);
}

function pctChange(today: number, yesterday: number): { direction: "up" | "down"; value: string } | null {
  if (yesterday === 0 && today === 0) return null;
  if (yesterday === 0) return { direction: "up", value: "Baru" };
  const pct = Math.round(((today - yesterday) / yesterday) * 100);
  if (pct === 0) return null;
  return { direction: pct > 0 ? "up" : "down", value: `${Math.abs(pct)}%` };
}

function overlapsDate(lr: LeaveRequest, dateISO: string) {
  return lr.status === "approved" && lr.start_date.slice(0, 10) <= dateISO && lr.end_date.slice(0, 10) >= dateISO;
}

export type DashboardData = {
  employees: Employee[];
  kpis: {
    totalKaryawan: number;
    hadirHariIni: number;
    hadirTrend: ReturnType<typeof pctChange>;
    terlambatHariIni: number;
    terlambatTrend: ReturnType<typeof pctChange>;
    izinHariIni: number;
    izinTrend: ReturnType<typeof pctChange>;
  };
  statusDistribution: { hadir: number; telat: number; izin: number; alpha: number };
  departmentAttendance: { name: string; value: number }[];
  recentAttendance: Attendance[];
};

// Assembles the dashboard entirely from endpoints that already exist --
// there is no dedicated "/admin/dashboard/summary" endpoint. Today's status
// breakdown and per-department attendance both reuse the SAME monthly recap
// endpoint the /admin/reports/recap page already calls, rather than
// re-deriving hadir/izin/alpha classification logic in the frontend (that
// logic -- shift/libur-awareness, approved-leave exclusion -- lives once in
// internal/usecase/recap and should stay there).
export async function loadDashboardData(): Promise<DashboardData> {
  const now = new Date();
  const today = isoDate(now);
  const yesterday = isoDate(new Date(now.getTime() - 24 * 60 * 60 * 1000));
  const year = now.getFullYear();
  const month = now.getMonth() + 1;
  const todayDay = now.getDate();

  const [employees, todayAttendance, yesterdayAttendance, allLeave, recap] = await Promise.all([
    adminApi.get<Employee[]>("/api/v1/admin/employees"),
    adminApi.get<Attendance[]>(`/api/v1/admin/attendance/monitoring?date=${today}`),
    adminApi.get<Attendance[]>(`/api/v1/admin/attendance/monitoring?date=${yesterday}`),
    adminApi.get<LeaveRequest[]>("/api/v1/admin/leave-requests"),
    adminApi.get<MonthRecap>(`/api/v1/admin/reports/recap?year=${year}&month=${month}`),
  ]);

  const employeeById = new Map(employees.map((e) => [e.id, e]));

  const lateToday = todayAttendance.filter((a) => a.is_late).length;
  const lateYesterday = yesterdayAttendance.filter((a) => a.is_late).length;
  const izinToday = allLeave.filter((lr) => overlapsDate(lr, today)).length;
  const izinYesterday = allLeave.filter((lr) => overlapsDate(lr, yesterday)).length;

  // Today's status breakdown + per-department hadir count, both derived from
  // the recap grid's classification for "today" (index todayDay - 1) across
  // every employee -- the same source of truth as the recap page itself.
  let hadir = 0;
  let telat = 0;
  let izinFromRecap = 0;
  let alpha = 0;
  const byDepartment = new Map<string, number>();

  for (const empRecap of recap.employees) {
    const day = empRecap.days[todayDay - 1];
    if (!day || day.status === "libur") continue;
    if (day.status === "hadir") {
      if (day.is_late) telat++;
      else hadir++;
      const dept = employeeById.get(empRecap.employee_id)?.department?.name ?? "Tanpa Departemen";
      byDepartment.set(dept, (byDepartment.get(dept) ?? 0) + 1);
    } else if (day.status === "izin" || day.status === "sakit") {
      izinFromRecap++;
    } else if (day.status === "alpha") {
      alpha++;
    }
  }

  const departmentAttendance = [...byDepartment.entries()]
    .map(([name, value]) => ({ name, value }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 6);

  const recentAttendance = [...todayAttendance]
    .sort((a, b) => (b.check_in_at ?? "").localeCompare(a.check_in_at ?? ""))
    .slice(0, 6);

  return {
    employees,
    kpis: {
      totalKaryawan: employees.length,
      hadirHariIni: todayAttendance.length,
      hadirTrend: pctChange(todayAttendance.length, yesterdayAttendance.length),
      terlambatHariIni: lateToday,
      terlambatTrend: pctChange(lateToday, lateYesterday),
      izinHariIni: izinToday,
      izinTrend: pctChange(izinToday, izinYesterday),
    },
    // Prefer the recap-derived izin count (accounts for approved leave
    // classification consistently with the rest of the app); izinToday
    // above (raw leave-requests overlap) is used only for the KPI trend
    // since recap for "yesterday" would require a second month's fetch
    // when yesterday crosses a month boundary.
    statusDistribution: { hadir, telat, izin: izinFromRecap, alpha },
    departmentAttendance,
    recentAttendance,
  };
}
