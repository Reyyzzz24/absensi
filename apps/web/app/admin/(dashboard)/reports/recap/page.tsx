import { redirect } from "next/navigation";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { MonthRecap, RecapDayStatus } from "@/lib/types";
import type { CompanySettings } from "@absensi-next/contracts";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { RecapTable } from "@/components/RecapTable";
import { Button } from "@/components/ui/button";
import { ExportRecapButton } from "@/components/ExportRecapButton";
import { PrintButton } from "@/components/PrintButton";

const STATUS_STYLE: Record<RecapDayStatus, string> = {
  hadir: "bg-status-hadir-bg text-status-hadir",
  izin: "bg-status-izin-bg text-status-izin",
  sakit: "bg-status-izin-bg text-status-izin",
  libur: "bg-secondary text-muted-foreground",
  alpha: "bg-status-alpha-bg text-status-alpha",
};

const STATUS_LETTER: Record<RecapDayStatus, string> = {
  hadir: "H",
  izin: "I",
  sakit: "S",
  libur: "L",
  alpha: "A",
};

function monthName(month: number) {
  return new Date(2000, month - 1, 1).toLocaleDateString("id-ID", { month: "long" });
}

// Auth check + <AdminShell> now live in the route group's layout.tsx.
export default async function RecapPage({
  searchParams,
}: {
  searchParams: Promise<{ year?: string; month?: string }>;
}) {
  const now = new Date();
  const { year: yearParam, month: monthParam } = await searchParams;
  const year = Number(yearParam) || now.getFullYear();
  const month = Number(monthParam) || now.getMonth() + 1;

  let recap: MonthRecap | null = null;
  try {
    recap = await adminApi.get<MonthRecap>(`/api/v1/admin/reports/recap?year=${year}&month=${month}`);
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
  }

  // Best-effort -- the print header still works with a generic title if this
  // fails (e.g. a plain admin without company-settings read access, though
  // currently that endpoint is admin-not-superadmin-only so this shouldn't
  // normally fail for anyone who can reach this page at all).
  let company: CompanySettings | null = null;
  try {
    company = await adminApi.get<CompanySettings>("/api/v1/admin/config/company");
  } catch {
    // ignored -- print header falls back to a generic title below
  }

  return (
    <div className="space-y-6">
      {/* Screen-only header (actions hidden on print via each button's own
          print:hidden). Print gets its own plain header further down instead
          of this one, so it isn't cluttered with buttons/description. */}
      <div className="print:hidden">
        <PageHeader
          title="Laporan"
          description="Rekap kehadiran bulanan per karyawan."
          actions={
            <>
              <PrintButton />
              <ExportRecapButton year={year} month={month} />
            </>
          }
        />
      </div>

      {/* Print-only header -- hidden on screen, shown only in @media print
          (see globals.css) via the `hidden print:block` pair below. Gives
          the printed page context (company + period) since the screen
          header above is deliberately hidden for print. */}
      <div className="hidden print:block">
        <p className="text-lg font-semibold">{company?.name ?? "Absensi"}</p>
        <p className="text-sm text-slate-600">
          Rekap Absensi Bulanan &middot; {monthName(month)} {year}
        </p>
      </div>

      {/* Plain GET form, no client JS -- consistent with /admin/monitoring. */}
      <form method="get" className="flex items-center gap-2 text-sm print:hidden">
        <select
          name="month"
          defaultValue={month}
          className="h-9 rounded-lg border border-border bg-white px-3 text-sm shadow-sm focus:ring-2 focus:ring-ring focus:outline-none"
        >
          {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
            <option key={m} value={m}>
              {monthName(m)}
            </option>
          ))}
        </select>
        <input
          type="number"
          name="year"
          defaultValue={year}
          className="h-9 w-24 rounded-lg border border-border bg-white px-3 text-sm shadow-sm focus:ring-2 focus:ring-ring focus:outline-none"
        />
        <Button type="submit" size="sm">
          Tampilkan
        </Button>
      </form>

      <div className="flex flex-wrap gap-3 text-xs">
        {(Object.keys(STATUS_LETTER) as RecapDayStatus[]).map((s) => (
          <span key={s} className="flex items-center gap-1">
            <span className={`inline-flex h-5 w-5 items-center justify-center rounded ${STATUS_STYLE[s]}`}>
              {STATUS_LETTER[s]}
            </span>
            <span className="capitalize text-muted-foreground">{s}</span>
          </span>
        ))}
        <span className="flex items-center gap-1">
          <span className="inline-flex h-5 w-5 items-center justify-center rounded bg-orange-100 text-orange-700">
            *
          </span>
          <span className="text-muted-foreground">Telat</span>
        </span>
      </div>

      {/* RecapTable owns its own overflow-x-auto/overflow-y scroll wrapper
          now -- no className override needed here. */}
      <ChartCard title="Rekap Bulanan">
        {!recap || recap.employees.length === 0 ? (
          <p className="text-sm text-muted-foreground">Tidak ada data untuk bulan ini.</p>
        ) : (
          <RecapTable recap={recap} />
        )}
      </ChartCard>
    </div>
  );
}
