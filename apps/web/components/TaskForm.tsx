"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useApiMutation } from "@/lib/useApiMutation";

const schema = z.object({
  title: z.string().min(1, "Judul wajib diisi"),
  detail: z.string().optional(),
  starts_at: z.string().min(1, "Jam mulai wajib diisi"),
  ends_at: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

// <input type="datetime-local"> gives "YYYY-MM-DDTHH:mm" with no timezone --
// not valid RFC3339, which is what the Go API's time.Time JSON decoding
// requires. Converted to a full ISO string (browser's local timezone) here
// before it ever leaves the client.
function toISOStringOrUndefined(value?: string) {
  if (!value) return undefined;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? undefined : date.toISOString();
}

export function TaskForm() {
  const mutation = useApiMutation<{
    title: string;
    detail?: string;
    starts_at?: string;
    ends_at?: string;
  }>("/api/tasks");

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  function onSubmit(values: FormValues) {
    mutation.mutate(
      {
        title: values.title,
        detail: values.detail || undefined,
        starts_at: toISOStringOrUndefined(values.starts_at),
        ends_at: toISOStringOrUndefined(values.ends_at),
      },
      { onSuccess: () => reset({ title: "", detail: "", starts_at: "", ends_at: "" }) },
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-3" noValidate>
      <div className="space-y-1.5">
        <Label htmlFor="title">Judul pekerjaan</Label>
        <Input id="title" type="text" {...register("title")} />
        {errors.title && <p className="text-sm text-red-600">{errors.title.message}</p>}
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="starts_at">Mulai</Label>
          <Input id="starts_at" type="datetime-local" {...register("starts_at")} />
          {errors.starts_at && <p className="text-sm text-red-600">{errors.starts_at.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="ends_at">Selesai (opsional)</Label>
          <Input id="ends_at" type="datetime-local" {...register("ends_at")} />
        </div>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="detail">Detail (opsional)</Label>
        <Textarea id="detail" rows={2} {...register("detail")} />
      </div>

      {mutation.isError && (
        <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{mutation.error.message}</div>
      )}

      <Button type="submit" disabled={mutation.isPending}>
        {mutation.isPending ? "Menyimpan…" : "Tambah Task"}
      </Button>
    </form>
  );
}
