"use client";

import { useEffect, useState } from "react";
import { MapPin, MapPinOff, MapPinCheck } from "lucide-react";
import { cn } from "@/lib/utils";

type State = "checking" | "granted" | "denied" | "prompt" | "unsupported";

// Reads the browser's current geolocation PERMISSION state without
// prompting for it -- this only reports "is the browser ready to share your
// location", never "are you inside the office radius". The actual radius
// check (D-1, 100m default from office_locations) only happens server-side
// at check-in submit time: the employee app has no endpoint exposing office
// coordinates/radius to compare against client-side, so a real "dalam
// radius kantor" claim here would be fabricated. The /checkin page's own
// flow (CheckInForm) is what surfaces the real accept/reject outcome.
export function LocationStatus() {
  const [state, setState] = useState<State>("checking");

  useEffect(() => {
    if (!("permissions" in navigator) || !("geolocation" in navigator)) {
      if (!("geolocation" in navigator)) {
        setState("unsupported");
        return;
      }
      // No Permissions API (older Safari) -- can't read state without
      // prompting, so don't guess.
      setState("prompt");
      return;
    }
    let cancelled = false;
    navigator.permissions
      .query({ name: "geolocation" })
      .then((status) => {
        if (cancelled) return;
        setState(status.state as State);
        status.onchange = () => setState(status.state as State);
      })
      .catch(() => setState("prompt"));
    return () => {
      cancelled = true;
    };
  }, []);

  const config: Record<State, { label: string; icon: typeof MapPin; className: string }> = {
    checking: { label: "Memeriksa izin lokasi…", icon: MapPin, className: "text-muted-foreground" },
    granted: { label: "Lokasi siap digunakan", icon: MapPinCheck, className: "text-status-hadir" },
    prompt: { label: "Izin lokasi akan diminta saat check-in", icon: MapPin, className: "text-muted-foreground" },
    denied: { label: "Akses lokasi ditolak -- aktifkan di pengaturan browser", icon: MapPinOff, className: "text-status-alpha" },
    unsupported: { label: "Perangkat tidak mendukung geolocation", icon: MapPinOff, className: "text-status-alpha" },
  };
  const { label, icon: Icon, className } = config[state];

  return (
    <div className={cn("flex items-center gap-1.5 text-xs font-medium", className)}>
      <Icon className="size-3.5 shrink-0" />
      {label}
    </div>
  );
}
