"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useApiMutation } from "@/lib/useApiMutation";

const schema = z.object({
  employee_id: z.coerce.number().int().positive("NIK/ID karyawan wajib diisi"),
  work_date: z.string().min(1, "Tanggal wajib diisi"),
  note: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

export function FieldAssignmentForm() {
  const mutation = useApiMutation<FormValues>("/api/admin/field-assignments");

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  function onSubmit(values: FormValues) {
    mutation.mutate(values, {
      onSuccess: () => reset({ employee_id: undefined, work_date: "", note: "" }),
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-3" noValidate>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="employee_id">ID Karyawan</Label>
          <Input id="employee_id" type="number" {...register("employee_id")} />
          {errors.employee_id && <p className="text-sm text-red-600">{errors.employee_id.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="work_date">Tanggal</Label>
          <Input id="work_date" type="date" {...register("work_date")} />
          {errors.work_date && <p className="text-sm text-red-600">{errors.work_date.message}</p>}
        </div>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="note">Catatan (opsional)</Label>
        <Input id="note" type="text" placeholder="Kunjungan klien di lokasi X" {...register("note")} />
      </div>

      {mutation.isError && (
        <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{mutation.error.message}</div>
      )}

      <Button type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? "Menyimpan…" : "Tambah Penugasan"}
      </Button>
    </form>
  );
}
