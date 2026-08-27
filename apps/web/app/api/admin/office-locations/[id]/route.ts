import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { adminApi, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({
  name: z.string().min(1),
  latitude: z.number(),
  longitude: z.number(),
  radius_meters: z.number().int().positive(),
  is_active: z.boolean(),
});

export async function PUT(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Data tidak valid" }, { status: 400 });
  }

  try {
    // Go API enforces superadmin-only for this mutation (D-7 RBAC).
    const updated = await adminApi.put(`/api/v1/admin/config/office-locations/${id}`, parsed.data);
    return NextResponse.json(updated, { status: 200 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menyimpan lokasi kantor" }, { status: 502 });
  }
}
