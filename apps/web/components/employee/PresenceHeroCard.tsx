"use client";

import Link from "next/link";
import { Clock3, MapPinned } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { LiveClock } from "@/components/employee/LiveClock";
import { LocationStatus } from "@/components/employee/LocationStatus";
import type { Attendance } from "@/lib/types";

function formatTime(iso?: string) {
  if (!iso) return "--:--";
  return new Date(iso).toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit", timeZone: "Asia/Jakarta" });
}

function formatDuration(startIso?: string, endIso?: string) {
  if (!startIso || !endIso) return "-";
  const ms = new Date(endIso).getTime() - new Date(startIso).getTime();
  if (ms <= 0) return "-";
  const totalMinutes = Math.floor(ms / 60000);
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  return `${hours}j ${minutes}m`;
}

export function PresenceHeroCard({ today }: { today: Attendance | null }) {
  const isOpen = today && (today.status === "open" || today.status === "flagged_no_checkout");
  const isClosed = today && today.status === "closed";

  return (
    <Card className="gap-5 rounded-2xl border border-border p-6 shadow-sm sm:p-7">
      <div className="flex flex-col justify-between gap-5 sm:flex-row sm:items-start">
        <LiveClock />

        <div className="flex flex-col items-start gap-2 sm:items-end">
          {/* Dedicated presence badge (not the day-status StatusBadge used
              elsewhere) -- its fixed vocabulary is "Hadir"/"Izin"/"Alpha"
              etc. for a full day's classification, which doesn't fit
              "Sedang bekerja"/"Sudah check-out" right now/today framing. */}
          <span
            className={cn(
              "inline-flex items-center rounded-full px-2.5 py-1 text-[13px] font-medium",
              !today && "bg-secondary text-muted-foreground",
              isOpen && "bg-status-hadir-bg text-status-hadir",
              isClosed && "bg-status-izin-bg text-status-izin",
            )}
          >
            {!today ? "Belum check-in" : isOpen ? "Sedang bekerja" : "Sudah check-out"}
          </span>
        </div>
      </div>

      {/* TODO: mock -- no endpoint exposes the employee's assigned shift for
          today ahead of/independent of an attendance row (only shift_id is
          stored once checked in, never the shift's name/times). */}
      <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
        <Clock3 className="size-4 shrink-0" />
        Shift Pagi &middot; 08:00–17:00
        <span className="text-muted-foreground/60">(contoh)</span>
      </div>

      <div className="flex flex-wrap items-center gap-x-6 gap-y-3 border-t border-border pt-5">
        <div>
          <p className="text-xs text-muted-foreground">Jam masuk</p>
          <p className="text-lg font-semibold text-foreground">
            {formatTime(today?.check_in_at)}
            {today?.is_late && <span className="ml-1.5 text-sm font-normal text-orange-600">(terlambat)</span>}
          </p>
        </div>
        <div>
          <p className="text-xs text-muted-foreground">Jam keluar</p>
          <p className="text-lg font-semibold text-foreground">
            {formatTime(today?.check_out_at)}
            {today?.is_early_leave && <span className="ml-1.5 text-sm font-normal text-orange-600">(pulang cepat)</span>}
          </p>
        </div>
        {isClosed && (
          <div>
            <p className="text-xs text-muted-foreground">Total jam kerja</p>
            <p className="text-lg font-semibold text-foreground">
              {formatDuration(today?.check_in_at, today?.check_out_at)}
            </p>
          </div>
        )}
        {today && (
          <div className="ml-auto flex items-center gap-1.5 text-xs text-muted-foreground">
            <MapPinned className="size-3.5" />
            {today.is_wfh ? "WFH" : "Kantor"}
          </div>
        )}
      </div>

      <div className="flex flex-col gap-3 border-t border-border pt-5 sm:flex-row sm:items-center sm:justify-between">
        <LocationStatus />

        {!isClosed ? (
          <Button
            render={<Link href="/checkin" />}
            nativeButton={false}
            size="lg"
            className="w-full sm:w-auto"
          >
            {isOpen ? "Check-out sekarang" : "Check-in sekarang"}
          </Button>
        ) : (
          <p className="text-sm font-medium text-status-hadir">Presensi hari ini selesai ✓</p>
        )}
      </div>
    </Card>
  );
}
