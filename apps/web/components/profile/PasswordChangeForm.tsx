"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useApiMutation } from "@/lib/useApiMutation";
import type { SelfAudience } from "@/lib/authedApi";

const passwordSchema = z
  .object({
    current_password: z.string().min(1, "Password saat ini wajib diisi"),
    new_password: z.string().min(8, "Password baru minimal 8 karakter"),
    confirm_password: z.string().min(1, "Konfirmasi password wajib diisi"),
  })
  .refine((v) => v.new_password === v.confirm_password, {
    message: "Konfirmasi password tidak cocok",
    path: ["confirm_password"],
  });
type PasswordValues = z.infer<typeof passwordSchema>;

// Shared between /profile's "Ganti password" link target and /settings'
// Akun section -- same component, one canonical place it actually lives
// (Settings > Akun), per the approved plan ("ganti password -> reuse
// endpoint di A").
export function PasswordChangeForm({ audience }: { audience: SelfAudience }) {
  const mutation = useApiMutation<Omit<PasswordValues, "confirm_password">>(
    `/api/me/change-password?aud=${audience}`,
    "POST",
  );
  const form = useForm<PasswordValues>({
    resolver: zodResolver(passwordSchema),
    defaultValues: { current_password: "", new_password: "", confirm_password: "" },
  });

  function onSubmit(values: PasswordValues) {
    mutation.mutate(
      { current_password: values.current_password, new_password: values.new_password },
      {
        onSuccess: () => {
          toast.success("Password berhasil diganti");
          form.reset();
        },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-3">
      <div className="space-y-1.5">
        <Label htmlFor="current_password">Password saat ini</Label>
        <Input id="current_password" type="password" {...form.register("current_password")} />
        {form.formState.errors.current_password && (
          <p className="text-sm text-red-600">{form.formState.errors.current_password.message}</p>
        )}
      </div>
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div className="space-y-1.5">
          <Label htmlFor="new_password">Password baru</Label>
          <Input id="new_password" type="password" {...form.register("new_password")} />
          {form.formState.errors.new_password && (
            <p className="text-sm text-red-600">{form.formState.errors.new_password.message}</p>
          )}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="confirm_password">Konfirmasi password baru</Label>
          <Input id="confirm_password" type="password" {...form.register("confirm_password")} />
          {form.formState.errors.confirm_password && (
            <p className="text-sm text-red-600">{form.formState.errors.confirm_password.message}</p>
          )}
        </div>
      </div>
      <Button type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? "Menyimpan…" : "Ganti Password"}
      </Button>
    </form>
  );
}
