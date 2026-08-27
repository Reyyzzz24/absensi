"use client";

import { useRef, useState } from "react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Camera, ChevronRight, KeyRound } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ChartCard } from "@/components/dashboard/ChartCard";
import { useApiMutation } from "@/lib/useApiMutation";
import type { Profile } from "@absensi-next/contracts";

function initials(name: string) {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase())
    .join("");
}

const phoneSchema = z.object({ phone: z.string().min(8, "Nomor HP minimal 8 digit").max(20) });
type PhoneValues = z.infer<typeof phoneSchema>;

const READ_ONLY_LABELS: Record<string, string> = {
  name: "Nama",
  identifier_employee: "NIK",
  identifier_admin: "Username",
  email: "Email",
  department: "Departemen",
  position: "Jabatan",
  role: "Peran",
};

export function ProfileForm({ profile, settingsHref }: { profile: Profile; settingsHref: string }) {
  const [avatarBust, setAvatarBust] = useState(0);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);

  const phoneMutation = useApiMutation<PhoneValues>(`/api/me?aud=${profile.audience}`, "PATCH");

  const phoneForm = useForm<PhoneValues>({
    resolver: zodResolver(phoneSchema),
    defaultValues: { phone: profile.phone ?? "" },
  });

  function onSubmitPhone(values: PhoneValues) {
    phoneMutation.mutate(values, {
      onSuccess: () => toast.success("Nomor HP tersimpan"),
      onError: (err) => toast.error(err.message),
    });
  }

  function onAvatarSelected(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!file) return;
    if (file.size > 2 * 1024 * 1024) {
      toast.error("Ukuran foto maksimal 2MB");
      return;
    }
    if (!["image/png", "image/jpeg", "image/webp"].includes(file.type)) {
      toast.error("Format foto harus PNG, JPEG, atau WebP");
      return;
    }
    setUploading(true);
    const reader = new FileReader();
    reader.onload = async () => {
      const dataUrl = reader.result as string;
      const base64 = dataUrl.split(",")[1] ?? "";
      try {
        const res = await fetch(`/api/me/avatar?aud=${profile.audience}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ photo: base64 }),
        });
        if (!res.ok) {
          const data = await res.json().catch(() => ({}));
          throw new Error(data.error ?? "Gagal mengunggah foto");
        }
        setAvatarBust((n) => n + 1);
        toast.success("Foto profil diperbarui");
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "Gagal mengunggah foto");
      } finally {
        setUploading(false);
      }
    };
    reader.readAsDataURL(file);
  }

  const isEmployee = profile.audience === "employee";
  const identifierLabel = isEmployee ? READ_ONLY_LABELS.identifier_employee : READ_ONLY_LABELS.identifier_admin;

  return (
    <div className="space-y-6">
      <ChartCard title="Profil">
        <div className="flex items-center gap-4">
          <div className="relative">
            <Avatar size="lg" className="size-16">
              {profile.photo_path && (
                <AvatarImage src={`/api/me/avatar?aud=${profile.audience}&v=${avatarBust}`} alt={profile.name} />
              )}
              <AvatarFallback className="bg-primary/10 text-base font-semibold text-primary">
                {initials(profile.name)}
              </AvatarFallback>
            </Avatar>
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploading}
              className="absolute -right-1 -bottom-1 flex size-6 items-center justify-center rounded-full bg-primary text-primary-foreground shadow ring-2 ring-white transition hover:brightness-110 disabled:opacity-50"
              aria-label="Ganti foto profil"
            >
              <Camera className="size-3.5" />
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/png,image/jpeg,image/webp"
              className="hidden"
              onChange={onAvatarSelected}
            />
          </div>
          <div className="min-w-0">
            <p className="truncate text-base font-semibold text-foreground">{profile.name}</p>
            <p className="truncate text-sm text-muted-foreground">
              {profile.position ?? profile.role ?? (isEmployee ? "Karyawan" : "Admin")}
            </p>
          </div>
        </div>

        <dl className="mt-5 grid grid-cols-1 gap-3 border-t border-border pt-4 text-sm sm:grid-cols-2">
          <div>
            <dt className="text-xs text-muted-foreground">{identifierLabel}</dt>
            <dd className="mt-0.5 font-medium text-foreground">{profile.identifier}</dd>
          </div>
          {profile.email && (
            <div>
              <dt className="text-xs text-muted-foreground">Email</dt>
              <dd className="mt-0.5 font-medium text-foreground">{profile.email}</dd>
            </div>
          )}
          {profile.department && (
            <div>
              <dt className="text-xs text-muted-foreground">Departemen</dt>
              <dd className="mt-0.5 font-medium text-foreground">{profile.department}</dd>
            </div>
          )}
          {profile.position && (
            <div>
              <dt className="text-xs text-muted-foreground">Jabatan</dt>
              <dd className="mt-0.5 font-medium text-foreground">{profile.position}</dd>
            </div>
          )}
          {profile.role && (
            <div>
              <dt className="text-xs text-muted-foreground">Peran</dt>
              <dd className="mt-0.5 font-medium text-foreground capitalize">{profile.role}</dd>
            </div>
          )}
        </dl>
      </ChartCard>

      <ChartCard title="Nomor HP" description="Satu-satunya kontak yang bisa Anda ubah sendiri di sini.">
        <form onSubmit={phoneForm.handleSubmit(onSubmitPhone)} className="flex flex-col gap-3 sm:flex-row sm:items-end">
          <div className="flex-1 space-y-1.5">
            <Label htmlFor="phone">Nomor HP</Label>
            <Input id="phone" type="tel" placeholder="08xxxxxxxxxx" {...phoneForm.register("phone")} />
            {phoneForm.formState.errors.phone && (
              <p className="text-sm text-red-600">{phoneForm.formState.errors.phone.message}</p>
            )}
          </div>
          <Button type="submit" disabled={phoneMutation.isPending}>
            {phoneMutation.isPending ? "Menyimpan…" : "Simpan"}
          </Button>
        </form>
      </ChartCard>

      <Link
        href={settingsHref}
        className="group flex items-center gap-3 rounded-2xl border border-border bg-white p-4 shadow-sm transition hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      >
        <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
          <KeyRound className="size-4" />
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium text-foreground">Ganti password</p>
          <p className="text-xs text-muted-foreground">Kelola di halaman Pengaturan &gt; Akun</p>
        </div>
        <ChevronRight className="size-4 shrink-0 text-muted-foreground transition group-hover:translate-x-0.5" />
      </Link>
    </div>
  );
}
