import type { LucideIcon } from "lucide-react";
import {
  LayoutDashboard,
  ClipboardCheck,
  Users,
  Building2,
  Clock,
  CalendarClock,
  MapPin,
  Briefcase,
  FileBarChart,
  Settings,
} from "lucide-react";

export type NavItem = {
  label: string;
  href: string;
  icon: LucideIcon;
};

export type NavGroup = {
  label: string;
  items: NavItem[];
};

// Mapped to routes that actually exist in apps/web/app/admin/** -- every
// href below has a real page behind it. "Absensi" points at /admin/monitoring
// (the live check-in/out feed), "Jadwal & Shift" at /admin/config/shifts
// (there's no separate scheduling UI yet, shift config is the closest real
// page), "Laporan" at the monthly recap grid.
export const NAV_GROUPS: NavGroup[] = [
  {
    label: "Menu Utama",
    items: [
      { label: "Dashboard", href: "/admin/dashboard", icon: LayoutDashboard },
      { label: "Absensi", href: "/admin/monitoring", icon: ClipboardCheck },
      { label: "Karyawan", href: "/admin/employees", icon: Users },
      { label: "Departemen", href: "/admin/departments", icon: Building2 },
      { label: "Jadwal & Shift", href: "/admin/config/shifts", icon: Clock },
    ],
  },
  {
    label: "Manajemen",
    items: [
      { label: "Cuti & Izin", href: "/admin/leave-requests", icon: CalendarClock },
      { label: "Lokasi / Geofence", href: "/admin/config/office-locations", icon: MapPin },
      { label: "Dinas Luar", href: "/admin/config/field-assignments", icon: Briefcase },
      { label: "Laporan", href: "/admin/reports/recap", icon: FileBarChart },
    ],
  },
  {
    label: "Lainnya",
    items: [{ label: "Pengaturan", href: "/admin/settings", icon: Settings }],
  },
];

// /admin/settings now exists (Akun, Preferensi Notifikasi, Profil
// Perusahaan, links to config domains) -- moved out of "coming soon" into
// NAV_GROUPS below. /admin/help still doesn't exist as a page; the
// topbar's former chat button was repurposed into a direct mailto: link
// instead (Topbar.tsx). Kept as an array (not deleted) in case a real
// "coming soon" item shows up later.
export const NAV_COMING_SOON: NavItem[] = [];
