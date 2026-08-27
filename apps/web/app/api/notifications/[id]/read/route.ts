import { NextRequest, NextResponse } from "next/server";
import { selfApi, audienceFromRequest, ApiError } from "@/lib/authedApi";

export async function PATCH(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  try {
    await selfApi.patch(audienceFromRequest(req), `/api/v1/notifications/${id}/read`, {});
    return new NextResponse(null, { status: 204 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menandai dibaca" }, { status: 502 });
  }
}
