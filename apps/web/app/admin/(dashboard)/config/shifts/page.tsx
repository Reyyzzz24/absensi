import { redirect } from "next/navigation";
import { getAdminSession } from "@/lib/session";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { Shift } from "@/lib/types";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { ShiftForm } from "@/components/ShiftForm";

// Auth check for <AdminShell> now lives in the route group's layout.tsx --
// this page still needs its own session read for session.role below.
export default async function ShiftsPage() {
  const session = await getAdminSession();
  if (!session) redirect("/admin/login");

  let shifts: Shift[] = [];
  try {
    shifts = await adminApi.get<Shift[]>("/api/v1/admin/config/shifts");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
  }

  const isSuperadmin = session.role === "superadmin";

  return (
    <div className="space-y-6">
      <PageHeader
        title="Jadwal & Shift"
        description="Toleransi keterlambatan (grace period) diatur per shift di sini (D-10) -- menggantikan ambang 09:15/09:00 yang dulu hardcoded dan tidak konsisten di sistem lama."
      />

      {isSuperadmin ? (
        <ChartCard title="Tambah shift">
          <ShiftForm />
        </ChartCard>
      ) : (
        <p className="rounded-2xl border border-orange-200 bg-orange-50 p-3 text-sm text-orange-700">
          Hanya superadmin yang bisa menambah/mengubah shift.
        </p>
      )}

      <ChartCard title="Daftar shift">
        {shifts.length === 0 && <p className="text-sm text-muted-foreground">Belum ada shift.</p>}
        <ul className="divide-y divide-border text-sm">
          {shifts.map((s) => (
            <li key={s.id} className="py-2">
              <div className="flex items-center justify-between">
                <p className="font-medium text-foreground">
                  {s.code} &middot; {s.name}
                </p>
                {s.is_overnight && (
                  <span className="text-xs font-medium text-[#8B5CF6]">Lintas tengah malam</span>
                )}
              </div>
              <p className="text-muted-foreground">
                {s.is_day_off
                  ? "Hari libur"
                  : `${s.start_time?.slice(0, 5)} – ${s.end_time?.slice(0, 5)} · toleransi ${s.late_grace_minutes} menit`}
              </p>
            </li>
          ))}
        </ul>
      </ChartCard>
    </div>
  );
}
