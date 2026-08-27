"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { DataTable } from "@/components/dashboard/DataTable";
import { StatusBadge } from "@/components/dashboard/StatusBadge";
import type { Attendance } from "@/lib/types";

function initials(name: string) {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase())
    .join("");
}

// Fixed to Asia/Jakarta (not the browser's local zone) on purpose: this is a
// Client Component whose initial HTML is still produced during SSR, so
// letting the format follow "wherever this happens to render" causes a
// server/client hydration mismatch whenever the server and browser are in
// different timezones -- and the app's source of truth is WIB regardless
// (LOGIC_SPEC.md §1), not "whatever timezone the viewer's machine is set to".
function formatTime(iso?: string) {
  if (!iso) return "—";
  return new Date(iso).toLocaleTimeString("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    timeZone: "Asia/Jakarta",
  });
}

const columns: ColumnDef<Attendance, unknown>[] = [
  {
    id: "employee",
    header: "Karyawan",
    cell: ({ row }) => {
      const name = row.original.employee?.full_name ?? `Karyawan #${row.original.employee_id}`;
      return (
        <div className="flex items-center gap-2.5">
          <Avatar className="size-8">
            <AvatarFallback className="bg-primary/10 text-[11px] font-semibold text-primary">
              {initials(name)}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <p className="truncate text-sm font-medium text-foreground">{name}</p>
            {row.original.employee?.nik && (
              <p className="truncate text-xs text-muted-foreground">{row.original.employee.nik}</p>
            )}
          </div>
        </div>
      );
    },
  },
  {
    id: "check_in",
    header: "Jam Masuk",
    cell: ({ row }) => <span className="text-sm text-foreground">{formatTime(row.original.check_in_at)}</span>,
  },
  {
    id: "check_out",
    header: "Jam Keluar",
    cell: ({ row }) => <span className="text-sm text-foreground">{formatTime(row.original.check_out_at)}</span>,
  },
  {
    id: "status",
    header: "Status",
    cell: ({ row }) => <StatusBadge status={row.original.is_late ? "telat" : "hadir"} />,
  },
];

export function RecentAttendanceTable({ data }: { data: Attendance[] }) {
  return <DataTable columns={columns} data={data} emptyMessage="Belum ada presensi hari ini." />;
}
