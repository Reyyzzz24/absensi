"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { EMPLOYEE_NAV } from "@/components/employee-shell/nav-config";
import { NavPendingDot } from "@/components/nav/NavPendingDot";

// Bottom tab bar for < md -- a better fit than a hamburger/sheet for
// exactly 4 top-level destinations on a self-service mobile portal (spec
// explicitly calls this out as the preferred pattern here).
export function EmployeeBottomTabs() {
  const pathname = usePathname();

  return (
    <nav
      className="fixed inset-x-0 bottom-0 z-20 border-t border-border bg-white/95 backdrop-blur pb-[env(safe-area-inset-bottom)] md:hidden"
      aria-label="Navigasi utama"
    >
      <div className="grid grid-cols-4">
        {EMPLOYEE_NAV.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              aria-current={isActive ? "page" : undefined}
              className={cn(
                "flex flex-col items-center gap-1 py-2.5 text-[11px] font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset",
                isActive ? "text-primary" : "text-muted-foreground",
              )}
            >
              <item.icon className={cn("size-5", isActive && "fill-primary/10")} />
              {item.label}
              <NavPendingDot className="mt-0.5 ml-0" />
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
