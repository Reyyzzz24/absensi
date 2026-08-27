import Link from "next/link";
import { ChevronRight, Languages } from "lucide-react";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { PasswordChangeForm } from "@/components/profile/PasswordChangeForm";
import { NotificationPreferencesForm } from "@/components/settings/NotificationPreferencesForm";
import { CompanySettingsForm } from "@/components/settings/CompanySettingsForm";
import type { CompanySettings } from "@absensi-next/contracts";

// Kept as sections rather than a full custom tab widget -- no Tabs
// primitive exists in components/ui yet, and one ChartCard per concern
// already reads as clearly separated (not a single "god" form) without
// introducing a new dependency just for this page.
export function SettingsBody({
  isAdmin,
  company,
  configLinks,
}: {
  isAdmin: boolean;
  company?: CompanySettings;
  configLinks?: { label: string; description: string; href: string }[];
}) {
  const audience = isAdmin ? "admin" : "employee";
  return (
    <div className="space-y-6">
      <ChartCard title="Akun" description="Ganti password akun Anda.">
        <PasswordChangeForm audience={audience} />
      </ChartCard>

      <ChartCard title="Preferensi Notifikasi" description="Pilih jenis notifikasi yang ingin Anda terima.">
        <NotificationPreferencesForm audience={audience} />
      </ChartCard>

      {isAdmin && company && (
        <ChartCard title="Profil Perusahaan" description="Nama dan logo yang tampil di panel admin.">
          <CompanySettingsForm company={company} />
        </ChartCard>
      )}

      {isAdmin && configLinks && configLinks.length > 0 && (
        <ChartCard
          title="Konfigurasi Operasional"
          description="Sudah punya halaman kelola tersendiri -- ditautkan di sini, bukan dibangun ulang."
        >
          <ul className="divide-y divide-border">
            {configLinks.map((link) => (
              <li key={link.href}>
                <Link
                  href={link.href}
                  className="group flex items-center gap-3 py-3 first:pt-0 last:pb-0 focus-visible:outline-none"
                >
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-foreground">{link.label}</p>
                    <p className="text-xs text-muted-foreground">{link.description}</p>
                  </div>
                  <ChevronRight className="size-4 shrink-0 text-muted-foreground transition group-hover:translate-x-0.5" />
                </Link>
              </li>
            ))}
          </ul>
        </ChartCard>
      )}

      {!isAdmin && (
        <ChartCard title="Bahasa" description="Preferensi bahasa aplikasi.">
          <div
            className="flex items-center justify-between gap-3 opacity-60"
            title="Segera hadir -- aplikasi baru tersedia dalam Bahasa Indonesia"
          >
            <div className="flex items-center gap-2 text-sm text-foreground">
              <Languages className="size-4" />
              Bahasa Indonesia / English
            </div>
            <span className="rounded-full bg-secondary px-2 py-0.5 text-xs font-medium text-muted-foreground">
              Segera hadir
            </span>
          </div>
        </ChartCard>
      )}
    </div>
  );
}
