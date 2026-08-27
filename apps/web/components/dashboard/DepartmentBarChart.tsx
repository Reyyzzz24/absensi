"use client";

import { Bar, BarChart, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

export type DepartmentBarDatum = { name: string; value: number };

const COLORS = ["#3B82F6", "#10B981", "#8B5CF6", "#F97316", "#F59E0B"];

export function DepartmentBarChart({ data }: { data: DepartmentBarDatum[] }) {
  const height = Math.max(160, data.length * 44);

  return (
    <ResponsiveContainer width="100%" height={height}>
      <BarChart data={data} layout="vertical" margin={{ top: 0, right: 16, left: 0, bottom: 0 }}>
        <XAxis type="number" hide />
        <YAxis
          type="category"
          dataKey="name"
          axisLine={false}
          tickLine={false}
          width={120}
          tick={{ fill: "#0F172A", fontSize: 13 }}
        />
        <Tooltip
          cursor={{ fill: "#F1F5F9" }}
          content={({ active, payload }) => {
            if (!active || !payload?.length) return null;
            const p = payload[0];
            return (
              <div className="rounded-lg border border-border bg-white px-3 py-2 shadow-md">
                <p className="text-xs font-medium text-foreground">
                  {p.payload.name}: <span className="font-semibold">{p.value}</span> hadir
                </p>
              </div>
            );
          }}
        />
        <Bar dataKey="value" radius={[0, 8, 8, 0]} maxBarSize={22}>
          {data.map((entry, i) => (
            <Cell key={entry.name} fill={COLORS[i % COLORS.length]} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
}
