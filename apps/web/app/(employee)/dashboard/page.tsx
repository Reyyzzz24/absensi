import { Suspense } from "react";
import { redirect } from "next/navigation";
import { employeeApi, ApiError } from "@/lib/authedApi";
import type { Attendance, LeaveRequest, Task } from "@/lib/types";
import { PresenceHeroCard } from "@/components/employee/PresenceHeroCard";
import { MonthlySummary } from "@/components/employee/MonthlySummary";
import { WeekStrip } from "@/components/employee/WeekStrip";
import { RecentActivityList } from "@/components/employee/RecentActivityList";
import { QuickLinksCards } from "@/components/employee/QuickLinksCards";
import { EmployeeDashboardSkeleton } from "@/components/employee/EmployeeDashboardSkeleton";
import Link from "next/link";

// The top-level auth-outside-Suspense check now lives in the route group's
// layout.tsx (perf fix -- see that file). This page only needs the
// data-fetch Suspense boundary, unaffected by that move since it was
// already isolated in <DashboardBody> below.
export default function EmployeeDashboardPage() {
  return (
    <Suspense fallback={<EmployeeDashboardSkeleton />}>
      <DashboardBody />
    </Suspense>
  );
}

async function DashboardBody() {
  let today: Attendance | null = null;
  let leaveRequests: LeaveRequest[] = [];
  let tasks: Task[] = [];
  try {
    [today, leaveRequests, tasks] = await Promise.all([
      employeeApi.get<Attendance | null>("/api/v1/attendance/today"),
      employeeApi.get<LeaveRequest[]>("/api/v1/leave-requests"),
      employeeApi.get<Task[]>("/api/v1/tasks"),
    ]);
  } catch (err) {
    // Stateless JWTs mean there's no server-side way to tell "expired" from
    // "invalid" here -- both are treated as "not logged in".
    if (err instanceof ApiError && err.status === 401) redirect("/");
  }

  const now = new Date();
  const thisMonthIzin = leaveRequests.filter((lr) => {
    const d = new Date(lr.start_date);
    return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth();
  }).length;

  const latestLeave = leaveRequests[0] ?? null; // API already orders by start_date DESC.
  // Every task defaults to (and today, can only ever be) "planned" -- there's
  // no "completed" status settable from the UI yet, so "active" == "all" for
  // now. Real count, not a mock -- just a currently-trivial one.
  const activeTaskCount = tasks.length;

  return (
    <div className="space-y-6">
      <PresenceHeroCard today={today} />

      <section className="space-y-3">
        <h2 className="text-sm font-semibold text-foreground">Ringkasan bulan ini</h2>
        <MonthlySummary izinCount={thisMonthIzin} />
      </section>

      <section className="space-y-3 rounded-2xl border border-border bg-white p-5 shadow-sm">
        <h2 className="text-sm font-semibold text-foreground">Minggu ini</h2>
        <WeekStrip />
      </section>

      <QuickLinksCards latestLeave={latestLeave} activeTaskCount={activeTaskCount} />

      <section className="rounded-2xl border border-border bg-white p-5 shadow-sm">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-foreground">Aktivitas terbaru</h2>
          {/* No dedicated attendance-history page exists yet (it would need
              the not-yet-implemented GET /attendance/history endpoint) --
              disabled rather than linking to a page that doesn't exist or
              to somewhere that pretends to be one. */}
          <span className="text-sm font-medium text-muted-foreground/50" title="Belum tersedia">
            Lihat semua
          </span>
        </div>
        <div className="mt-2">
          <RecentActivityList />
        </div>
      </section>
    </div>
  );
}
