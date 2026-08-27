import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { StatCardSkeleton } from "@/components/dashboard/StatCard";

export function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <StatCardSkeleton key={i} />
        ))}
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <Card className="rounded-2xl border border-border p-5 shadow-sm lg:col-span-2">
          <Skeleton className="h-[260px] w-full" />
        </Card>
        <Card className="rounded-2xl border border-border p-5 shadow-sm">
          <Skeleton className="h-[220px] w-full" />
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <Card className="rounded-2xl border border-border p-5 shadow-sm lg:col-span-1">
          <Skeleton className="h-[200px] w-full" />
        </Card>
        <Card className="rounded-2xl border border-border p-5 shadow-sm lg:col-span-2">
          <Skeleton className="h-[200px] w-full" />
        </Card>
      </div>
    </div>
  );
}
