import { redirect } from "next/navigation";
import { getEmployeeSession } from "@/lib/session";
import { employeeApi, ApiError } from "@/lib/authedApi";
import type { Profile } from "@absensi-next/contracts";
import { ProfileForm } from "@/components/profile/ProfileForm";

// Auth check for the SHELL now lives in the route group's layout.tsx --
// this page's own check stays too (it needs the session either way to know
// whether to redirect before fetching profile data below; harmless
// duplicate cookie read, not a network call).
export default async function EmployeeProfilePage() {
  const session = await getEmployeeSession();
  if (!session) redirect("/");

  let profile: Profile;
  try {
    profile = await employeeApi.get<Profile>("/api/v1/me");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/");
    throw err;
  }

  return (
    <div className="mx-auto max-w-lg space-y-4">
      <div>
        <h1 className="text-xl font-semibold text-foreground">Profil</h1>
        <p className="mt-1 text-sm text-muted-foreground">Kelola foto, nomor HP, dan password akun Anda.</p>
      </div>
      <ProfileForm profile={profile} settingsHref="/settings" />
    </div>
  );
}
