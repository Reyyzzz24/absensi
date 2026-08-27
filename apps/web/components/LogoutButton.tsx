"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/button";

export function LogoutButton({ audience, redirectTo }: { audience: "employee" | "admin"; redirectTo: string }) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  async function handleLogout() {
    setLoading(true);
    await fetch(`/api/auth/logout?aud=${audience}`, { method: "POST" });
    router.push(redirectTo);
    router.refresh();
  }

  return (
    <Button onClick={handleLogout} disabled={loading} variant="outline" size="sm">
      {loading ? "Keluar…" : "Logout"}
    </Button>
  );
}
