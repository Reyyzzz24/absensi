"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import { useApiMutation } from "@/lib/useApiMutation";
import type { CompanySettings } from "@absensi-next/contracts";

// ISO weekday numbers (1=Monday..7=Sunday) -- matches the resolver
// (services/api/internal/usecase/holiday) exactly, not JS's Sunday=0.
const WEEKDAYS: { iso: number; label: string }[] = [
  { iso: 1, label: "Senin" },
  { iso: 2, label: "Selasa" },
  { iso: 3, label: "Rabu" },
  { iso: 4, label: "Kamis" },
  { iso: 5, label: "Jumat" },
  { iso: 6, label: "Sabtu" },
  { iso: 7, label: "Minggu" },
];

export function WorkingWeekdaysForm({ company }: { company: CompanySettings }) {
  const [selected, setSelected] = useState(new Set(company.working_weekdays));
  const mutation = useApiMutation<{ working_weekdays: number[] }, CompanySettings>(
    "/api/admin/config/working-weekdays",
    "PUT",
  );

  function toggle(iso: number) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(iso)) next.delete(iso);
      else next.add(iso);
      return next;
    });
  }

  function handleSave() {
    if (selected.size === 0) {
      toast.error("Minimal satu hari kerja harus dipilih");
      return;
    }
    mutation.mutate(
      { working_weekdays: Array.from(selected).sort((a, b) => a - b) },
      {
        onSuccess: () => toast.success("Hari kerja tersimpan"),
        onError: (err) => toast.error(err.message),
      },
    );
  }

  return (
    <div className="space-y-3">
      <p className="text-sm text-muted-foreground">
        Hari yang TIDAK dicentang otomatis dianggap akhir pekan/libur di seluruh laporan dan validasi absensi --
        tidak harus Sabtu-Minggu, mendukung skema 6 hari kerja.
      </p>
      <div className="flex flex-wrap gap-3">
        {WEEKDAYS.map((d) => (
          <label key={d.iso} className="flex items-center gap-2 text-sm">
            <Checkbox checked={selected.has(d.iso)} onCheckedChange={() => toggle(d.iso)} />
            {d.label}
          </label>
        ))}
      </div>
      <Button size="sm" onClick={handleSave} disabled={mutation.isPending}>
        {mutation.isPending ? "Menyimpan…" : "Simpan Hari Kerja"}
      </Button>
    </div>
  );
}
