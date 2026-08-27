"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { CalendarCheck2, LogOut, Settings, User as UserIcon } from "lucide-react";
import { NotificationBell } from "@/components/notifications/NotificationBell";
import { NavPendingDot } from "@/components/nav/NavPendingDot";
import { cn } from "@/lib/utils";
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
import { EMPLOYEE_NAV } from "@/components/employee-shell/nav-config";
import { useApiMutation } from "@/lib/useApiMutation";

function initials(name: string) {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase())
    .join("");
}

// displayName: no "/auth/me" endpoint exists for employees either (same gap
// as the admin topbar) -- the JWT only carries the employee id, and
// /attendance/today only echoes the employee's name back once they have an
// attendance row for today. Falls back to a generic label rather than
// inventing one.
export function EmployeeTopBar({ displayName }: { displayName: string }) {
  const pathname = usePathname();
  const router = useRouter();
  const logout = useApiMutation<Record<string, never>>("/api/auth/logout?aud=employee");

  async function handleLogout() {
    await logout.mutateAsync({});
    router.push("/");
  }

  return (
    <header className="sticky top-0 z-20 border-b border-border bg-white/90 backdrop-blur supports-backdrop-filter:bg-white/70">
      <div className="mx-auto flex h-16 max-w-5xl items-center gap-4 px-4 sm:px-6">
        <Link href="/dashboard" className="flex shrink-0 items-center gap-2">
          <div className="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <CalendarCheck2 className="size-4.5" />
          </div>
          <span className="hidden text-sm font-semibold text-foreground sm:inline">Absensi Next</span>
        </Link>

        <nav className="hidden flex-1 items-center gap-1 md:flex">
          {EMPLOYEE_NAV.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                aria-current={isActive ? "page" : undefined}
                className={cn(
                  "relative flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                  isActive ? "text-primary" : "text-muted-foreground hover:text-foreground hover:bg-secondary",
                )}
              >
                <item.icon className="size-4" />
                {item.label}
                <NavPendingDot />
                {isActive && <span className="absolute inset-x-3 -bottom-[1px] h-0.5 rounded-full bg-primary" />}
              </Link>
            );
          })}
        </nav>

        <div className="ml-auto flex items-center gap-1">
          <NotificationBell audience="employee" />

          <DropdownMenu>
            <DropdownMenuTrigger className="ml-1 flex items-center gap-2 rounded-lg p-1 outline-none transition hover:bg-secondary focus-visible:ring-2 focus-visible:ring-ring">
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
                  <span className="text-xs font-normal text-muted-foreground">Akun karyawan</span>
                </DropdownMenuLabel>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuGroup>
                <DropdownMenuItem render={<Link href="/profile" />} nativeButton={false}>
                  <UserIcon />
                  Profil
                </DropdownMenuItem>
                <DropdownMenuItem render={<Link href="/settings" />} nativeButton={false}>
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
      </div>
    </header>
  );
}
