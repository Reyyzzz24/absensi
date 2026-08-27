"use client";

import { useState } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { CompanyHoliday } from "@absensi-next/contracts";

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" });
}

export function CompanyHolidayList({ holidays }: { holidays: CompanyHoliday[] }) {
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const router = useRouter();

  async function handleDelete(id: number) {
    setDeletingId(id);
    try {
      const res = await fetch(`/api/admin/holidays/company/${id}`, { method: "DELETE" });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error ?? "Gagal menghapus");
      }
      toast.success("Libur perusahaan dihapus");
      router.refresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Gagal menghapus");
    } finally {
      setDeletingId(null);
    }
  }

  if (holidays.length === 0) {
    return <p className="text-sm text-muted-foreground">Belum ada libur perusahaan manual.</p>;
  }

  return (
    <ul className="divide-y divide-border text-sm">
      {holidays.map((h) => (
        <li key={h.id} className="flex items-center justify-between gap-3 py-2">
          <div className="min-w-0">
            <p className="font-medium text-foreground">{h.name}</p>
            <p className="text-xs text-muted-foreground">
              {h.start_date === h.end_date
                ? formatDate(h.start_date)
                : `${formatDate(h.start_date)} – ${formatDate(h.end_date)}`}
              {h.type === "cuti_bersama" && " · Cuti Bersama"}
            </p>
            {h.note && <p className="text-xs text-muted-foreground">{h.note}</p>}
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleDelete(h.id)}
            disabled={deletingId === h.id}
            aria-label={`Hapus ${h.name}`}
            className="shrink-0 text-red-600 hover:bg-red-50 hover:text-red-700"
          >
            <Trash2 className="size-4" />
          </Button>
        </li>
      ))}
    </ul>
  );
}
