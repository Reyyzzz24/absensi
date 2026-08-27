import type { LucideIcon } from "lucide-react";
import { Card } from "@/components/ui/card";
import { IconTile, type IconTileTone } from "@/components/dashboard/IconTile";
import { TrendPill } from "@/components/dashboard/TrendPill";
import { Skeleton } from "@/components/ui/skeleton";

export function StatCard({
  label,
  value,
  icon,
  tone,
  trend,
}: {
  label: string;
  value: string | number;
  icon: LucideIcon;
  tone: IconTileTone;
  trend?: { direction: "up" | "down"; value: string };
}) {
  return (
    <Card className="gap-3 rounded-2xl border border-border p-5 shadow-sm">
      <div className="flex items-start justify-between">
        <IconTile icon={icon} tone={tone} />
        {trend && <TrendPill direction={trend.direction} value={trend.value} />}
      </div>
      <div>
        <p className="text-[28px] font-bold leading-tight text-foreground">{value}</p>
        <p className="mt-0.5 text-sm text-muted-foreground">{label}</p>
      </div>
    </Card>
  );
}

export function StatCardSkeleton() {
  return (
    <Card className="gap-3 rounded-2xl border border-border p-5 shadow-sm">
      <div className="flex items-start justify-between">
        <Skeleton className="size-11 rounded-xl" />
        <Skeleton className="h-5 w-12 rounded-full" />
      </div>
      <div className="space-y-2">
        <Skeleton className="h-7 w-16" />
        <Skeleton className="h-4 w-24" />
      </div>
    </Card>
  );
}
