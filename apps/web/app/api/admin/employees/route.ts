import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { adminApi, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({
  nik: z.string().min(1),
  full_name: z.string().min(1),
  password: z.string().min(6),
  department_id: z.number().int().positive().optional(),
  position: z.string().optional(),
  phone: z.string().optional(),
});

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Data tidak lengkap (password minimal 6 karakter)" }, { status: 400 });
  }

  try {
    const created = await adminApi.post("/api/v1/admin/employees", parsed.data);
    return NextResponse.json(created, { status: 201 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menyimpan karyawan" }, { status: 502 });
  }
}
