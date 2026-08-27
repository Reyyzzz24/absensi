import { redirect } from "next/navigation";
import { getEmployeeSession } from "@/lib/session";
import { employeeApi, ApiError } from "@/lib/authedApi";
import type { OfficeLocation } from "@/lib/types";
import { CheckInForm } from "@/components/CheckInForm";

export default async function CheckInPage() {
  const session = await getEmployeeSession();
  if (!session) redirect("/");

  // Fetched server-side (same pattern as every other employee page) and
  // passed down as a prop -- avoids needing a separate client-side proxy
  // route just to read one GET endpoint.
  let geofences: OfficeLocation[] = [];
  try {
    geofences = await employeeApi.get<OfficeLocation[]>("/api/v1/attendance/geofence");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/");
    // Any other failure (e.g. API hiccup): fall back to an empty list --
    // CheckInForm already handles "no geofence data" as its own state
    // (map/badge hidden, check-in itself is unaffected since the real
    // enforcement is server-side anyway).
  }

  return (
    <div className="mx-auto max-w-lg space-y-4">
      <div>
        <h1 className="text-xl font-semibold text-foreground">Check-in</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Ambil foto dan izinkan akses lokasi untuk mencatat presensi. Sistem otomatis
          menentukan apakah ini check-in atau check-out berdasarkan status Anda hari ini.
        </p>
      </div>

      <div className="rounded-2xl border border-border bg-white p-4 shadow-sm">
        <CheckInForm geofences={geofences} />
      </div>
    </div>
  );
}
