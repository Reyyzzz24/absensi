import { NextRequest, NextResponse } from "next/server";
import { selfApi, audienceFromRequest, ApiError } from "@/lib/authedApi";

export async function GET(req: NextRequest) {
  try {
    const data = await selfApi.get<{ unread_count: number }>(audienceFromRequest(req), "/api/v1/notifications/unread-count");
    return NextResponse.json(data);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal memuat jumlah notifikasi" }, { status: 502 });
  }
}
