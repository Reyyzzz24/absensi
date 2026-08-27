import { redirect } from "next/navigation";
import { Search } from "lucide-react";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { Department, Employee } from "@/lib/types";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { EmployeeForm } from "@/components/EmployeeForm";
import { EmployeeEditRow } from "@/components/EmployeeEditRow";

// Auth check + <AdminShell> now live in the route group's layout.tsx.
export default async function EmployeesPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string }>;
}) {
  const { q } = await searchParams;
  const query = q?.trim() ?? "";

  let employees: Employee[] = [];
  let departments: Department[] = [];
  try {
    [employees, departments] = await Promise.all([
      adminApi.get<Employee[]>(
        `/api/v1/admin/employees${query ? `?q=${encodeURIComponent(query)}` : ""}`,
      ),
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
        {/* Plain GET form, no client JS -- same pattern as Laporan's
            month/year filter. Also the landing target for the topbar's
            global search (Topbar.tsx submits here with the same `q` param),
            so refining the search here keeps the exact same URL shape. */}
        <form method="get" className="relative mb-4 max-w-sm">
          <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="search"
            name="q"
            defaultValue={query}
            placeholder="Cari nama atau NIK..."
            className="h-9 w-full rounded-lg border border-border bg-white pl-9 pr-3 text-sm text-foreground shadow-sm placeholder:text-muted-foreground focus:ring-2 focus:ring-ring focus:outline-none"
          />
        </form>

        {employees.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            {query ? `Tidak ada karyawan yang cocok dengan "${query}".` : "Belum ada karyawan."}
          </p>
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
