import type { ReactNode } from "react";
import { EmployeeTopBar } from "@/components/employee-shell/EmployeeTopBar";
import { EmployeeBottomTabs } from "@/components/employee-shell/EmployeeBottomTabs";

export function EmployeeShell({ displayName, children }: { displayName: string; children: ReactNode }) {
  return (
    <div className="min-h-screen bg-background">
      <EmployeeTopBar displayName={displayName} />
      {/* pb-20 clears the fixed bottom tab bar on mobile; md:pb-0 once the
          bottom tabs are hidden and the desktop top nav takes over. */}
      <main className="mx-auto max-w-5xl px-4 pt-6 pb-20 sm:px-6 md:pb-10">{children}</main>
      <EmployeeBottomTabs />
    </div>
  );
}
