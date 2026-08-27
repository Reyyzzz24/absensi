import type { LucideIcon } from "lucide-react";
import { LayoutDashboard, ScanFace, CalendarClock, ListChecks } from "lucide-react";

export type EmployeeNavItem = {
  label: string;
  href: string;
  icon: LucideIcon;
};

// All four routes already exist (apps/web/app/{dashboard,checkin,leave,tasks}) --
// this is the same set the old EmployeeNav.tsx linked to, just with icons for
// the new app bar / bottom tab bar.
export const EMPLOYEE_NAV: EmployeeNavItem[] = [
  { label: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
  { label: "Check-in", href: "/checkin", icon: ScanFace },
  { label: "Izin/Sakit", href: "/leave", icon: CalendarClock },
  { label: "Task", href: "/tasks", icon: ListChecks },
];
