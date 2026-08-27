import { ArrowDown, ArrowUp } from "lucide-react";
import { cn } from "@/lib/utils";

export function TrendPill({ direction, value }: { direction: "up" | "down"; value: string }) {
  const isUp = direction === "up";
  return (
    <span
      className={cn(
        "inline-flex items-center gap-0.5 rounded-full px-1.5 py-0.5 text-xs font-medium",
        isUp ? "bg-trend-up-bg text-trend-up" : "bg-trend-down-bg text-trend-down",
      )}
    >
      {isUp ? <ArrowUp className="size-3" /> : <ArrowDown className="size-3" />}
      {value}
    </span>
  );
}
