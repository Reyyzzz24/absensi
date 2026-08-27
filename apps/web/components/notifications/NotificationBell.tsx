"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Bell } from "lucide-react";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import type { Notification } from "@absensi-next/contracts";
import type { SelfAudience } from "@/lib/authedApi";

// MVP polling, per the approved plan (websocket/push deferred). The
// unread-count query polls on this interval regardless of whether the
// panel is open; the list itself only fetches while the panel is open, and
// again on every open (staleTime: 0 default), so it's never stale from a
// prior open. Swapping in a push transport later only means replacing this
// polling query with a subscription -- the list/mark-read/mark-all-read
// wiring below does not need to change.
const UNREAD_POLL_MS = 45_000;

async function fetchUnreadCount(audience: SelfAudience): Promise<number> {
  const res = await fetch(`/api/notifications/unread-count?aud=${audience}`);
  if (!res.ok) return 0;
  const data = await res.json();
  return data.unread_count ?? 0;
}

async function fetchNotifications(audience: SelfAudience): Promise<Notification[]> {
  const res = await fetch(`/api/notifications?aud=${audience}`);
  if (!res.ok) return [];
  return res.json();
}

function timeAgo(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime();
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) return "Baru saja";
  if (minutes < 60) return `${minutes} menit lalu`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} jam lalu`;
  const days = Math.floor(hours / 24);
  return `${days} hari lalu`;
}

export function NotificationBell({ audience }: { audience: SelfAudience }) {
  const [open, setOpen] = useState(false);
  const router = useRouter();
  const queryClient = useQueryClient();

  const { data: unreadCount = 0 } = useQuery({
    queryKey: ["notifications", audience, "unread-count"],
    queryFn: () => fetchUnreadCount(audience),
    refetchInterval: UNREAD_POLL_MS,
  });

  const { data: notifications = [], isLoading } = useQuery({
    queryKey: ["notifications", audience, "list"],
    queryFn: () => fetchNotifications(audience),
    enabled: open,
  });

  function invalidate() {
    queryClient.invalidateQueries({ queryKey: ["notifications", audience] });
  }

  const markReadMutation = useMutation({
    mutationFn: async (id: number) => {
      await fetch(`/api/notifications/${id}/read?aud=${audience}`, { method: "PATCH" });
    },
    onSuccess: invalidate,
  });

  const markAllReadMutation = useMutation({
    mutationFn: async () => {
      await fetch(`/api/notifications/read-all?aud=${audience}`, { method: "PATCH" });
    },
    onSuccess: invalidate,
  });

  function handleItemClick(n: Notification) {
    if (!n.read_at) markReadMutation.mutate(n.id);
    setOpen(false);
    if (n.link) router.push(n.link);
  }

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger
        className="relative flex size-9 items-center justify-center rounded-lg text-muted-foreground outline-none transition hover:bg-secondary hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        aria-label={unreadCount > 0 ? `Notifikasi, ${unreadCount} belum dibaca` : "Notifikasi"}
      >
        <Bell className="size-[18px]" />
        {unreadCount > 0 && (
          <span className="absolute top-1 right-1 flex size-4 items-center justify-center rounded-full bg-red-500 text-[10px] font-semibold text-white">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        )}
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-80 p-0">
        <div className="flex items-center justify-between border-b border-border px-3 py-2">
          <p className="text-sm font-semibold text-foreground">Notifikasi</p>
          {unreadCount > 0 && (
            <button
              type="button"
              onClick={() => markAllReadMutation.mutate()}
              className="text-xs font-medium text-primary hover:underline"
            >
              Tandai semua dibaca
            </button>
          )}
        </div>
        <div className="max-h-80 overflow-y-auto">
          {isLoading && <p className="p-4 text-center text-sm text-muted-foreground">Memuat…</p>}
          {!isLoading && notifications.length === 0 && (
            <p className="p-4 text-center text-sm text-muted-foreground">Belum ada notifikasi.</p>
          )}
          {notifications.map((n) => (
            <button
              key={n.id}
              type="button"
              onClick={() => handleItemClick(n)}
              className={cn(
                "flex w-full flex-col gap-0.5 border-b border-border px-3 py-2.5 text-left transition last:border-0 hover:bg-secondary",
                !n.read_at && "bg-primary/5",
              )}
            >
              <span className="flex items-center gap-2 text-sm font-medium text-foreground">
                {!n.read_at && <span className="size-1.5 shrink-0 rounded-full bg-primary" />}
                {n.title}
              </span>
              {n.body && <span className="text-xs text-muted-foreground">{n.body}</span>}
              <span className="text-[11px] text-muted-foreground/70">{timeAgo(n.created_at)}</span>
            </button>
          ))}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
