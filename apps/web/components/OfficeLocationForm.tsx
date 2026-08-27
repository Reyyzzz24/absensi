"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useApiMutation } from "@/lib/useApiMutation";

const schema = z.object({
  name: z.string().min(1, "Nama wajib diisi"),
  latitude: z.coerce.number().min(-90).max(90),
  longitude: z.coerce.number().min(-180).max(180),
  radius_meters: z.coerce.number().int().positive(),
});

type FormValues = z.infer<typeof schema>;

export function OfficeLocationForm() {
  const mutation = useApiMutation<FormValues>("/api/admin/office-locations");

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { radius_meters: 100 },
  });

  function onSubmit(values: FormValues) {
    mutation.mutate(values, {
      onSuccess: () => reset({ name: "", latitude: undefined, longitude: undefined, radius_meters: 100 }),
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-3" noValidate>
      <div className="space-y-1.5">
        <Label htmlFor="name">Nama lokasi</Label>
        <Input id="name" type="text" {...register("name")} />
        {errors.name && <p className="text-sm text-red-600">{errors.name.message}</p>}
      </div>

      <div className="grid grid-cols-3 gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="latitude">Latitude</Label>
          <Input id="latitude" type="number" step="any" {...register("latitude")} />
          {errors.latitude && <p className="text-sm text-red-600">{errors.latitude.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="longitude">Longitude</Label>
          <Input id="longitude" type="number" step="any" {...register("longitude")} />
          {errors.longitude && <p className="text-sm text-red-600">{errors.longitude.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="radius_meters">Radius (m)</Label>
          <Input id="radius_meters" type="number" {...register("radius_meters")} />
          {errors.radius_meters && <p className="text-sm text-red-600">{errors.radius_meters.message}</p>}
        </div>
      </div>

      {mutation.isError && (
        <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{mutation.error.message}</div>
      )}

      <Button type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? "Menyimpan…" : "Tambah Lokasi"}
      </Button>
    </form>
  );
}
