import { redirect } from "next/navigation";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { Attendance } from "@/lib/types";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { Button } from "@/components/ui/button";

function todayISO() {
  return new Date().toISOString().slice(0, 10);
}

function formatTime(iso?: string) {
  if (!iso) return "-";
  return new Date(iso).toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit", timeZone: "Asia/Jakarta" });
}

function mapsLink(lat?: number, lng?: number) {
  if (lat === undefined || lng === undefined) return null;
  return `https://www.google.com/maps?q=${lat},${lng}`;
}

// Auth check + <AdminShell> now live in the route group's layout.tsx.
export default async function MonitoringPage({
  searchParams,
}: {
  searchParams: Promise<{ date?: string }>;
}) {
  const { date } = await searchParams;
  const activeDate = date || todayISO();

  let records: Attendance[] = [];
  try {
    records = await adminApi.get<Attendance[]>(`/api/v1/admin/attendance/monitoring?date=${activeDate}`);
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
  }

  return (
    <div className="space-y-6">
      <PageHeader title="Monitoring Presensi" description="Lihat lokasi dan foto presensi karyawan per tanggal." />

      {/* Plain GET form -- no client JS needed to change the date. */}
      <form method="get" className="flex items-center gap-2 text-sm">
        <input
          type="date"
          name="date"
          defaultValue={activeDate}
          className="h-9 rounded-lg border border-border bg-white px-3 text-sm shadow-sm focus:ring-2 focus:ring-ring focus:outline-none"
        />
        <Button type="submit" size="sm">
          Tampilkan
        </Button>
      </form>

      <ChartCard title={`Presensi ${activeDate}`}>
          {records.length === 0 && <p className="text-sm text-muted-foreground">Tidak ada presensi pada tanggal ini.</p>}
          <ul className="divide-y divide-border text-sm">
            {records.map((r) => {
              const inMapsLink = mapsLink(r.check_in_lat, r.check_in_lng);
              const outMapsLink = mapsLink(r.check_out_lat, r.check_out_lng);
              return (
                <li key={r.id} className="py-3">
                  <div className="flex items-center justify-between">
                    <p className="font-medium text-foreground">
                      {r.employee ? `${r.employee.full_name} (${r.employee.nik})` : `Karyawan #${r.employee_id}`}
                    </p>
                    <span className="text-xs text-muted-foreground">
                      {r.is_wfh ? "WFH" : "Kantor"} &middot; {r.status}
                    </span>
                  </div>

                  <div className="mt-2 grid grid-cols-2 gap-4">
                    <div>
                      <p className="text-xs font-medium text-muted-foreground">
                        Masuk {formatTime(r.check_in_at)}
                        {r.is_late && <span className="ml-1 text-orange-600">(terlambat)</span>}
                        {r.check_in_distance_m !== undefined && ` · ${r.check_in_distance_m}m dari kantor`}
                      </p>
                      <div className="mt-1 flex items-center gap-3">
                        {/* Same-origin proxy route -- browser sends the admin's
                            httpOnly cookie automatically; the Go API itself
                            never sees a request without a Bearer token. */}
                        {r.check_in_photo_path && (
                          // eslint-disable-next-line @next/next/no-img-element
                          <img
                            src={`/api/admin/attendance/${r.id}/photo?type=in`}
                            alt="Foto check-in"
                            className="h-16 w-16 rounded-md border border-border object-cover"
                          />
                        )}
                        {inMapsLink && (
                          <a href={inMapsLink} target="_blank" rel="noreferrer" className="text-xs text-primary underline">
                            Lihat peta
                          </a>
                        )}
                      </div>
                    </div>

                    <div>
                      <p className="text-xs font-medium text-muted-foreground">
                        Keluar {formatTime(r.check_out_at)}
                        {r.is_early_leave && <span className="ml-1 text-orange-600">(pulang cepat)</span>}
                        {r.check_out_distance_m !== undefined && ` · ${r.check_out_distance_m}m dari kantor`}
                      </p>
                      {r.check_out_at && (
                        <div className="mt-1 flex items-center gap-3">
                          {r.check_out_photo_path && (
                            // eslint-disable-next-line @next/next/no-img-element
                            <img
                              src={`/api/admin/attendance/${r.id}/photo?type=out`}
                              alt="Foto check-out"
                              className="h-16 w-16 rounded-md border border-border object-cover"
                            />
                          )}
                          {outMapsLink && (
                            <a href={outMapsLink} target="_blank" rel="noreferrer" className="text-xs text-primary underline">
                              Lihat peta
                            </a>
                          )}
                        </div>
                      )}
                    </div>
                  </div>
                </li>
              );
            })}
          </ul>
        </ChartCard>
    </div>
  );
}
