import { redirect } from "next/navigation";
import Link from "next/link";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { LeaveRequest, LeaveStatus } from "@/lib/types";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { LeaveReviewButtons } from "@/components/LeaveReviewButtons";
import { cn } from "@/lib/utils";

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

// Auth check + <AdminShell> now live in the route group's layout.tsx.
export default async function AdminLeaveRequestsPage({
  searchParams,
}: {
  searchParams: Promise<{ status?: string }>;
}) {
  const { status } = await searchParams;
  const query = status ? `?status=${encodeURIComponent(status)}` : "";

  let requests: LeaveRequest[] = [];
  try {
    requests = await adminApi.get<LeaveRequest[]>(`/api/v1/admin/leave-requests${query}`);
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
  }

  const filters: { label: string; value: string }[] = [
    { label: "Semua", value: "" },
    { label: "Menunggu", value: "pending" },
    { label: "Disetujui", value: "approved" },
    { label: "Ditolak", value: "rejected" },
  ];

  return (
    <div className="space-y-6">
      <PageHeader title="Cuti & Izin" description="Tinjau dan proses pengajuan izin/sakit karyawan." />

      <div className="flex gap-2 text-xs">
        {filters.map((f) => (
          <Link
            key={f.value}
            href={f.value ? `/admin/leave-requests?status=${f.value}` : "/admin/leave-requests"}
            className={cn(
              "rounded-full px-3 py-1 font-medium transition",
              (status ?? "") === f.value
                ? "bg-primary text-primary-foreground"
                : "bg-secondary text-muted-foreground hover:bg-secondary/80",
            )}
          >
            {f.label}
          </Link>
        ))}
      </div>

      <ChartCard title="Daftar pengajuan">
        {requests.length === 0 && <p className="text-sm text-muted-foreground">Tidak ada pengajuan.</p>}
        <ul className="divide-y divide-border text-sm">
          {requests.map((r) => (
            <li key={r.id} className="flex items-center justify-between py-3">
              <div>
                <p className="font-medium text-foreground capitalize">
                  {r.type} &middot;{" "}
                  {r.employee ? `${r.employee.full_name} (${r.employee.nik})` : `Karyawan #${r.employee_id}`}
                </p>
                <p className="text-muted-foreground">
                  {formatDate(r.start_date)} &ndash; {formatDate(r.end_date)}
                </p>
                {r.reason && <p className="mt-1 text-foreground">{r.reason}</p>}
              </div>
              {r.status === "pending" ? (
                <LeaveReviewButtons id={r.id} />
              ) : (
                <span className={cn("text-xs font-medium", STATUS_COLOR[r.status])}>
                  {STATUS_LABEL[r.status]}
                </span>
              )}
            </li>
          ))}
        </ul>
      </ChartCard>
    </div>
  );
}
