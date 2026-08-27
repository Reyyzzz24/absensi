"use client";

import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useApiMutation } from "@/lib/useApiMutation";

const schema = z
  .object({
    type: z.enum(["izin", "sakit"]),
    start_date: z.string().min(1, "Tanggal mulai wajib diisi"),
    end_date: z.string().min(1, "Tanggal selesai wajib diisi"),
    reason: z.string().optional(),
  })
  .refine((data) => data.end_date >= data.start_date, {
    message: "Tanggal selesai tidak boleh sebelum tanggal mulai",
    path: ["end_date"],
  });

type FormValues = z.infer<typeof schema>;

export function LeaveRequestForm() {
  const mutation = useApiMutation<FormValues>("/api/leave-requests");

  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { type: "izin" },
  });

  function onSubmit(values: FormValues) {
    mutation.mutate(values, {
      onSuccess: () => reset({ type: "izin", start_date: "", end_date: "", reason: "" }),
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-3" noValidate>
      <div className="space-y-1.5">
        <Label htmlFor="type">Jenis</Label>
        <Controller
          name="type"
          control={control}
          render={({ field }) => (
            <Select
              items={[
                { value: "izin", label: "Izin" },
                { value: "sakit", label: "Sakit" },
              ]}
              value={field.value}
              onValueChange={field.onChange}
            >
              <SelectTrigger id="type" className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="izin">Izin</SelectItem>
                <SelectItem value="sakit">Sakit</SelectItem>
              </SelectContent>
            </Select>
          )}
        />
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="start_date">Mulai</Label>
          <Input id="start_date" type="date" {...register("start_date")} />
          {errors.start_date && <p className="text-sm text-red-600">{errors.start_date.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="end_date">Selesai</Label>
          <Input id="end_date" type="date" {...register("end_date")} />
          {errors.end_date && <p className="text-sm text-red-600">{errors.end_date.message}</p>}
        </div>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="reason">Keterangan (opsional)</Label>
        <Textarea id="reason" rows={2} {...register("reason")} />
      </div>

      {mutation.isError && (
        <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{mutation.error.message}</div>
      )}

      <Button type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? "Mengirim…" : "Ajukan"}
      </Button>
    </form>
  );
}
