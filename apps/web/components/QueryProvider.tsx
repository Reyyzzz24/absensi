"use client";

import { useState, type ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

// Client-only mutations go through this cache (TanStack Query, D-22/CLAUDE.md
// §2). Reads stay on the existing Server Component + authedApi pattern --
// this is not a fetch-everything-on-the-client rewrite, just a shared home
// for the useMutation calls that used to hand-roll submitting/error state,
// plus the handful of useQuery calls in NotificationBell/settings forms.
//
// Perf fix (nav-speed audit): this used to be `new QueryClient()` with zero
// options, i.e. staleTime: 0. Every query was marked stale the instant it
// mounted, so every remount (before the layout.tsx fix, that meant every
// single page navigation) fired an immediate background refetch. staleTime
// here doesn't disable refetching -- it just means data already fetched in
// the last 30s renders instantly from cache instead of every mount showing
// a fresh network round-trip; gcTime keeps that cache around for 5 minutes
// after a component unmounts, which is what makes "back" navigation to a
// page you just left feel instant instead of hollow-then-populated.
// refetchOnWindowFocus is off because nothing here is realtime enough to
// justify surprise refetches whenever a user alt-tabs back to the browser.
export function QueryProvider({ children }: { children: ReactNode }) {
  const [client] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 30_000,
            gcTime: 5 * 60_000,
            refetchOnWindowFocus: false,
          },
        },
      }),
  );
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}
