import { Skeleton } from "@/components/ui/skeleton";

// Same fix as app/admin/(dashboard)/loading.tsx -- see that file's comment.
// /dashboard already streams its own body behind an internal Suspense
// (EmployeeDashboardSkeleton); this covers the other employee pages
// (/checkin, /leave, /tasks, /profile, /settings) that block on a single
// server-side fetch with no Suspense boundary of their own.
export default function EmployeeSectionLoading() {
  return (
    <div className="mx-auto max-w-lg space-y-4">
      <div className="space-y-2">
        <Skeleton className="h-6 w-40" />
        <Skeleton className="h-4 w-full max-w-sm" />
      </div>
      <div className="space-y-3 rounded-2xl border border-border bg-white p-4 shadow-sm">
        <Skeleton className="h-5 w-28" />
        <Skeleton className="h-20 w-full" />
      </div>
      <div className="space-y-3 rounded-2xl border border-border bg-white p-4 shadow-sm">
        <Skeleton className="h-5 w-28" />
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-10 w-full" />
        ))}
      </div>
    </div>
  );
}
