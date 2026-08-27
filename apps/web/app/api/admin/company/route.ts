import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { adminApi, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({ name: z.string().min(1) });

export async function PUT(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Nama wajib diisi" }, { status: 400 });
  }
  try {
    // Go API enforces superadmin-only for this mutation (D-7 RBAC).
    const updated = await adminApi.put("/api/v1/admin/config/company", parsed.data);
    return NextResponse.json(updated);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menyimpan profil perusahaan" }, { status: 502 });
  }
}
