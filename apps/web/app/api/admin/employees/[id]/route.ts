import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { adminApi, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({
  full_name: z.string().min(1),
  password: z.string().min(6).optional().or(z.literal("")),
  department_id: z.number().int().positive().optional(),
  position: z.string().optional(),
  phone: z.string().optional(),
  is_active: z.boolean(),
});

export async function PUT(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Data tidak lengkap (password, jika diisi, minimal 6 karakter)" }, { status: 400 });
  }

  try {
    const updated = await adminApi.put(`/api/v1/admin/employees/${id}`, parsed.data);
    return NextResponse.json(updated);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal memperbarui karyawan" }, { status: 502 });
  }
}
