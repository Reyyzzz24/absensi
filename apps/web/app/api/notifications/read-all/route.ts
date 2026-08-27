import { NextRequest, NextResponse } from "next/server";
import { selfApi, audienceFromRequest, ApiError } from "@/lib/authedApi";

export async function PATCH(req: NextRequest) {
  try {
    await selfApi.patch(audienceFromRequest(req), "/api/v1/notifications/read-all", {});
    return new NextResponse(null, { status: 204 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menandai semua dibaca" }, { status: 502 });
  }
}
