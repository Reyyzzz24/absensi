"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Download } from "lucide-react";
import { Button } from "@/components/ui/button";

// Downloads via fetch+blob (not a bare <a href>) specifically so we can show
// a loading state and a proper error toast if the export fails -- a plain
// link would silently navigate to a JSON error body instead. The server-side
// proxy route re-attaches the admin's httpOnly cookie as a Bearer token
// (browser can't do that itself) and streams back whatever the Go API
// returns, error or file, byte for byte.
export function ExportRecapButton({ year, month }: { year: number; month: number }) {
  const [loading, setLoading] = useState(false);

  async function handleExport() {
    setLoading(true);
    try {
      const res = await fetch(`/api/admin/reports/recap/export?year=${year}&month=${month}`);
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error ?? "Gagal mengekspor laporan");
      }
      const blob = await res.blob();
      const disposition = res.headers.get("Content-Disposition") ?? "";
      const match = /filename="([^"]+)"/.exec(disposition);
      const filename = match?.[1] ?? `rekap-absensi-${month}-${year}.xlsx`;

      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);

      toast.success("Laporan Excel berhasil diunduh");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Gagal mengekspor laporan");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Button variant="outline" size="sm" onClick={handleExport} disabled={loading} className="print:hidden">
      <Download />
      {loading ? "Mengekspor…" : "Export Excel"}
    </Button>
  );
}
