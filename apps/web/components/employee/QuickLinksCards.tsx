import Link from "next/link";
import { ChevronRight } from "lucide-react";
import { IconTile } from "@/components/dashboard/IconTile";
import { CalendarClock, ListChecks } from "lucide-react";
import type { LeaveRequest } from "@/lib/types";

const LEAVE_STATUS_LABEL: Record<LeaveRequest["status"], string> = {
  pending: "Menunggu persetujuan",
  approved: "Disetujui",
  rejected: "Ditolak",
};

export function QuickLinksCards({
  latestLeave,
  activeTaskCount,
}: {
  latestLeave: LeaveRequest | null;
  activeTaskCount: number;
}) {
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <Link
        href="/leave"
        className="group flex items-center gap-3 rounded-2xl border border-border bg-white p-4 shadow-sm transition hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      >
        <IconTile icon={CalendarClock} tone="blue" />
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-foreground">Pengajuan izin/sakit</p>
          <p className="truncate text-xs text-muted-foreground">
            {latestLeave ? LEAVE_STATUS_LABEL[latestLeave.status] : "Belum ada pengajuan"}
          </p>
        </div>
        <ChevronRight className="size-4 shrink-0 text-muted-foreground transition group-hover:translate-x-0.5" />
      </Link>

      <Link
        href="/tasks"
        className="group flex items-center gap-3 rounded-2xl border border-border bg-white p-4 shadow-sm transition hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      >
        <IconTile icon={ListChecks} tone="violet" />
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-foreground">Task aktif</p>
          <p className="text-xs text-muted-foreground">{activeTaskCount} task berjalan</p>
        </div>
        <ChevronRight className="size-4 shrink-0 text-muted-foreground transition group-hover:translate-x-0.5" />
      </Link>
    </div>
  );
}
