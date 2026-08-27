import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { selfApi, audienceFromRequest, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({
  current_password: z.string().min(1),
  new_password: z.string().min(8),
});

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Password baru minimal 8 karakter" }, { status: 400 });
  }
  try {
    await selfApi.post(audienceFromRequest(req), "/api/v1/me/change-password", parsed.data);
    return new NextResponse(null, { status: 204 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal mengganti password" }, { status: 502 });
  }
}
