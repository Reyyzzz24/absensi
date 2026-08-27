"use client";

import { useRef, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Camera, Building2 } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useApiMutation } from "@/lib/useApiMutation";
import type { CompanySettings } from "@absensi-next/contracts";

const schema = z.object({ name: z.string().min(1, "Nama perusahaan wajib diisi") });
type FormValues = z.infer<typeof schema>;

export function CompanySettingsForm({ company }: { company: CompanySettings }) {
  const [logoBust, setLogoBust] = useState(0);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const mutation = useApiMutation<FormValues>("/api/admin/company", "PUT");
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { name: company.name } });

  function onSubmit(values: FormValues) {
    mutation.mutate(values, {
      onSuccess: () => toast.success("Profil perusahaan tersimpan"),
      onError: (err) => toast.error(err.message),
    });
  }

  function onLogoSelected(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!file) return;
    if (file.size > 2 * 1024 * 1024) {
      toast.error("Ukuran logo maksimal 2MB");
      return;
    }
    if (!["image/png", "image/jpeg", "image/webp"].includes(file.type)) {
      toast.error("Format logo harus PNG, JPEG, atau WebP");
      return;
    }
    setUploading(true);
    const reader = new FileReader();
    reader.onload = async () => {
      const base64 = (reader.result as string).split(",")[1] ?? "";
      try {
        const res = await fetch("/api/admin/company/logo", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ photo: base64 }),
        });
        if (!res.ok) {
          const data = await res.json().catch(() => ({}));
          throw new Error(data.error ?? "Gagal mengunggah logo");
        }
        setLogoBust((n) => n + 1);
        toast.success("Logo perusahaan diperbarui");
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "Gagal mengunggah logo");
      } finally {
        setUploading(false);
      }
    };
    reader.readAsDataURL(file);
  }

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
      <div className="flex items-center gap-4">
        <div className="relative">
          <Avatar size="lg" className="size-14 rounded-lg">
            {company.logo_path && <AvatarImage src={`/api/admin/company/logo?v=${logoBust}`} alt={company.name} />}
            <AvatarFallback className="rounded-lg bg-primary/10 text-primary">
              <Building2 className="size-6" />
            </AvatarFallback>
          </Avatar>
          <button
            type="button"
            onClick={() => fileInputRef.current?.click()}
            disabled={uploading}
            className="absolute -right-1 -bottom-1 flex size-6 items-center justify-center rounded-full bg-primary text-primary-foreground shadow ring-2 ring-white transition hover:brightness-110 disabled:opacity-50"
            aria-label="Ganti logo perusahaan"
          >
            <Camera className="size-3.5" />
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/png,image/jpeg,image/webp"
            className="hidden"
            onChange={onLogoSelected}
          />
        </div>
        <div className="flex-1 space-y-1.5">
          <Label htmlFor="company-name">Nama perusahaan</Label>
          <Input id="company-name" {...form.register("name")} />
          {form.formState.errors.name && <p className="text-sm text-red-600">{form.formState.errors.name.message}</p>}
        </div>
      </div>
      <Button type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? "Menyimpan…" : "Simpan"}
      </Button>
    </form>
  );
}
