import { redirect } from "next/navigation";
import { getAdminSession } from "@/lib/session";
import { adminApi, ApiError } from "@/lib/authedApi";
import type { Profile } from "@absensi-next/contracts";
import { PageHeader } from "@/components/dashboard/PageHeader";
import { ProfileForm } from "@/components/profile/ProfileForm";

// Auth check for <AdminShell> now lives in the route group's layout.tsx --
// this page still needs its own session read before fetching /api/v1/me.
export default async function AdminProfilePage() {
  const session = await getAdminSession();
  if (!session) redirect("/admin/login");

  let profile: Profile;
  try {
    profile = await adminApi.get<Profile>("/api/v1/me");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/admin/login");
    throw err;
  }

  return (
    <div className="mx-auto max-w-lg space-y-6">
      <PageHeader title="Profil" description="Kelola foto, nomor HP, dan password akun Anda." />
      <ProfileForm profile={profile} settingsHref="/admin/settings" />
    </div>
  );
}
