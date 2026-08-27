import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { employeeApi, ApiError } from "@/lib/authedApi";

// Proxies to the Go API with the caller's httpOnly access-token cookie
// attached -- a client component can't read that cookie itself, so writes
// on behalf of the logged-in employee always go through a route like this.
const bodySchema = z.object({
  type: z.enum(["izin", "sakit"]),
  start_date: z.string().min(1),
  end_date: z.string().min(1),
  reason: z.string().optional(),
});

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Data tidak valid" }, { status: 400 });
  }

  try {
    const created = await employeeApi.post("/api/v1/leave-requests", parsed.data);
    return NextResponse.json(created, { status: 201 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal mengirim pengajuan" }, { status: 502 });
  }
}
