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

export async function PUT(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Nama dan tanggal mulai wajib diisi" }, { status: 400 });
  }
  try {
    const updated = await adminApi.put(`/api/v1/admin/holidays/company/${id}`, parsed.data);
    return NextResponse.json(updated);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menyimpan libur perusahaan" }, { status: 502 });
  }
}

export async function DELETE(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    await adminApi.delete(`/api/v1/admin/holidays/company/${id}`);
    return new NextResponse(null, { status: 204 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menghapus libur perusahaan" }, { status: 502 });
  }
}
