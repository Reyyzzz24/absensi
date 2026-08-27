import { redirect } from "next/navigation";
import { getAdminSession } from "@/lib/session";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { OfficeLocation } from "@/lib/types";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { OfficeLocationForm } from "@/components/OfficeLocationForm";
import { OfficeLocationEditForm } from "@/components/OfficeLocationEditForm";

// Auth check for <AdminShell> now lives in the route group's layout.tsx --
// this page still needs its own session read for session.role below.
export default async function OfficeLocationsPage() {
  const session = await getAdminSession();
  if (!session) redirect("/admin/login");

  let locations: OfficeLocation[] = [];
  try {
    locations = await adminApi.get<OfficeLocation[]>("/api/v1/admin/config/office-locations");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
  }

  const isSuperadmin = session.role === "superadmin";

  return (
    <div className="space-y-6">
      <PageHeader
        title="Lokasi / Geofence"
        description="Radius di sini benar-benar ditegakkan saat karyawan check-in (D-1) -- berbeda dari sistem lama yang menghitung jarak tapi tidak pernah menolak berdasarkan itu."
      />

      {isSuperadmin ? (
        <ChartCard title="Tambah lokasi">
          <OfficeLocationForm />
        </ChartCard>
      ) : (
        <p className="rounded-2xl border border-orange-200 bg-orange-50 p-3 text-sm text-orange-700">
          Hanya superadmin yang bisa menambah/mengubah lokasi kantor.
        </p>
      )}

      {locations.length === 0 && (
        <ChartCard title="Daftar lokasi">
          <p className="text-sm text-muted-foreground">Belum ada lokasi.</p>
        </ChartCard>
      )}

      {/* One card per office location -- structured as a list from the
          start (not a single hardcoded location) so multi-location /
          "dinas luar" whitelisting later just means rendering more of
          these, no reshaping needed. */}
      {locations.map((loc) => (
        <ChartCard key={loc.id} title={isSuperadmin ? "Ubah lokasi" : loc.name}>
          {isSuperadmin ? (
            <OfficeLocationEditForm location={loc} />
          ) : (
            <div className="space-y-2 text-sm">
              <div className="flex items-center justify-between">
                <p className="font-medium text-foreground">{loc.name}</p>
                <span className={loc.is_active ? "text-xs text-status-hadir" : "text-xs text-muted-foreground"}>
                  {loc.is_active ? "Aktif" : "Nonaktif"}
                </span>
              </div>
              <p className="text-muted-foreground">
                {loc.latitude}, {loc.longitude} &middot; radius {loc.radius_meters}m
              </p>
            </div>
          )}
        </ChartCard>
      ))}
    </div>
  );
}
