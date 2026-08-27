import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { adminApi, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({ decision: z.enum(["approved", "rejected"]) });

export async function POST(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "decision harus 'approved' atau 'rejected'" }, { status: 400 });
  }

  try {
    const result = await adminApi.post(`/api/v1/admin/leave-requests/${id}/review`, parsed.data);
    return NextResponse.json(result);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal memproses pengajuan" }, { status: 502 });
  }
}
