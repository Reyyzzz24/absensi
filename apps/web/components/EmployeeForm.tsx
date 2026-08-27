"use client";

import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import type { Department } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useApiMutation } from "@/lib/useApiMutation";

const schema = z.object({
  nik: z.string().min(1, "NIK wajib diisi"),
  full_name: z.string().min(1, "Nama wajib diisi"),
  password: z.string().min(6, "Password minimal 6 karakter"),
  department_id: z.string().optional(),
  position: z.string().optional(),
  phone: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

// Legacy KaryawanController::store used a hardcoded "12345" initial password
// for every new employee (LOGIC_SPEC.md §9). The admin sets it explicitly
// here instead -- closes that gap rather than reproducing it (D-3).
export function EmployeeForm({ departments }: { departments: Department[] }) {
  const mutation = useApiMutation<{
    nik: string;
    full_name: string;
    password: string;
    department_id?: number;
    position?: string;
    phone?: string;
  }>("/api/admin/employees");

  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { department_id: "" } });

  function onSubmit(values: FormValues) {
    mutation.mutate(
      {
        nik: values.nik,
        full_name: values.full_name,
        password: values.password,
        department_id: values.department_id ? Number(values.department_id) : undefined,
        position: values.position || undefined,
        phone: values.phone || undefined,
      },
      {
        onSuccess: () =>
          reset({ nik: "", full_name: "", password: "", department_id: "", position: "", phone: "" }),
      },
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-3" noValidate>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="nik">NIK</Label>
          <Input id="nik" type="text" {...register("nik")} />
          {errors.nik && <p className="text-sm text-red-600">{errors.nik.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="full_name">Nama lengkap</Label>
          <Input id="full_name" type="text" {...register("full_name")} />
          {errors.full_name && <p className="text-sm text-red-600">{errors.full_name.message}</p>}
        </div>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="password">Password awal</Label>
        <Input id="password" type="text" {...register("password")} />
        {errors.password && <p className="text-sm text-red-600">{errors.password.message}</p>}
      </div>

      <div className="grid grid-cols-3 gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="department_id">Departemen</Label>
          <Controller
            name="department_id"
            control={control}
            render={({ field }) => (
              <Select
                items={departments.map((d) => ({ value: String(d.id), label: d.code }))}
                value={field.value}
                onValueChange={field.onChange}
              >
                <SelectTrigger id="department_id" className="w-full">
                  <SelectValue placeholder="-" />
                </SelectTrigger>
                <SelectContent>
                  {departments.map((d) => (
                    <SelectItem key={d.id} value={String(d.id)}>
                      {d.code}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="position">Jabatan</Label>
          <Input id="position" type="text" {...register("position")} />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="phone">Telepon</Label>
          <Input id="phone" type="text" {...register("phone")} />
        </div>
      </div>

      {mutation.isError && (
        <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{mutation.error.message}</div>
      )}

      <Button type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? "Menyimpan…" : "Tambah Karyawan"}
      </Button>
    </form>
  );
}
