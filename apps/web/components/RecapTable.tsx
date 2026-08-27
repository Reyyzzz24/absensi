"use client";

import { useMemo, useState } from "react";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
  type SortingState,
} from "@tanstack/react-table";
import type { EmployeeRecap, MonthRecap, RecapDayStatus } from "@/lib/types";

const STATUS_STYLE: Record<RecapDayStatus, string> = {
  hadir: "bg-green-100 text-green-700",
  izin: "bg-blue-100 text-blue-700",
  sakit: "bg-purple-100 text-purple-700",
  libur: "bg-slate-100 text-slate-400",
  alpha: "bg-red-100 text-red-700",
};

const STATUS_LETTER: Record<RecapDayStatus, string> = {
  hadir: "H",
  izin: "I",
  sakit: "S",
  libur: "L",
  alpha: "A",
};

const columnHelper = createColumnHelper<EmployeeRecap>();

// TanStack Table drives this grid (D-22/CLAUDE.md §2) instead of a hand-rolled
// <table> -- gives sortable summary columns (click "H"/"Telat"/"A" headers)
// for free without hand-writing comparator functions per column.
export function RecapTable({ recap }: { recap: MonthRecap }) {
  const [sorting, setSorting] = useState<SortingState>([]);

  const days = useMemo(() => Array.from({ length: recap.days_in_month }, (_, i) => i + 1), [recap.days_in_month]);

  const columns = useMemo(
    () => [
      columnHelper.accessor("full_name", {
        header: "Karyawan",
        cell: (info) => (
          <span className="whitespace-nowrap font-medium text-slate-700">
            {info.getValue()} <span className="font-mono text-slate-400">({info.row.original.nik})</span>
          </span>
        ),
      }),
      ...days.map((d) =>
        columnHelper.display({
          id: `day-${d}`,
          header: () => <span className="text-slate-400">{d}</span>,
          cell: (info) => {
            const day = info.row.original.days[d - 1];
            if (!day) return null;
            // Hadir on a resolved holiday (D-25) -- e.g. voluntary/overtime
            // work on a libur day -- still shows "H" (they DID show up),
            // just with a ring so it reads differently from an ordinary
            // workday check-in.
            const holidayHadir = day.status === "hadir" && day.is_holiday;
            return (
              <span
                className={`inline-flex h-5 w-5 items-center justify-center rounded ${STATUS_STYLE[day.status]} ${
                  holidayHadir ? "ring-2 ring-amber-400" : ""
                }`}
                title={holidayHadir ? `${day.date} · hadir di hari libur` : day.date}
              >
                {day.is_late ? "*" : STATUS_LETTER[day.status]}
              </span>
            );
          },
        }),
      ),
      columnHelper.accessor((row) => row.summary.hadir ?? 0, {
        id: "hadir",
        header: "H",
        cell: (info) => <span className="text-slate-700">{info.getValue()}</span>,
      }),
      columnHelper.accessor((row) => row.summary.telat ?? 0, {
        id: "telat",
        header: "Telat",
        cell: (info) => <span className="text-orange-600">{info.getValue()}</span>,
      }),
      columnHelper.accessor((row) => row.summary.izin ?? 0, {
        id: "izin",
        header: "I",
        cell: (info) => <span className="text-slate-700">{info.getValue()}</span>,
      }),
      columnHelper.accessor((row) => row.summary.sakit ?? 0, {
        id: "sakit",
        header: "S",
        cell: (info) => <span className="text-slate-700">{info.getValue()}</span>,
      }),
      columnHelper.accessor((row) => row.summary.alpha ?? 0, {
        id: "alpha",
        header: "A",
        cell: (info) => <span className="text-red-600">{info.getValue()}</span>,
      }),
    ],
    [days],
  );

  const table = useReactTable({
    data: recap.employees,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  return (
    // Self-contained scroll wrapper -- this component owns its own
    // overflow-x-auto rather than relying on whatever className a caller
    // happens to pass to an outer <ChartCard>. A table this wide (up to 31
    // day columns + name + 5 summary columns) MUST be scrollable within
    // this box and never allowed to force the page itself to grow (see the
    // layout-overflow fix in components/ui/sidebar.tsx for the other half
    // of that bug). max-h caps vertical scroll too so the sticky header
    // below has a scroll container to stick within.
    // print:max-h-none + print:overflow-visible: on screen this box clips
    // to 70vh with its own scrollbar; when printing there's no scrolling,
    // so every row/column must render in full flow instead of being cut
    // off at the clipped box's edge. print:static on the sticky cells below
    // is for the same reason PLUS so the browser's native "repeat <thead>
    // per printed page" behavior applies -- some engines only do that for
    // plain table-layout positioning, not for position:sticky cells.
    <div className="max-h-[70vh] overflow-auto print:max-h-none print:overflow-visible">
      <table className="min-w-max border-collapse text-xs print:text-[8px]">
        <thead>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id} className="border-b border-slate-100 print:break-inside-avoid">
              {headerGroup.headers.map((header, i) => {
                const sortable = header.column.getCanSort();
                return (
                  <th
                    key={header.id}
                    className={`sticky top-0 bg-white p-2 text-center font-medium text-slate-500 print:static ${
                      i === 0 ? "left-0 z-[2] text-left" : "z-[1]"
                    } ${sortable ? "cursor-pointer select-none" : ""}`}
                    onClick={header.column.getToggleSortingHandler()}
                  >
                    {flexRender(header.column.columnDef.header, header.getContext())}
                    {{ asc: " ▲", desc: " ▼" }[header.column.getIsSorted() as string] ?? ""}
                  </th>
                );
              })}
            </tr>
          ))}
        </thead>
        <tbody>
          {table.getRowModel().rows.map((row) => (
            <tr key={row.id} className="border-b border-slate-50 print:break-inside-avoid">
              {row.getVisibleCells().map((cell, i) => (
                <td
                  key={cell.id}
                  className={`p-1 text-center ${
                    i === 0 ? "sticky left-0 z-[1] whitespace-nowrap bg-white p-2 text-left print:static" : ""
                  }`}
                >
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
