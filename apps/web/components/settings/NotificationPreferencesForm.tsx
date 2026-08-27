"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import type { NotificationPreferences } from "@absensi-next/contracts";
import type { SelfAudience } from "@/lib/authedApi";

// One label per known backend notification type (internal/usecase/notification.KnownTypes).
// Add an entry here whenever a new type is wired up server-side.
const TYPE_LABELS: Record<string, { label: string; description: string }> = {
  leave_status_change: {
    label: "Status pengajuan izin/sakit",
    description: "Diberitahu saat pengajuan Anda disetujui atau ditolak.",
  },
};

async function fetchPreferences(audience: SelfAudience): Promise<NotificationPreferences> {
  const res = await fetch(`/api/notification-preferences?aud=${audience}`);
  if (!res.ok) throw new Error("Gagal memuat preferensi");
  return res.json();
}

export function NotificationPreferencesForm({ audience }: { audience: SelfAudience }) {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ["notification-preferences", audience],
    queryFn: () => fetchPreferences(audience),
  });

  const mutation = useMutation({
    mutationFn: async ({ type, enabled }: { type: string; enabled: boolean }) => {
      const res = await fetch(`/api/notification-preferences?aud=${audience}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ type, enabled }),
      });
      if (!res.ok) throw new Error("Gagal menyimpan preferensi");
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["notification-preferences", audience] });
      toast.success("Preferensi disimpan");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  if (isLoading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
      </div>
    );
  }

  const types = Object.keys(TYPE_LABELS);

  return (
    <div className="divide-y divide-border">
      {types.map((type) => {
        const enabled = data?.[type] ?? true;
        const meta = TYPE_LABELS[type];
        return (
          <label key={type} className="flex items-center justify-between gap-3 py-3 first:pt-0 last:pb-0">
            <div className="min-w-0">
              <p className="text-sm font-medium text-foreground">{meta.label}</p>
              <p className="text-xs text-muted-foreground">{meta.description}</p>
            </div>
            <Checkbox
              checked={enabled}
              onCheckedChange={(checked) => mutation.mutate({ type, enabled: checked === true })}
              disabled={mutation.isPending}
            />
          </label>
        );
      })}
    </div>
  );
}
