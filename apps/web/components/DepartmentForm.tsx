"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useApiMutation } from "@/lib/useApiMutation";

const schema = z.object({
  code: z.string().min(1, "Kode wajib diisi"),
  name: z.string().min(1, "Nama wajib diisi"),
});

type FormValues = z.infer<typeof schema>;

export function DepartmentForm() {
  const mutation = useApiMutation<FormValues>("/api/admin/departments");

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  function onSubmit(values: FormValues) {
    mutation.mutate(values, { onSuccess: () => reset({ code: "", name: "" }) });
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-3" noValidate>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="code">Kode</Label>
          <Input id="code" type="text" placeholder="GA" {...register("code")} />
          {errors.code && <p className="text-sm text-red-600">{errors.code.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="name">Nama</Label>
          <Input id="name" type="text" placeholder="General Affairs" {...register("name")} />
          {errors.name && <p className="text-sm text-red-600">{errors.name.message}</p>}
        </div>
      </div>

      {mutation.isError && (
        <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{mutation.error.message}</div>
      )}

      <Button type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? "Menyimpan…" : "Tambah Departemen"}
      </Button>
    </form>
  );
}
