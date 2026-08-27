import type { ReactNode } from "react";
import { Card } from "@/components/ui/card";

export function ChartCard({
  title,
  description,
  filter,
  children,
  className,
}: {
  title: string;
  description?: string;
  filter?: ReactNode;
  children: ReactNode;
  className?: string;
}) {
  return (
    <Card className={`gap-4 rounded-2xl border border-border p-5 shadow-sm sm:p-6 ${className ?? ""}`}>
      <div className="flex items-start justify-between gap-3">
        <div>
          <h3 className="text-sm font-semibold text-foreground">{title}</h3>
          {description && <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>}
        </div>
        {filter && <div className="shrink-0">{filter}</div>}
      </div>
      {children}
    </Card>
  );
}
