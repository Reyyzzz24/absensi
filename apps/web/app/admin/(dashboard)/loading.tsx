import { Skeleton } from "@/components/ui/skeleton";

// Route-group loading fallback (nav-speed audit): most admin pages block on
// a single `await adminApi.get(...)` in the page component itself with no
// internal Suspense boundary of their own (unlike /admin/dashboard, which
// already streams). Before this file existed, that wait showed nothing --
// the layout (sidebar/topbar) now renders instantly regardless since it's
// a separate segment, and this skeleton fills the content area while the
// page's own fetch is still in flight.
export default function AdminSectionLoading() {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <Skeleton className="h-7 w-48" />
        <Skeleton className="h-4 w-96 max-w-full" />
      </div>
      <div className="space-y-4 rounded-2xl border border-border bg-white p-5 shadow-sm">
        <Skeleton className="h-5 w-32" />
        <Skeleton className="h-24 w-full" />
      </div>
      <div className="space-y-3 rounded-2xl border border-border bg-white p-5 shadow-sm">
        <Skeleton className="h-5 w-32" />
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    </div>
  );
}
