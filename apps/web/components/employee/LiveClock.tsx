"use client";

import { useEffect, useState } from "react";

// Renders nothing time-dependent until after mount -- ticking clocks are a
// classic SSR/hydration mismatch source (server's render time will never
// equal the client's). `now` starts null so the first client render still
// matches the server's (empty) output, then a client-only effect starts the
// per-second tick.
export function LiveClock() {
  const [now, setNow] = useState<Date | null>(null);

  useEffect(() => {
    setNow(new Date());
    const id = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(id);
  }, []);

  const time = now
    ? now.toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit", second: "2-digit", timeZone: "Asia/Jakarta" })
    : "--:--:--";
  const date = now
    ? now.toLocaleDateString("id-ID", { weekday: "long", day: "numeric", month: "long", year: "numeric", timeZone: "Asia/Jakarta" })
    : " ";

  return (
    <div>
      <p className="font-mono text-[32px] leading-tight font-bold tracking-tight text-foreground tabular-nums sm:text-4xl">
        {time}
      </p>
      <p className="mt-0.5 text-sm text-muted-foreground">{date}</p>
    </div>
  );
}
