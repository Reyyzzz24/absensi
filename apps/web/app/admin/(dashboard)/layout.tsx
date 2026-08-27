import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { getAdminSession } from "@/lib/session";
import { AdminShell } from "@/components/admin-shell/AdminShell";

// Perf fix (nav-speed audit): AdminShell (sidebar + topbar) used to be
// instantiated fresh by every individual admin page.tsx -- confirmed via a
// DOM-identity probe that this fully unmounts/remounts the shell (and
// everything inside it, incl. NotificationBell's query) on every
// navigation between admin pages, even though the RSC payload itself was
// already small (~2KB/125ms). Hoisting it here, into a layout that wraps
// every route in this group, lets Next.js's router keep the same shell
// instance across navigations -- only `children` swaps. `/admin/login`
// deliberately sits OUTSIDE this route group (sibling `app/admin/login/`)
// since it must not show the authenticated shell and has its own
// already-logged-in redirect the other direction.
//
// Individual pages below still call getAdminSession() themselves too where
// they need `session.role` for RBAC display logic (D-7) -- that's an
// unchanged, cheap cookie-decode, not a network call, so duplicating it is
// harmless and was deliberately left alone rather than threading session
// down as a prop (smaller, safer diff; not part of what was measured slow).
export default async function AdminDashboardLayout({ children }: { children: ReactNode }) {
  const session = await getAdminSession();
  if (!session) redirect("/admin/login");

  return <AdminShell session={session}>{children}</AdminShell>;
}
