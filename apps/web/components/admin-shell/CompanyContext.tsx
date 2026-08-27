"use client";

import { createContext, useContext, useState, type ReactNode } from "react";
import type { CompanySettings } from "@absensi-next/contracts";

// Single source of truth for company identity (name + logo) on the client
// side of the admin shell. Seeded once from the server-rendered value
// (AdminDashboardLayout's adminApi.get), then updated directly -- not
// refetched -- whenever CompanySettingsForm's save/upload succeeds, so the
// sidebar workspace card reflects a change the instant it's saved instead
// of waiting for a full page reload. router.refresh() (already triggered by
// useApiMutation) keeps server-rendered readers (settings page itself,
// Laporan's print header, <title>) in sync independently of this context.
const CompanyContext = createContext<{
  company: CompanySettings | null;
  setCompany: (c: CompanySettings) => void;
} | null>(null);

export function CompanyProvider({
  initial,
  children,
}: {
  initial: CompanySettings | null;
  children: ReactNode;
}) {
  const [company, setCompany] = useState(initial);
  return <CompanyContext.Provider value={{ company, setCompany }}>{children}</CompanyContext.Provider>;
}

export function useCompany() {
  const ctx = useContext(CompanyContext);
  if (!ctx) throw new Error("useCompany must be used within a CompanyProvider");
  return ctx;
}
