import type { ReactNode } from "react";
import { cookies } from "next/headers";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/admin-shell/AppSidebar";
import { CompanyProvider } from "@/components/admin-shell/CompanyContext";
import { Topbar } from "@/components/admin-shell/Topbar";
import { adminApi } from "@/lib/authedApi";
import type { CompanySettings } from "@absensi-next/contracts";
import type { SessionClaims } from "@/lib/session";

const SIDEBAR_COOKIE_NAME = "sidebar_state";

// Server Component wrapper: reads the sidebar's open/collapsed cookie before
// the first paint so SidebarProvider's defaultOpen matches what the user
// left it as -- no expanded->collapsed flash on reload (the cookie itself is
// written client-side by the shadcn Sidebar primitive on toggle).
export async function AdminShell({ session, children }: { session: SessionClaims; children: ReactNode }) {
  const cookieStore = await cookies();
  const sidebarOpen = cookieStore.get(SIDEBAR_COOKIE_NAME)?.value !== "false";

  // Best-effort: GetCompanySettings is admin-tier (any role can read, D-7),
  // so this should succeed whenever the shell itself is reachable. A null
  // fallback just means AppSidebar shows its generic default instead of a
  // broken page.
  let company: CompanySettings | null = null;
  try {
    company = await adminApi.get<CompanySettings>("/api/v1/admin/config/company");
  } catch {
    // ignored -- sidebar falls back to a generic label below
  }

  return (
    <SidebarProvider
      defaultOpen={sidebarOpen}
      style={{ "--sidebar-width": "16.25rem", "--sidebar-width-icon": "4.5rem" } as React.CSSProperties}
    >
      <CompanyProvider initial={company}>
        <AppSidebar />
        <SidebarInset className="bg-background">
          {/* No "/auth/me" endpoint exists yet -- the JWT only carries user id
              + role, not a display name. Topbar shows the role as identity
              (e.g. "Superadmin") rather than inventing a username. */}
          <Topbar role={session.role} />
          {/* min-w-0: same reasoning as SidebarInset's -- this is also a flex
              child (of the flex-col <main>), so it needs the same floor
              removed in case a future page nests another flex row here. */}
          <div className="min-w-0 flex-1 p-4 md:p-6">{children}</div>
        </SidebarInset>
      </CompanyProvider>
    </SidebarProvider>
  );
}
