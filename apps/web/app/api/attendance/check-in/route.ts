import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { employeeApi, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({
  latitude: z.number(),
  longitude: z.number(),
  photo: z.string().min(1),
  is_wfh: z.boolean(),
});

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Data lokasi/foto tidak lengkap" }, { status: 400 });
  }

  try {
    const result = await employeeApi.post("/api/v1/attendance/check-in", parsed.data);
    return NextResponse.json(result, { status: 200 });
  } catch (err) {
    if (err instanceof ApiError) {
      // Pass through the Go API's status (422 outside geofence, 409
      // cooldown/already-checked-out) so the client can show the right
      // message instead of a generic failure.
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal check-in" }, { status: 502 });
  }
}
