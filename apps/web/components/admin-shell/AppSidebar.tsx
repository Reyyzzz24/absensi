"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Building2, CalendarCheck2, ChevronsUpDown } from "lucide-react";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { NAV_GROUPS, NAV_COMING_SOON } from "@/components/admin-shell/nav-config";
import { NavPendingDot } from "@/components/nav/NavPendingDot";

export function AppSidebar() {
  const pathname = usePathname();

  return (
    <Sidebar
      collapsible="icon"
      className="border-none bg-gradient-to-b from-[#0B1130] to-[#141B4D] text-sidebar-foreground [&_[data-slot=sidebar-container]]:border-none"
    >
      <SidebarHeader className="gap-3 p-3">
        <div className="flex items-center justify-between gap-2 px-1 group-data-[collapsible=icon]:flex-col group-data-[collapsible=icon]:gap-2">
          {/* Org identity -- this app has no multi-tenant/workspace concept,
              so this is a static brand card (not a functional switcher) to
              match the reference layout's visual slot. */}
          <div className="flex min-w-0 items-center gap-2 overflow-hidden">
            <div className="relative flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-[0_0_16px_rgba(47,107,255,0.55)]">
              <CalendarCheck2 className="size-4.5" />
            </div>
            <div className="flex min-w-0 flex-col group-data-[collapsible=icon]:hidden">
              <span className="truncate text-sm font-semibold text-white">Absensi Next</span>
              <span className="truncate text-xs text-sidebar-foreground/60">Panel Admin</span>
            </div>
          </div>
          {/* Kept visible (not hidden) when collapsed -- otherwise there is
              no mouse-accessible way back to expanded mode, only the
              keyboard shortcut. */}
          <SidebarTrigger className="size-7 shrink-0 text-sidebar-foreground/70 hover:bg-white/8 hover:text-white" />
        </div>

        {/* Workspace card -- decorative, mirrors the reference's org
            switcher shape, but there is nothing to switch to. */}
        <button
          type="button"
          className="flex items-center gap-2 rounded-xl bg-white/5 px-2.5 py-2 text-left transition hover:bg-white/8 group-data-[collapsible=icon]:hidden"
        >
          <div className="flex size-7 shrink-0 items-center justify-center rounded-md bg-[#3B82F6]/20 text-[#93C5FD]">
            <Building2 className="size-4" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-xs font-medium text-white">PT Absensi Digital</p>
            <p className="truncate text-[11px] text-sidebar-foreground/50">Workspace</p>
          </div>
          <ChevronsUpDown className="size-3.5 shrink-0 text-sidebar-foreground/40" />
        </button>
      </SidebarHeader>

      <SidebarContent className="px-2">
        {NAV_GROUPS.map((group) => (
          <SidebarGroup key={group.label}>
            <SidebarGroupLabel className="px-2 text-[10px] font-semibold tracking-wider text-sidebar-foreground/40 uppercase">
              {group.label}
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {group.items.map((item) => {
                  const isActive = pathname === item.href || pathname?.startsWith(item.href + "/");
                  return (
                    <SidebarMenuItem key={item.href}>
                      <SidebarMenuButton
                        render={<Link href={item.href} />}
                        isActive={isActive}
                        tooltip={item.label}
                        className="text-sidebar-foreground/80 data-active:bg-white data-active:text-slate-900 data-active:shadow-sm hover:bg-white/8 data-active:hover:bg-white"
                      >
                        <item.icon />
                        <span>{item.label}</span>
                        <NavPendingDot />
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  );
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}

        {NAV_COMING_SOON.length > 0 && (
          <SidebarGroup>
            <SidebarGroupLabel className="px-2 text-[10px] font-semibold tracking-wider text-sidebar-foreground/40 uppercase">
              Segera Hadir
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {NAV_COMING_SOON.map((item) => (
                  <SidebarMenuItem key={item.label}>
                    <SidebarMenuButton
                      disabled
                      tooltip={`${item.label} (segera hadir)`}
                      className="cursor-not-allowed text-sidebar-foreground/35 opacity-60"
                    >
                      <item.icon />
                      <span>{item.label}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        )}
      </SidebarContent>
    </Sidebar>
  );
}
