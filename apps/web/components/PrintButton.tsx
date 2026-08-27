"use client";

import { Printer } from "lucide-react";
import { Button } from "@/components/ui/button";

// Plain window.print() -- no library. Page-specific @media print rules live
// next to the markup they affect (RecapTable.tsx, this button's own
// print:hidden, and the shell's print:hidden on sidebar/topbar); the only
// truly global piece is the @page landscape rule in globals.css.
export function PrintButton() {
  return (
    <Button variant="outline" size="sm" onClick={() => window.print()} className="print:hidden">
      <Printer />
      Print
    </Button>
  );
}
