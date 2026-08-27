import { redirect } from "next/navigation";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { Department } from "@/lib/types";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { DepartmentForm } from "@/components/DepartmentForm";

// Auth check + <AdminShell> now live in the route group's layout.tsx.
export default async function DepartmentsPage() {
  let departments: Department[] = [];
  try {
    departments = await adminApi.get<Department[]>("/api/v1/admin/departments");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
  }

  return (
    <div className="space-y-6">
      <PageHeader title="Departemen" description="Kelola daftar departemen perusahaan." />

      <ChartCard title="Tambah departemen">
        <DepartmentForm />
      </ChartCard>

      <ChartCard title="Daftar departemen">
        {departments.length === 0 ? (
          <p className="text-sm text-muted-foreground">Belum ada departemen.</p>
        ) : (
          <ul className="divide-y divide-border text-sm">
            {departments.map((d) => (
              <li key={d.id} className="py-2">
                <span className="font-mono text-muted-foreground">{d.code}</span> &middot; {d.name}
              </li>
            ))}
          </ul>
        )}
      </ChartCard>
    </div>
  );
}
