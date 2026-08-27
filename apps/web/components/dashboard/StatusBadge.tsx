import { cn } from "@/lib/utils";

export type AttendanceStatusLabel = "hadir" | "telat" | "izin" | "sakit" | "alpha" | "libur";

const STYLES: Record<AttendanceStatusLabel, string> = {
  hadir: "bg-status-hadir-bg text-status-hadir",
  telat: "bg-status-telat-bg text-status-telat",
  izin: "bg-status-izin-bg text-status-izin",
  sakit: "bg-status-izin-bg text-status-izin",
  alpha: "bg-status-alpha-bg text-status-alpha",
  libur: "bg-secondary text-muted-foreground",
};

const LABELS: Record<AttendanceStatusLabel, string> = {
  hadir: "Hadir",
  telat: "Telat",
  izin: "Izin",
  sakit: "Sakit",
  alpha: "Alpha",
  libur: "Libur",
};

export function StatusBadge({ status, className }: { status: AttendanceStatusLabel; className?: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
        STYLES[status],
        className,
      )}
    >
      {LABELS[status]}
    </span>
  );
}
