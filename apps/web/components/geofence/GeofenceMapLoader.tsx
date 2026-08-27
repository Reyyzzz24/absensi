"use client";

import dynamic from "next/dynamic";
import { Skeleton } from "@/components/ui/skeleton";
import type { GeofenceMapProps } from "@/components/geofence/GeofenceMap";

// Leaflet touches `window` at import time -- it breaks Next.js SSR if the
// module is ever evaluated on the server. `ssr: false` here is what keeps
// it out of the server bundle entirely; every consumer must import
// <GeofenceMap> from THIS file, never directly from GeofenceMap.tsx.
const GeofenceMap = dynamic(() => import("@/components/geofence/GeofenceMap").then((m) => m.GeofenceMap), {
  ssr: false,
  loading: () => <Skeleton className="h-64 w-full rounded-2xl" />,
});

export function GeofenceMapLoader(props: GeofenceMapProps) {
  return <GeofenceMap {...props} />;
}
