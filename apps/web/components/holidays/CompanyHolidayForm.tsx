"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useApiMutation } from "@/lib/useApiMutation";

const schema = z
  .object({
    name: z.string().min(1, "Nama wajib diisi"),
    start_date: z.string().min(1, "Tanggal mulai wajib diisi"),
    end_date: z.string().optional(),
    type: z.enum(["libur", "cuti_bersama"]),
    note: z.string().optional(),
  })
  .refine((v) => !v.end_date || v.end_date >= v.start_date, {
    message: "Tanggal selesai tidak boleh sebelum tanggal mulai",
    path: ["end_date"],
  });

type FormValues = z.infer<typeof schema>;

export function CompanyHolidayForm() {
  const mutation = useApiMutation<FormValues>("/api/admin/holidays/company");
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { type: "libur" } });

  function onSubmit(values: FormValues) {
    mutation.mutate(values, {
      onSuccess: () => reset({ name: "", start_date: "", end_date: "", type: "libur", note: "" }),
    });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-3" noValidate>
      <div className="space-y-1.5">
        <Label htmlFor="name">Nama</Label>
        <Input id="name" placeholder="Cuti bersama perusahaan" {...register("name")} />
        {errors.name && <p className="text-sm text-red-600">{errors.name.message}</p>}
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="start_date">Tanggal mulai</Label>
          <Input id="start_date" type="date" {...register("start_date")} />
          {errors.start_date && <p className="text-sm text-red-600">{errors.start_date.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="end_date">Tanggal selesai (opsional, untuk rentang)</Label>
          <Input id="end_date" type="date" {...register("end_date")} />
          {errors.end_date && <p className="text-sm text-red-600">{errors.end_date.message}</p>}
        </div>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="type">Tipe</Label>
        <select
          id="type"
          {...register("type")}
          className="h-9 w-full rounded-lg border border-border bg-white px-3 text-sm shadow-sm focus:ring-2 focus:ring-ring focus:outline-none"
        >
          <option value="libur">Libur perusahaan</option>
          <option value="cuti_bersama">Cuti bersama</option>
        </select>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="note">Catatan (opsional)</Label>
        <Input id="note" placeholder="Alasan/konteks" {...register("note")} />
      </div>

      {mutation.isError && (
        <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{mutation.error.message}</div>
      )}

      <Button type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? "Menyimpan…" : "Tambah Libur"}
      </Button>
    </form>
  );
}
