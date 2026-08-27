"use client";

import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

const MONTHS = ["Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"];

// TODO: mock -- there is no endpoint returning aggregate attendance counts
// per month across a year (the real /admin/reports/recap endpoint returns a
// day-by-day grid for ONE month at a time, not a 12-month rollup). Values
// below are placeholder data so the chart has something to render; replace
// once a monthly-aggregate endpoint exists.
const MOCK_DATA = MONTHS.map((month, i) => ({
  month,
  hadir: Math.round(70 + Math.sin(i / 2) * 15 + (i % 3) * 4),
}));

function ChartTooltip({ active, payload, label }: { active?: boolean; payload?: { value: number }[]; label?: string }) {
  if (!active || !payload?.length) return null;
  return (
    <div className="rounded-lg border border-border bg-white px-3 py-2 shadow-md">
      <p className="text-xs font-medium text-foreground">{label}</p>
      <p className="text-xs text-muted-foreground">
        Hadir: <span className="font-semibold text-primary">{payload[0].value}</span>
      </p>
    </div>
  );
}

export function AttendanceTrendChart() {
  return (
    <ResponsiveContainer width="100%" height={260}>
      <AreaChart data={MOCK_DATA} margin={{ top: 8, right: 8, left: -16, bottom: 0 }}>
        <defs>
          <linearGradient id="trendFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#2F6BFF" stopOpacity={0.28} />
            <stop offset="100%" stopColor="#2F6BFF" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid vertical={false} stroke="#EDF0F4" />
        <XAxis
          dataKey="month"
          axisLine={false}
          tickLine={false}
          tick={{ fill: "#94A3B8", fontSize: 12 }}
        />
        <YAxis axisLine={false} tickLine={false} tick={{ fill: "#94A3B8", fontSize: 12 }} width={32} />
        <Tooltip content={<ChartTooltip />} />
        <Area
          type="monotone"
          dataKey="hadir"
          stroke="#2F6BFF"
          strokeWidth={2.5}
          fill="url(#trendFill)"
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
