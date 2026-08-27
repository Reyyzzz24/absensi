"use client";

import { cn } from "@/lib/utils";

const DAY_LABELS = ["Sen", "Sel", "Rab", "Kam", "Jum", "Sab", "Min"];

const DOT_COLOR: Record<string, string> = {
  hadir: "bg-status-hadir",
  telat: "bg-status-telat",
  izin: "bg-status-izin",
  alpha: "bg-status-alpha",
  future: "bg-border",
};

// TODO: mock -- there is no employee-facing endpoint for a date-ranged
// attendance history (only /attendance/today, one day at a time). Pattern
// below is illustrative only; days after "today" are rendered as empty/
// future regardless of the mock data.
const MOCK_WEEK: { status: keyof typeof DOT_COLOR }[] = [
  { status: "hadir" },
  { status: "hadir" },
  { status: "telat" },
  { status: "hadir" },
  { status: "izin" },
  { status: "future" },
  { status: "future" },
];

export function WeekStrip() {
  const todayIndex = (new Date().getDay() + 6) % 7; // Mon=0 .. Sun=6

  return (
    <div className="flex items-center justify-between gap-1 sm:gap-2">
      {DAY_LABELS.map((label, i) => {
        const isToday = i === todayIndex;
        const entry = MOCK_WEEK[i];
        return (
          <div
            key={label}
            className={cn(
              "flex flex-1 flex-col items-center gap-1.5 rounded-xl py-2.5",
              isToday && "bg-accent",
            )}
          >
            <span className={cn("text-xs font-medium", isToday ? "text-primary" : "text-muted-foreground")}>
              {label}
            </span>
            <span className={cn("size-2.5 rounded-full", DOT_COLOR[entry.status])} />
          </div>
        );
      })}
    </div>
  );
}
