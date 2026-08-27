import { redirect } from "next/navigation";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { FieldAssignment } from "@/lib/types";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { FieldAssignmentForm } from "@/components/FieldAssignmentForm";

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" });
}

// Auth check + <AdminShell> now live in the route group's layout.tsx.
export default async function FieldAssignmentsPage() {
  let assignments: FieldAssignment[] = [];
  try {
    assignments = await adminApi.get<FieldAssignment[]>("/api/v1/admin/config/field-assignments");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Dinas Luar"
        description="Penugasan dinas luar (D-21) melewati pengecekan radius kantor saat karyawan check-in di tanggal yang ditugaskan -- ini pengganti mekanisme sebelumnya yang tidak ada sama sekali."
      />

      <ChartCard title="Tambah penugasan">
        <FieldAssignmentForm />
      </ChartCard>

      <ChartCard title="Daftar penugasan">
        {assignments.length === 0 && (
          <p className="text-sm text-muted-foreground">Belum ada penugasan dinas luar.</p>
        )}
        <ul className="divide-y divide-border text-sm">
          {assignments.map((a) => (
            <li key={a.id} className="py-2">
              <p className="font-medium text-foreground">
                Karyawan #{a.employee_id} &middot; {formatDate(a.work_date)}
              </p>
              {a.note && <p className="text-muted-foreground">{a.note}</p>}
            </li>
          ))}
        </ul>
      </ChartCard>
    </div>
  );
}
