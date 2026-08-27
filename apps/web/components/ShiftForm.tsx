"use client";

import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { useApiMutation } from "@/lib/useApiMutation";

const schema = z.object({
  code: z.string().min(1, "Kode wajib diisi"),
  name: z.string().min(1, "Nama wajib diisi"),
  is_day_off: z.boolean(),
  start_time: z.string().optional(),
  end_time: z.string().optional(),
  late_grace_minutes: z.coerce.number().int().min(0),
});

type FormValues = z.infer<typeof schema>;

export function ShiftForm() {
  const mutation = useApiMutation<
    FormValues & { start_time?: string; end_time?: string }
  >("/api/admin/shifts");

  const {
    register,
    handleSubmit,
    watch,
    reset,
    control,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { is_day_off: false, late_grace_minutes: 15 },
  });

  const isDayOff = watch("is_day_off");

  function onSubmit(values: FormValues) {
    mutation.mutate(
      {
        ...values,
        start_time: values.is_day_off ? undefined : withSeconds(values.start_time),
        end_time: values.is_day_off ? undefined : withSeconds(values.end_time),
      },
      {
        onSuccess: () =>
          reset({ code: "", name: "", is_day_off: false, start_time: "", end_time: "", late_grace_minutes: 15 }),
      },
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-3" noValidate>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="code">Kode</Label>
          <Input id="code" type="text" placeholder="SH01" {...register("code")} />
          {errors.code && <p className="text-sm text-red-600">{errors.code.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="name">Nama</Label>
          <Input id="name" type="text" placeholder="Reguler 09-17" {...register("name")} />
          {errors.name && <p className="text-sm text-red-600">{errors.name.message}</p>}
        </div>
      </div>

      <Controller
        name="is_day_off"
        control={control}
        render={({ field }) => (
          <label className="flex items-center gap-2 text-sm text-slate-700">
            <Checkbox checked={field.value} onCheckedChange={field.onChange} />
            Hari libur (tanpa jam masuk/pulang)
          </label>
        )}
      />

      {!isDayOff && (
        <div className="grid grid-cols-3 gap-3">
          <div className="space-y-1.5">
            <Label htmlFor="start_time">Jam masuk</Label>
            <Input id="start_time" type="time" {...register("start_time")} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="end_time">Jam pulang</Label>
            <Input id="end_time" type="time" {...register("end_time")} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="late_grace_minutes">Toleransi (menit)</Label>
            <Input id="late_grace_minutes" type="number" {...register("late_grace_minutes")} />
          </div>
        </div>
      )}

      {mutation.isError && (
        <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{mutation.error.message}</div>
      )}

      <Button type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? "Menyimpan…" : "Tambah Shift"}
      </Button>
    </form>
  );
}

// <input type="time"> gives "HH:MM" -- the Go API's shift start_time/end_time
// columns are TIME and the handler expects "HH:MM:SS" strings (see
// internal/usecase/config.ShiftInput / parseTimeOfDay in attendance.go).
function withSeconds(value?: string) {
  if (!value) return undefined;
  return value.length === 5 ? `${value}:00` : value;
}
