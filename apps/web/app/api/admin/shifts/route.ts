import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { adminApi, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({
  code: z.string().min(1),
  name: z.string().min(1),
  is_day_off: z.boolean(),
  start_time: z.string().optional(),
  end_time: z.string().optional(),
  late_grace_minutes: z.number().int().min(0).optional(),
});

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Data tidak valid" }, { status: 400 });
  }

  try {
    // Superadmin-only on the Go API (D-7) -- a plain admin's 403 passes
    // through unchanged.
    const created = await adminApi.post("/api/v1/admin/config/shifts", parsed.data);
    return NextResponse.json(created, { status: 201 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menyimpan shift" }, { status: 502 });
  }
}
