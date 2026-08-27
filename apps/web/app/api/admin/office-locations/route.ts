import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { adminApi, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({
  name: z.string().min(1),
  latitude: z.number(),
  longitude: z.number(),
  radius_meters: z.number().int().positive().optional(),
});

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Data tidak valid" }, { status: 400 });
  }

  try {
    // Go API enforces superadmin-only for this mutation (D-7 RBAC) -- a
    // plain admin gets a 403 straight from there, passed through as-is.
    const created = await adminApi.post("/api/v1/admin/config/office-locations", parsed.data);
    return NextResponse.json(created, { status: 201 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menyimpan lokasi kantor" }, { status: 502 });
  }
}
