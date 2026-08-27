"use client";

import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";

export type StatusDonutDatum = { name: string; value: number; color: string };

export function StatusDonutChart({ data }: { data: StatusDonutDatum[] }) {
  const total = data.reduce((sum, d) => sum + d.value, 0);

  return (
    <div className="flex flex-col items-center gap-4 sm:flex-row sm:gap-6">
      <div className="relative h-[180px] w-[180px] shrink-0">
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={data}
              dataKey="value"
              nameKey="name"
              innerRadius={58}
              outerRadius={80}
              paddingAngle={total > 0 ? 3 : 0}
              strokeWidth={0}
            >
              {data.map((entry) => (
                <Cell key={entry.name} fill={entry.color} />
              ))}
            </Pie>
            <Tooltip
              content={({ active, payload }) => {
                if (!active || !payload?.length) return null;
                const p = payload[0];
                return (
                  <div className="rounded-lg border border-border bg-white px-3 py-2 shadow-md">
                    <p className="text-xs font-medium text-foreground">
                      {p.name}: <span className="font-semibold">{p.value}</span>
                    </p>
                  </div>
                );
              }}
            />
          </PieChart>
        </ResponsiveContainer>
        <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
          <span className="text-2xl font-bold text-foreground">{total}</span>
          <span className="text-[11px] text-muted-foreground">Karyawan</span>
        </div>
      </div>

      <ul className="flex flex-1 flex-col gap-2.5">
        {data.map((entry) => (
          <li key={entry.name} className="flex items-center justify-between gap-3 text-sm">
            <span className="flex items-center gap-2 text-foreground">
              <span className="size-2.5 shrink-0 rounded-full" style={{ backgroundColor: entry.color }} />
              {entry.name}
            </span>
            <span className="font-medium text-foreground">{entry.value}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
