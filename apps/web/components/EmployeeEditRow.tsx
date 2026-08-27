"use client";

import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import type { Department, Employee } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useApiMutation } from "@/lib/useApiMutation";

const schema = z.object({
  full_name: z.string().min(1, "Nama wajib diisi"),
  password: z.string().min(6, "Password minimal 6 karakter").optional().or(z.literal("")),
  department_id: z.string().optional(),
  position: z.string().optional(),
  phone: z.string().optional(),
  is_active: z.boolean(),
});

type FormValues = z.infer<typeof schema>;

export function EmployeeEditRow({ employee, departments }: { employee: Employee; departments: Department[] }) {
  const [editing, setEditing] = useState(false);
  const mutation = useApiMutation<{
    full_name: string;
    password?: string;
    department_id?: number;
    position?: string;
    phone?: string;
    is_active: boolean;
  }>(`/api/admin/employees/${employee.id}`, "PUT");

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      full_name: employee.full_name,
      password: "",
      department_id: employee.department_id ? String(employee.department_id) : "",
      position: employee.position ?? "",
      phone: employee.phone ?? "",
      is_active: employee.is_active,
    },
  });

  function onSubmit(values: FormValues) {
    mutation.mutate(
      {
        full_name: values.full_name,
        password: values.password || undefined,
        department_id: values.department_id ? Number(values.department_id) : undefined,
        position: values.position || undefined,
        phone: values.phone || undefined,
        is_active: values.is_active,
      },
      { onSuccess: () => setEditing(false) },
    );
  }

  if (!editing) {
    return (
      <li className="py-2">
        <div className="flex items-center justify-between">
          <p className="flex items-center gap-2 font-medium">
            {employee.full_name}
            {!employee.is_active && <Badge variant="secondary">Nonaktif</Badge>}
          </p>
          <div className="flex items-center gap-3">
            <span className="font-mono text-xs text-slate-500">{employee.nik}</span>
            <Button type="button" variant="link" size="sm" className="h-auto p-0" onClick={() => setEditing(true)}>
              Edit
            </Button>
          </div>
        </div>
        <p className="text-slate-500">
          {employee.department?.name ?? "-"}
          {employee.position ? ` · ${employee.position}` : ""}
        </p>
      </li>
    );
  }

  return (
    <li className="py-3">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-2" noValidate>
        <div className="grid grid-cols-2 gap-2">
          <div className="space-y-1">
            <Label className="text-xs">Nama lengkap</Label>
            <Input type="text" {...register("full_name")} />
            {errors.full_name && <p className="text-xs text-red-600">{errors.full_name.message}</p>}
          </div>
          <div className="space-y-1">
            <Label className="text-xs">Password baru (opsional)</Label>
            <Input type="text" {...register("password")} placeholder="Kosongkan jika tidak diganti" />
            {errors.password && <p className="text-xs text-red-600">{errors.password.message}</p>}
          </div>
        </div>

        <div className="grid grid-cols-3 gap-2">
          <div className="space-y-1">
            <Label className="text-xs">Departemen</Label>
            <Controller
              name="department_id"
              control={control}
              render={({ field }) => (
                <Select
                items={departments.map((d) => ({ value: String(d.id), label: d.code }))}
                value={field.value}
                onValueChange={field.onChange}
              >
                  <SelectTrigger className="w-full">
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
          <div className="space-y-1">
            <Label className="text-xs">Jabatan</Label>
            <Input type="text" {...register("position")} />
          </div>
          <div className="space-y-1">
            <Label className="text-xs">Telepon</Label>
            <Input type="text" {...register("phone")} />
          </div>
        </div>

        <Controller
          name="is_active"
          control={control}
          render={({ field }) => (
            <label className="flex items-center gap-2 text-sm text-slate-700">
              <Checkbox checked={field.value} onCheckedChange={field.onChange} />
              Aktif
            </label>
          )}
        />

        {mutation.isError && (
          <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{mutation.error.message}</div>
        )}

        <div className="flex gap-2">
          <Button type="submit" disabled={mutation.isPending} size="sm">
            {mutation.isPending ? "Menyimpan…" : "Simpan"}
          </Button>
          <Button type="button" variant="outline" size="sm" onClick={() => setEditing(false)}>
            Batal
          </Button>
        </div>
      </form>
    </li>
  );
}
