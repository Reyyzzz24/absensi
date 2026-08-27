import { redirect } from "next/navigation";
import { getEmployeeSession } from "@/lib/session";
import { employeeApi, ApiError } from "@/lib/authedApi";
import type { LeaveRequest, LeaveStatus } from "@/lib/types";
import { LeaveRequestForm } from "@/components/LeaveRequestForm";

const STATUS_LABEL: Record<LeaveStatus, string> = {
  pending: "Menunggu",
  approved: "Disetujui",
  rejected: "Ditolak",
};

const STATUS_COLOR: Record<LeaveStatus, string> = {
  pending: "text-orange-600",
  approved: "text-status-hadir",
  rejected: "text-status-alpha",
};

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" });
}

export default async function LeavePage() {
  const session = await getEmployeeSession();
  if (!session) redirect("/");

  let requests: LeaveRequest[] = [];
  try {
    requests = await employeeApi.get<LeaveRequest[]>("/api/v1/leave-requests");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/");
  }

  return (
    <div className="mx-auto max-w-lg space-y-4">
      <h1 className="text-xl font-semibold text-foreground">Izin / Sakit</h1>

      <div className="rounded-2xl border border-border bg-white p-4 shadow-sm">
        <h2 className="text-sm font-medium text-foreground">Ajukan izin/sakit baru</h2>
        <p className="mt-1 text-xs text-muted-foreground">
          Pengajuan menunggu persetujuan admin sebelum berlaku.
        </p>
        <div className="mt-3">
          <LeaveRequestForm />
        </div>
      </div>

      <div className="rounded-2xl border border-border bg-white p-4 shadow-sm">
        <h2 className="text-sm font-medium text-foreground">Riwayat pengajuan</h2>
        {requests.length === 0 && (
          <p className="mt-2 text-sm text-muted-foreground">Belum ada pengajuan.</p>
        )}
        <ul className="mt-3 divide-y divide-border text-sm">
          {requests.map((r) => (
            <li key={r.id} className="flex items-center justify-between py-2">
              <div>
                <p className="font-medium text-foreground capitalize">{r.type}</p>
                <p className="text-muted-foreground">
                  {formatDate(r.start_date)} &ndash; {formatDate(r.end_date)}
                </p>
              </div>
              <span className={`text-xs font-medium ${STATUS_COLOR[r.status]}`}>
                {STATUS_LABEL[r.status]}
              </span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
