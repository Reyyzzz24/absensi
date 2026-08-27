import { NextRequest, NextResponse } from "next/server";
import { adminApi, ApiError } from "@/lib/authedApi";

// Admin-triggered only -- this is the sole place a request to the external
// holiday calendar source happens (D-25), never on the attendance/recap
// request path. On failure the existing cache is left untouched
// server-side; this route just relays whichever outcome the Go API reports.
export async function POST(req: NextRequest) {
  const year = req.nextUrl.searchParams.get("year");
  if (!year) {
    return NextResponse.json({ error: "year wajib diisi" }, { status: 400 });
  }
  try {
    const result = await adminApi.post(`/api/v1/admin/holidays/national/sync?year=${encodeURIComponent(year)}`);
    return NextResponse.json(result);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal sinkronisasi kalender libur nasional" }, { status: 502 });
  }
}
