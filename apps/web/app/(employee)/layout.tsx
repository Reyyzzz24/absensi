import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { getEmployeeSession } from "@/lib/session";
import { EmployeeShell } from "@/components/employee-shell/EmployeeShell";

// Same fix as app/admin/(dashboard)/layout.tsx, mirrored for the employee
// shell -- see that file's comment for the measured reasoning. `displayName`
// stays the same hardcoded "Karyawan" every page already used (no real name
// endpoint was wired into the shell before; keeping it identical here is
// intentional -- this refactor changes WHERE the shell mounts, not what it
// renders). Root `/` (employee login) sits outside this route group.
export default async function EmployeeLayout({ children }: { children: ReactNode }) {
  const session = await getEmployeeSession();
  if (!session) redirect("/");

  return <EmployeeShell displayName="Karyawan">{children}</EmployeeShell>;
}
