import { SettingsBody } from "@/components/settings/SettingsBody";

// Auth check + <EmployeeShell> now live in the route group's layout.tsx
// (perf fix -- see that file's comment). This page renders content only.
export default function EmployeeSettingsPage() {
  return (
    <div className="mx-auto max-w-lg space-y-4">
      <div>
        <h1 className="text-xl font-semibold text-foreground">Pengaturan</h1>
        <p className="mt-1 text-sm text-muted-foreground">Akun dan preferensi notifikasi.</p>
      </div>
      <SettingsBody isAdmin={false} />
    </div>
  );
}
