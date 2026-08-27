"use client";

import Link from "next/link";
import { LogOut, MessageSquare, Search, Settings, User as UserIcon } from "lucide-react";
import { NotificationBell } from "@/components/notifications/NotificationBell";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useApiMutation } from "@/lib/useApiMutation";
import { useRouter } from "next/navigation";

function initials(name: string) {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase())
    .join("");
}

export function Topbar({ role }: { role?: string }) {
  const router = useRouter();
  const displayName = role ? role.charAt(0).toUpperCase() + role.slice(1) : "Admin";
  const logout = useApiMutation<Record<string, never>>("/api/auth/logout?aud=admin");

  async function handleLogout() {
    await logout.mutateAsync({});
    router.push("/admin/login");
  }

  return (
    // z-0 (not z-10): the fixed sidebar container is z-10 (components/ui/sidebar.tsx).
    // This header only needs `sticky` for its own vertical stacking; giving
    // it an explicit LOWER z-index than the sidebar means that even if some
    // future layout bug lets this header's box overlap the sidebar's
    // horizontal position again, the sidebar still paints on top -- this
    // relied on DOM order (header after sidebar => same z-index tie broken
    // in the header's favor) before, which is exactly what let it visually
    // cover the sidebar during the horizontal-overflow bug fixed above.
    <header className="sticky top-0 z-0 flex h-16 shrink-0 items-center gap-3 border-b border-border bg-white px-4 md:px-6 print:hidden">
      <SidebarTrigger className="text-muted-foreground md:hidden" />

      {/* Plain GET form -- submitting (Enter) navigates to Karyawan with the
          query in ?q=, which the employees page reads and forwards to
          GET /admin/employees?q= (ILIKE on name/NIK, backend). Scoped to
          employee search only for now, matching the placeholder copy --
          not a cross-entity command palette (no absensi/laporan results). */}
      <form method="get" action="/admin/employees" className="relative flex-1 max-w-md">
        <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
        <input
          type="search"
          name="q"
          placeholder="Cari karyawan..."
          className="h-9 w-full rounded-lg border-0 bg-secondary pl-9 pr-14 text-sm text-foreground placeholder:text-muted-foreground focus:ring-2 focus:ring-ring focus:outline-none"
        />
        <kbd className="pointer-events-none absolute top-1/2 right-2 -translate-y-1/2 rounded-md border border-border bg-white px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
          ⌘K
        </kbd>
      </form>

      <div className="ml-auto flex items-center gap-1">
        {/* Repurposed from a dead "chat" placeholder (no messaging feature
            exists in this domain) into a real action: opens the user's mail
            client addressed to support. Not a fake chat widget. */}
        <a
          href="mailto:support@absensi.local?subject=Bantuan%20Panel%20Admin"
          className="relative flex size-9 items-center justify-center rounded-lg text-muted-foreground transition hover:bg-secondary hover:text-foreground"
          aria-label="Bantuan"
          title="Bantuan"
        >
          <MessageSquare className="size-[18px]" />
        </a>
        <NotificationBell audience="admin" />

        <div className="mx-2 h-6 w-px bg-border" />

        <DropdownMenu>
          <DropdownMenuTrigger className="flex items-center gap-2 rounded-lg p-1 outline-none transition hover:bg-secondary focus-visible:ring-2 focus-visible:ring-ring">
            <Avatar className="size-8">
              <AvatarFallback className="bg-primary/10 text-xs font-semibold text-primary">
                {initials(displayName)}
              </AvatarFallback>
            </Avatar>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuGroup>
              <DropdownMenuLabel className="flex flex-col">
                <span className="text-sm font-medium text-foreground">{displayName}</span>
                <span className="text-xs font-normal text-muted-foreground">Akun admin</span>
              </DropdownMenuLabel>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem render={<Link href="/admin/profile" />} nativeButton={false}>
                <UserIcon />
                Profil
              </DropdownMenuItem>
              <DropdownMenuItem render={<Link href="/admin/settings" />} nativeButton={false}>
                <Settings />
                Pengaturan
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem variant="destructive" onClick={handleLogout}>
                <LogOut />
                Logout
              </DropdownMenuItem>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
