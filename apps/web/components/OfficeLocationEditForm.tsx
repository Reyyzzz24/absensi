"use client";

import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { GeofenceMapLoader } from "@/components/geofence/GeofenceMapLoader";
import type { OfficeLocation } from "@/lib/types";
import { useApiMutation } from "@/lib/useApiMutation";
import { Search } from "lucide-react";

const schema = z.object({
  name: z.string().min(1, "Nama wajib diisi"),
  latitude: z.coerce.number().min(-90).max(90),
  longitude: z.coerce.number().min(-180).max(180),
  radius_meters: z.coerce.number().int().positive(),
  is_active: z.boolean(),
});

type FormValues = z.infer<typeof schema>;

export function OfficeLocationEditForm({ location }: { location: OfficeLocation }) {
  const mutation = useApiMutation<FormValues>(`/api/admin/office-locations/${location.id}`, "PUT");
  const [addressQuery, setAddressQuery] = useState("");
  const [searching, setSearching] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    reset,
    formState: { errors, isDirty },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: location.name,
      latitude: location.latitude,
      longitude: location.longitude,
      radius_meters: location.radius_meters,
      is_active: location.is_active,
    },
  });

  const latitude = watch("latitude");
  const longitude = watch("longitude");
  const radius = watch("radius_meters");
  const isActive = watch("is_active");

  // Feeds the map's center -- recomputed whenever the lat/lng inputs change,
  // so typing a coordinate moves the pin just like dragging it does.
  const center = useMemo(() => ({ lat: latitude, lng: longitude }), [latitude, longitude]);

  function onSubmit(values: FormValues) {
    mutation.mutate(values, {
      onSuccess: () => {
        reset(values); // clears isDirty so the badge flips back to "Tersimpan"
        toast.success(`Lokasi "${values.name}" tersimpan`);
      },
      onError: (err) => toast.error(err.message),
    });
  }

  // On-demand only (button click, not keystroke-debounced) -- Nominatim's
  // usage policy caps unattended/bulk traffic, and a manual trigger already
  // keeps volume to "one lookup per user action".
  async function searchAddress() {
    if (!addressQuery.trim()) return;
    setSearching(true);
    try {
      const res = await fetch(
        `https://nominatim.openstreetmap.org/search?format=json&limit=1&q=${encodeURIComponent(addressQuery)}`,
        { headers: { "Accept-Language": "id" } },
      );
      const data = await res.json();
      if (Array.isArray(data) && data.length > 0) {
        setValue("latitude", parseFloat(data[0].lat), { shouldDirty: true });
        setValue("longitude", parseFloat(data[0].lon), { shouldDirty: true });
      } else {
        toast.error("Alamat tidak ditemukan");
      }
    } catch {
      toast.error("Pencarian alamat gagal, coba lagi");
    } finally {
      setSearching(false);
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
      <div className="flex items-center justify-between">
        <Input
          className="max-w-xs font-medium"
          {...register("name")}
          aria-label="Nama lokasi"
        />
        <span
          className={
            isDirty
              ? "rounded-full bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700"
              : "rounded-full bg-[var(--status-hadir-bg)] px-2.5 py-1 text-xs font-medium text-[var(--status-hadir)]"
          }
        >
          {isDirty ? "Perubahan belum disimpan" : "Tersimpan"}
        </span>
      </div>
      {errors.name && <p className="text-sm text-red-600">{errors.name.message}</p>}

      <div className="flex gap-2">
        <Input
          type="text"
          placeholder="Cari alamat (opsional)…"
          value={addressQuery}
          onChange={(e) => setAddressQuery(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              searchAddress();
            }
          }}
        />
        <Button type="button" variant="outline" onClick={searchAddress} disabled={searching}>
          <Search className="size-4" />
          {searching ? "Mencari…" : "Cari"}
        </Button>
      </div>

      <GeofenceMapLoader
        mode="edit"
        center={center}
        radius={radius || 0}
        draggable
        heightClassName="h-72"
        onChange={(c) => {
          setValue("latitude", Number(c.lat.toFixed(6)), { shouldDirty: true });
          setValue("longitude", Number(c.lng.toFixed(6)), { shouldDirty: true });
        }}
      />
      <p className="text-xs text-muted-foreground">
        Seret pin atau klik pada peta untuk memindahkan titik lokasi kantor.
      </p>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <div className="space-y-1.5">
          <Label htmlFor={`lat-${location.id}`}>Latitude</Label>
          <Input id={`lat-${location.id}`} type="number" step="any" {...register("latitude")} />
          {errors.latitude && <p className="text-sm text-red-600">{errors.latitude.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor={`lng-${location.id}`}>Longitude</Label>
          <Input id={`lng-${location.id}`} type="number" step="any" {...register("longitude")} />
          {errors.longitude && <p className="text-sm text-red-600">{errors.longitude.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor={`radius-${location.id}`}>Radius (m)</Label>
          <Input id={`radius-${location.id}`} type="number" {...register("radius_meters")} />
          {errors.radius_meters && <p className="text-sm text-red-600">{errors.radius_meters.message}</p>}
        </div>
      </div>

      <input
        type="range"
        min={20}
        max={1000}
        step={10}
        value={radius || 0}
        onChange={(e) => setValue("radius_meters", Number(e.target.value), { shouldDirty: true })}
        className="w-full accent-[var(--primary)]"
        aria-label="Radius geofence (slider)"
      />

      <label className="flex items-center gap-2 text-sm text-slate-700">
        <Checkbox checked={isActive} onCheckedChange={(checked) => setValue("is_active", checked === true, { shouldDirty: true })} />
        Aktif
      </label>

      <Button type="submit" disabled={mutation.isPending || !isDirty}>
        {mutation.isPending ? "Menyimpan…" : "Simpan Perubahan"}
      </Button>
    </form>
  );
}
