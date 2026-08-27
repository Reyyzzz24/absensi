import { StatusBadge, type AttendanceStatusLabel } from "@/components/dashboard/StatusBadge";

type ActivityRow = {
  date: string;
  checkIn: string;
  checkOut: string;
  status: AttendanceStatusLabel;
  duration: string;
};

// TODO: mock -- docs/openapi.yaml documents a `GET /attendance/history`
// endpoint (month/year query params) but it was never implemented on the
// backend (no handler, no route). This list is illustrative only; wiring it
// to real data needs that endpoint built first (backend change, out of
// scope for this frontend-only task -- flagged for a decision, not silently
// added).
const MOCK_ACTIVITY: ActivityRow[] = [
  { date: "Rab, 26 Agu", checkIn: "08:12", checkOut: "17:05", status: "telat", duration: "8j 53m" },
  { date: "Sel, 25 Agu", checkIn: "07:55", checkOut: "17:00", status: "hadir", duration: "9j 5m" },
  { date: "Sen, 24 Agu", checkIn: "-", checkOut: "-", status: "izin", duration: "-" },
  { date: "Jum, 21 Agu", checkIn: "07:58", checkOut: "16:30", status: "hadir", duration: "8j 32m" },
  { date: "Kam, 20 Agu", checkIn: "-", checkOut: "-", status: "alpha", duration: "-" },
];

export function RecentActivityList() {
  return (
    <div className="divide-y divide-border">
      {MOCK_ACTIVITY.map((row) => (
        <div key={row.date} className="flex items-center justify-between gap-3 py-3 text-sm">
          <div className="min-w-0">
            <p className="font-medium text-foreground">{row.date}</p>
            <p className="text-xs text-muted-foreground">
              {row.checkIn} &ndash; {row.checkOut}
              {row.duration !== "-" && ` · ${row.duration}`}
            </p>
          </div>
          <StatusBadge status={row.status} />
        </div>
      ))}
    </div>
  );
}
