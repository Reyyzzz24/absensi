import { CalendarClock, CheckCircle2, Clock3, XCircle } from "lucide-react";
import { StatCard } from "@/components/dashboard/StatCard";

export function MonthlySummary({ izinCount }: { izinCount: number }) {
  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 sm:gap-4">
      {/* TODO: mock -- Hadir/Terlambat/Alpha need a per-employee monthly
          attendance history, which no employee-facing endpoint provides yet
          (only /attendance/today, a single day). */}
      <StatCard label="Hadir Bulan Ini" value="--" icon={CheckCircle2} tone="teal" />
      <StatCard label="Terlambat" value="--" icon={Clock3} tone="orange" />
      {/* Real: employee's own leave-requests this month, from the existing
          GET /leave-requests endpoint (already ownership-scoped). */}
      <StatCard label="Izin/Sakit" value={izinCount} icon={CalendarClock} tone="blue" />
      <StatCard label="Alpha" value="--" icon={XCircle} tone="violet" />
    </div>
  );
}
