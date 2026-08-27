import { Suspense } from "react";
import Link from "next/link";
import { redirect } from "next/navigation";
import { Download, Users, ClipboardCheck, AlarmClock, CalendarClock } from "lucide-react";
import { getAdminSession, type SessionClaims } from "@/lib/session";
import { ApiError } from "@/lib/authedApi";
import { loadDashboardData } from "@/lib/dashboard-data";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { StatCard } from "@/components/dashboard/StatCard";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { AttendanceTrendChart } from "@/components/dashboard/AttendanceTrendChart";
import { StatusDonutChart, type StatusDonutDatum } from "@/components/dashboard/StatusDonutChart";
import { DepartmentBarChart } from "@/components/dashboard/DepartmentBarChart";
import { RecentAttendanceTable } from "@/components/dashboard/RecentAttendanceTable";
import { DashboardSkeleton } from "@/components/dashboard/DashboardSkeleton";
import { Button } from "@/components/ui/button";

// <AdminShell> now lives in the route group's layout.tsx (perf fix -- see
// that file). This page's own session check stays (session.role is used in
// the greeting below), still outside the Suspense boundary for the same
// instant-307-vs-soft-redirect reason as before -- that property is
// per-page, not affected by where the shell itself is mounted.
export default async function AdminDashboardPage() {
  const session = await getAdminSession();
  if (!session) redirect("/admin/login");

  return (
    <Suspense fallback={<DashboardSkeleton />}>
      <DashboardBody session={session} />
    </Suspense>
  );
}

async function DashboardBody({ session }: { session: SessionClaims }) {
  let data: Awaited<ReturnType<typeof loadDashboardData>> | null = null;
  try {
    data = await loadDashboardData();
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
  }

  const kpis = data?.kpis;
  const distribution = data?.statusDistribution;

  const donutData: StatusDonutDatum[] = distribution
    ? [
        { name: "Hadir", value: distribution.hadir, color: "#10B981" },
        { name: "Telat", value: distribution.telat, color: "#F59E0B" },
        { name: "Izin", value: distribution.izin, color: "#3B82F6" },
        { name: "Alpha", value: distribution.alpha, color: "#DC2626" },
      ]
    : [];
  const donutTotal = donutData.reduce((s, d) => s + d.value, 0);

  return (
    <div className="space-y-6">
      <PageHeader
        title={`Hai, ${session.role ? session.role.charAt(0).toUpperCase() + session.role.slice(1) : "Admin"} 👋`}
        description={new Date().toLocaleDateString("id-ID", {
          weekday: "long",
          day: "numeric",
          month: "long",
          year: "numeric",
        })}
        actions={
          <>
            {/* TODO: mock -- period filter is presentational only, dashboard
                always shows "bulan ini" / hari ini data for now. */}
            <Button variant="outline" size="sm">
              Bulan ini
            </Button>
            {/* TODO: mock -- no export endpoint exists yet. */}
            <Button size="sm" disabled title="Belum tersedia">
              <Download />
              Export
            </Button>
          </>
        }
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Total Karyawan" value={kpis?.totalKaryawan ?? 0} icon={Users} tone="blue" />
        <StatCard
          label="Hadir Hari Ini"
          value={kpis?.hadirHariIni ?? 0}
          icon={ClipboardCheck}
          tone="teal"
          trend={kpis?.hadirTrend ?? undefined}
        />
        <StatCard
          label="Terlambat Hari Ini"
          value={kpis?.terlambatHariIni ?? 0}
          icon={AlarmClock}
          tone="orange"
          trend={kpis?.terlambatTrend ?? undefined}
        />
        <StatCard
          label="Izin/Cuti Hari Ini"
          value={kpis?.izinHariIni ?? 0}
          icon={CalendarClock}
          tone="violet"
          trend={kpis?.izinTrend ?? undefined}
        />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <ChartCard
          title="Tren Kehadiran"
          description="Rata-rata kehadiran karyawan per bulan (data contoh)"
          className="lg:col-span-2"
        >
          <AttendanceTrendChart />
        </ChartCard>

        <ChartCard title="Distribusi Status Hari Ini">
          {donutTotal > 0 ? (
            <StatusDonutChart data={donutData} />
          ) : (
            <p className="flex h-[180px] items-center justify-center text-sm text-muted-foreground">
              Belum ada data presensi hari ini.
            </p>
          )}
        </ChartCard>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <ChartCard title="Kehadiran per Departemen" description="Top departemen hari ini" className="lg:col-span-1">
          {data && data.departmentAttendance.length > 0 ? (
            <DepartmentBarChart data={data.departmentAttendance} />
          ) : (
            <p className="flex h-[160px] items-center justify-center text-sm text-muted-foreground">
              Belum ada data.
            </p>
          )}
        </ChartCard>

        <ChartCard
          title="Absensi Terbaru"
          className="lg:col-span-2"
          filter={
            <Link href="/admin/monitoring" className="text-sm font-medium text-primary hover:underline">
              Lihat semua
            </Link>
          }
        >
          <RecentAttendanceTable data={data?.recentAttendance ?? []} />
        </ChartCard>
      </div>
    </div>
  );
}
