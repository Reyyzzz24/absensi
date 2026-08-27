import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { StatCardSkeleton } from "@/components/dashboard/StatCard";

export function EmployeeDashboardSkeleton() {
  return (
    <div className="space-y-6">
      <Card className="rounded-2xl border border-border p-6 shadow-sm sm:p-7">
        <Skeleton className="h-10 w-48" />
        <Skeleton className="mt-3 h-4 w-64" />
        <Skeleton className="mt-6 h-20 w-full" />
      </Card>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 sm:gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <StatCardSkeleton key={i} />
        ))}
      </div>

      <Card className="rounded-2xl border border-border p-5 shadow-sm">
        <Skeleton className="h-16 w-full" />
      </Card>

      <Card className="rounded-2xl border border-border p-5 shadow-sm">
        <Skeleton className="h-48 w-full" />
      </Card>
    </div>
  );
}
