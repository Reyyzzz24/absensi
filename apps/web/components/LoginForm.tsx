"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type Audience = "employee" | "admin";

const schemaByAudience = {
  employee: z.object({
    identifier: z.string().min(1, "NIK wajib diisi"),
    password: z.string().min(1, "Password wajib diisi"),
  }),
  admin: z.object({
    identifier: z.string().min(1, "Username wajib diisi"),
    password: z.string().min(1, "Password wajib diisi"),
  }),
};

type FormValues = z.infer<(typeof schemaByAudience)["employee"]>;

const COPY: Record<
  Audience,
  { identifierLabel: string; identifierField: "nik" | "username"; endpoint: string; redirectTo: string }
> = {
  employee: {
    identifierLabel: "NIK",
    identifierField: "nik",
    endpoint: "/api/auth/employee",
    redirectTo: "/dashboard",
  },
  admin: {
    identifierLabel: "Username",
    identifierField: "username",
    endpoint: "/api/auth/admin",
    redirectTo: "/admin/dashboard",
  },
};

export function LoginForm({ audience }: { audience: Audience }) {
  const router = useRouter();
  const [serverError, setServerError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const copy = COPY[audience];

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schemaByAudience[audience]),
  });

  async function onSubmit(values: FormValues) {
    setServerError(null);
    setSubmitting(true);
    try {
      const res = await fetch(copy.endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          [copy.identifierField]: values.identifier,
          password: values.password,
        }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({ error: "Login gagal" }));
        setServerError(data.error ?? "Login gagal");
        return;
      }

      router.push(copy.redirectTo);
      router.refresh();
    } catch {
      setServerError("Tidak bisa menghubungi server. Coba lagi.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
      <div className="space-y-1.5">
        <Label htmlFor="identifier">{copy.identifierLabel}</Label>
        <Input id="identifier" type="text" autoComplete="username" {...register("identifier")} />
        {errors.identifier && (
          <p className="text-sm text-red-600">{errors.identifier.message}</p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="password">Password</Label>
        <Input id="password" type="password" autoComplete="current-password" {...register("password")} />
        {errors.password && (
          <p className="text-sm text-red-600">{errors.password.message}</p>
        )}
      </div>

      {serverError && (
        <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{serverError}</div>
      )}

      <Button type="submit" disabled={submitting} className="w-full">
        {submitting ? "Memproses…" : "Masuk"}
      </Button>
    </form>
  );
}
