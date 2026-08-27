import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { selfApi, audienceFromRequest, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({ phone: z.string() });

export async function PATCH(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Data tidak valid" }, { status: 400 });
  }
  try {
    const updated = await selfApi.patch(audienceFromRequest(req), "/api/v1/me", parsed.data);
    return NextResponse.json(updated);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menyimpan profil" }, { status: 502 });
  }
}
