import { redirect } from "next/navigation";
import { getAdminSession } from "@/lib/session";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { CompanySettings, NationalHoliday, CompanyHoliday } from "@absensi-next/contracts";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { WorkingWeekdaysForm } from "@/components/holidays/WorkingWeekdaysForm";
import { NationalHolidaySyncPanel } from "@/components/holidays/NationalHolidaySyncPanel";
import { CompanyHolidayForm } from "@/components/holidays/CompanyHolidayForm";
import { CompanyHolidayList } from "@/components/holidays/CompanyHolidayList";

// Auth check + <AdminShell> now live in the route group's layout.tsx.
export default async function HolidaysPage() {
  const session = await getAdminSession();
  if (!session) redirect("/admin/login");
  const isSuperadmin = session.role === "superadmin";

  const year = new Date().getFullYear();

  let company: CompanySettings | null = null;
  let nationalHolidays: NationalHoliday[] = [];
  let companyHolidays: CompanyHoliday[] = [];
  try {
    [company, nationalHolidays, companyHolidays] = await Promise.all([
      adminApi.get<CompanySettings>("/api/v1/admin/config/company"),
      adminApi.get<NationalHoliday[]>(`/api/v1/admin/holidays/national?year=${year}`),
      adminApi.get<CompanyHoliday[]>("/api/v1/admin/holidays/company"),
    ]);
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Hari Libur"
        description="Tiga sumber digabung satu resolver: akhir pekan (bisa diatur), libur nasional (disinkronkan, termasuk cuti bersama), dan libur manual perusahaan. Tidak ada baris absensi yang dibuat untuk hari libur -- status dihitung saat laporan dibuka, jadi perubahan kebijakan langsung berlaku ke data lama tanpa perlu dibersihkan."
      />

      <ChartCard title="Hari Kerja Perusahaan" description="Menentukan hari mana yang dianggap akhir pekan.">
        {company && isSuperadmin ? (
          <WorkingWeekdaysForm company={company} />
        ) : (
          <p className="text-sm text-muted-foreground">Hanya superadmin yang dapat mengubah pengaturan ini.</p>
        )}
      </ChartCard>

      <ChartCard title={`Libur Nasional ${year}`} description="Disinkronkan dari kalender publik, termasuk cuti bersama.">
        {isSuperadmin ? (
          <NationalHolidaySyncPanel year={year} holidays={nationalHolidays} />
        ) : nationalHolidays.length === 0 ? (
          <p className="text-sm text-muted-foreground">Belum ada data untuk {year}.</p>
        ) : (
          <ul className="divide-y divide-border text-sm">
            {nationalHolidays.map((h) => (
              <li key={h.id} className="py-2">
                {h.name} &middot; {new Date(h.holiday_date).toLocaleDateString("id-ID")}
              </li>
            ))}
          </ul>
        )}
      </ChartCard>

      {isSuperadmin && (
        <ChartCard title="Tambah Libur Perusahaan">
          <CompanyHolidayForm />
        </ChartCard>
      )}

      <ChartCard title="Daftar Libur Perusahaan" description="Tanggal tunggal atau rentang, dikelola manual oleh admin.">
        {isSuperadmin ? (
          <CompanyHolidayList holidays={companyHolidays} />
        ) : companyHolidays.length === 0 ? (
          <p className="text-sm text-muted-foreground">Belum ada libur perusahaan manual.</p>
        ) : (
          <ul className="divide-y divide-border text-sm">
            {companyHolidays.map((h) => (
              <li key={h.id} className="py-2">
                {h.name} &middot; {h.start_date} – {h.end_date}
              </li>
            ))}
          </ul>
        )}
      </ChartCard>
    </div>
  );
}
