"use client";

import { useLinkStatus } from "next/link";
import { cn } from "@/lib/utils";

// Must be rendered as a CHILD of next/link's <Link> -- useLinkStatus() only
// reports the pending state of the nearest ancestor <Link>. Gives every nav
// item instant "yes, I heard your click" feedback the moment a transition
// starts, rather than nothing happening until the new route is ready (the
// gap that made clicks feel unresponsive before the layout.tsx fix, and
// still matters for whatever residual latency remains -- e.g. a page's own
// data fetch after the shared shell has already re-rendered).
export function NavPendingDot({ className }: { className?: string }) {
  const { pending } = useLinkStatus();
  if (!pending) return null;
  return (
    <span
      aria-hidden
      className={cn("ml-auto size-1.5 shrink-0 animate-pulse rounded-full bg-current opacity-70", className)}
    />
  );
}
