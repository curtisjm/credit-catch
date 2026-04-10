"use client";

import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from "recharts";
import type { MonthlyDashboard } from "@/types/api";

const MONTHS = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];

function formatDollars(cents: number): string {
  return `$${(cents / 100).toFixed(0)}`;
}

export function MonthlyChart({ data }: { data: MonthlyDashboard[] }) {
  const chartData = data.map((d) => ({
    month: MONTHS[d.month - 1],
    used: d.total_used_cents / 100,
    unused: (d.total_available_cents - d.total_used_cents) / 100,
  }));

  return (
    <ResponsiveContainer width="100%" height={240}>
      <BarChart data={chartData} barGap={2}>
        <CartesianGrid
          strokeDasharray="3 3"
          stroke="var(--border)"
          vertical={false}
        />
        <XAxis
          dataKey="month"
          tick={{ fill: "var(--muted-foreground)", fontSize: 12 }}
          axisLine={false}
          tickLine={false}
        />
        <YAxis
          tick={{ fill: "var(--muted-foreground)", fontSize: 12 }}
          axisLine={false}
          tickLine={false}
          tickFormatter={(v) => `$${v}`}
          width={50}
        />
        <Tooltip
          contentStyle={{
            background: "var(--popover)",
            border: "1px solid var(--border)",
            borderRadius: "var(--radius)",
            color: "var(--popover-foreground)",
            fontSize: 13,
          }}
          formatter={(value, name) => [
            formatDollars(Number(value ?? 0) * 100),
            name === "used" ? "Used" : "Unused",
          ]}
        />
        <Bar
          dataKey="used"
          fill="var(--primary)"
          radius={[4, 4, 0, 0]}
          stackId="credits"
        />
        <Bar
          dataKey="unused"
          fill="var(--muted)"
          radius={[4, 4, 0, 0]}
          stackId="credits"
        />
      </BarChart>
    </ResponsiveContainer>
  );
}
