import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { adminApi, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({
  start_date: z.string().min(1),
  end_date: z.string().optional(),
  name: z.string().min(1),
  type: z.enum(["libur", "cuti_bersama"]).default("libur"),
  note: z.string().optional(),
});

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Nama dan tanggal mulai wajib diisi" }, { status: 400 });
  }
  try {
    // Go API enforces superadmin-only for this mutation (D-7 RBAC).
    const created = await adminApi.post("/api/v1/admin/holidays/company", parsed.data);
    return NextResponse.json(created, { status: 201 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menyimpan libur perusahaan" }, { status: 502 });
  }
}
