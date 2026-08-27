import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { adminApi, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({
  employee_id: z.number().int().positive(),
  work_date: z.string().min(1),
  note: z.string().optional(),
});

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "employee_id dan work_date wajib diisi" }, { status: 400 });
  }

  try {
    // approved_by is taken server-side from the admin's own token (see Go
    // handler ConfigHandler.CreateFieldAssignment) -- not something the
    // client can spoof.
    const created = await adminApi.post("/api/v1/admin/config/field-assignments", parsed.data);
    return NextResponse.json(created, { status: 201 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menyimpan penugasan dinas luar" }, { status: 502 });
  }
}
