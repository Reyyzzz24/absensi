import { redirect } from "next/navigation";
import { getAdminSession } from "@/lib/session";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { CompanySettings } from "@absensi-next/contracts";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { SettingsBody } from "@/components/settings/SettingsBody";

const CONFIG_LINKS = [
  { label: "Lokasi / Geofence", description: "Radius check-in, koordinat kantor.", href: "/admin/config/office-locations" },
  { label: "Jadwal & Shift", description: "Jam kerja, ambang telat per shift.", href: "/admin/config/shifts" },
  { label: "Dinas Luar", description: "Penugasan lapangan yang mengecualikan geofence.", href: "/admin/config/field-assignments" },
  { label: "Hari Libur", description: "Hari kerja, sinkronisasi libur nasional, libur manual perusahaan.", href: "/admin/config/holidays" },
];

// Auth check for <AdminShell> now lives in the route group's layout.tsx --
// this page still needs its own session read for session.role below.
export default async function AdminSettingsPage() {
  const session = await getAdminSession();
  if (!session) redirect("/admin/login");

  let company: CompanySettings | undefined;
  try {
    company = await adminApi.get<CompanySettings>("/api/v1/admin/config/company");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
    // Non-401 failure: settings page still renders, just without the
    // company card, rather than taking the whole page down.
  }

  return (
    <div className="mx-auto max-w-lg space-y-6">
      <PageHeader title="Pengaturan" description="Akun, notifikasi, dan profil perusahaan." />
      {/* Company profile edit is superadmin-only server-side (D-7, same
          tier as office-locations/shifts) -- a plain admin still sees
          everything else, just not that card, rather than a card that
          errors on save. */}
      <SettingsBody
        isAdmin
        company={session.role === "superadmin" ? company : undefined}
        configLinks={CONFIG_LINKS}
      />
    </div>
  );
}
