import type { LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils";

const TONES = {
  blue: "bg-[#3B82F6]/10 text-[#3B82F6]",
  teal: "bg-[#10B981]/10 text-[#10B981]",
  violet: "bg-[#8B5CF6]/10 text-[#8B5CF6]",
  orange: "bg-[#F97316]/10 text-[#F97316]",
  amber: "bg-[#F59E0B]/10 text-[#F59E0B]",
} as const;

export type IconTileTone = keyof typeof TONES;

export function IconTile({
  icon: Icon,
  tone = "blue",
  className,
}: {
  icon: LucideIcon;
  tone?: IconTileTone;
  className?: string;
}) {
  return (
    <div className={cn("flex size-11 shrink-0 items-center justify-center rounded-xl", TONES[tone], className)}>
      <Icon className="size-5" />
    </div>
  );
}
