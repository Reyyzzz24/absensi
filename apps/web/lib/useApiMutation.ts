"use client";

import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

// Shared client-mutation pattern (TanStack Query, D-22): POST/PUT to a
// same-origin Route Handler, then router.refresh() so the owning Server
// Component re-fetches from the Go API. Replaces the submitting/serverError
// useState pair that used to be hand-rolled identically in every form.
export function useApiMutation<TInput>(url: string, method: "POST" | "PUT" | "PATCH" = "POST") {
  const router = useRouter();

  return useMutation<unknown, Error, TInput>({
    mutationFn: async (body: TInput) => {
      const res = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({ error: "Gagal menyimpan" }));
        throw new Error(data.error ?? "Gagal menyimpan");
      }
      if (res.status === 204) return null;
      return res.json().catch(() => null);
    },
    onSuccess: () => router.refresh(),
  });
}
