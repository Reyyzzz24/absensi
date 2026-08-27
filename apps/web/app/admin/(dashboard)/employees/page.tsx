import { redirect } from "next/navigation";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { Department, Employee } from "@/lib/types";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { EmployeeForm } from "@/components/EmployeeForm";
import { EmployeeEditRow } from "@/components/EmployeeEditRow";

// Auth check + <AdminShell> now live in the route group's layout.tsx.
export default async function EmployeesPage() {
  let employees: Employee[] = [];
  let departments: Department[] = [];
  try {
    [employees, departments] = await Promise.all([
      adminApi.get<Employee[]>("/api/v1/admin/employees"),
      adminApi.get<Department[]>("/api/v1/admin/departments"),
    ]);
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
  }

  return (
    <div className="space-y-6">
      <PageHeader title="Karyawan" description="Kelola data karyawan dan penempatan departemen." />

      <ChartCard title="Tambah karyawan">
        <EmployeeForm departments={departments} />
      </ChartCard>

      <ChartCard title="Daftar karyawan">
        {employees.length === 0 ? (
          <p className="text-sm text-muted-foreground">Belum ada karyawan.</p>
        ) : (
          <ul className="divide-y divide-border text-sm">
            {employees.map((e) => (
              <EmployeeEditRow key={e.id} employee={e} departments={departments} />
            ))}
          </ul>
        )}
      </ChartCard>
    </div>
  );
}
