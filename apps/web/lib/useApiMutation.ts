"use client";

import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

// Shared client-mutation pattern (TanStack Query, D-22): POST/PUT to a
// same-origin Route Handler, then router.refresh() so the owning Server
// Component re-fetches from the Go API. Replaces the submitting/serverError
// useState pair that used to be hand-rolled identically in every form.
export function useApiMutation<TInput, TOutput = unknown>(url: string, method: "POST" | "PUT" | "PATCH" | "DELETE" = "POST") {
  const router = useRouter();

  return useMutation<TOutput, Error, TInput>({
    mutationFn: async (body: TInput) => {
      const res = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: method === "DELETE" ? undefined : JSON.stringify(body),
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
