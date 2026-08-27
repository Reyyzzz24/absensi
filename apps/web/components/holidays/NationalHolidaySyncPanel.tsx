"use client";

import { useState } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { NationalHoliday } from "@absensi-next/contracts";

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric", weekday: "short" });
}

// Sync is admin-triggered only -- this button is the sole place a request
// to the external holiday calendar source happens (D-25), never on any
// read path. router.refresh() afterwards re-fetches the server-rendered
// list below from the now-updated cache.
export function NationalHolidaySyncPanel({ year, holidays }: { year: number; holidays: NationalHoliday[] }) {
  const [syncing, setSyncing] = useState(false);
  const router = useRouter();

  async function handleSync() {
    setSyncing(true);
    try {
      const res = await fetch(`/api/admin/holidays/national/sync?year=${year}`, { method: "POST" });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error ?? "Gagal sinkronisasi");
      toast.success(`${data.synced ?? 0} hari libur nasional ${year} disinkronkan`);
      router.refresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Gagal sinkronisasi");
    } finally {
      setSyncing(false);
    }
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm text-muted-foreground">
          Disinkronkan dari kalender publik (termasuk cuti bersama), lalu disimpan lokal -- sistem tidak pernah
          memanggil sumber ini saat memuat laporan/absensi. Entri yang sudah diedit manual tidak akan tertimpa.
        </p>
        <Button variant="outline" size="sm" onClick={handleSync} disabled={syncing} className="shrink-0">
          <RefreshCw className={syncing ? "animate-spin" : ""} />
          {syncing ? "Sinkronisasi…" : `Sinkronkan ${year}`}
        </Button>
      </div>

      {holidays.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          Belum ada data untuk {year}. Klik &quot;Sinkronkan&quot; untuk mengambil dari kalender nasional.
        </p>
      ) : (
        <ul className="divide-y divide-border text-sm">
          {holidays.map((h) => (
            <li key={h.id} className="flex items-center justify-between gap-3 py-2">
              <div>
                <p className="font-medium text-foreground">{h.name}</p>
                <p className="text-xs text-muted-foreground">{formatDate(h.holiday_date)}</p>
              </div>
              <div className="flex items-center gap-2">
                {h.is_cuti_bersama && (
                  <span className="rounded-full bg-secondary px-2 py-0.5 text-xs font-medium text-muted-foreground">
                    Cuti Bersama
                  </span>
                )}
                {h.source === "manual" && (
                  <span className="rounded-full bg-accent px-2 py-0.5 text-xs font-medium text-primary">
                    Dikoreksi admin
                  </span>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
