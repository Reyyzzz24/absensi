import { NextRequest, NextResponse } from "next/server";
import { selfApi, audienceFromRequest, ApiError } from "@/lib/authedApi";
import type { Notification } from "@absensi-next/contracts";

export async function GET(req: NextRequest) {
  const page = req.nextUrl.searchParams.get("page") ?? "1";
  try {
    const list = await selfApi.get<Notification[]>(audienceFromRequest(req), `/api/v1/notifications?page=${page}`);
    return NextResponse.json(list);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal memuat notifikasi" }, { status: 502 });
  }
}
