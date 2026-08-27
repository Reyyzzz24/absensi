import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { selfApi, audienceFromRequest, ApiError } from "@/lib/authedApi";
import type { NotificationPreferences } from "@absensi-next/contracts";

export async function GET(req: NextRequest) {
  try {
    const prefs = await selfApi.get<NotificationPreferences>(audienceFromRequest(req), "/api/v1/notification-preferences");
    return NextResponse.json(prefs);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal memuat preferensi" }, { status: 502 });
  }
}

const bodySchema = z.object({ type: z.string().min(1), enabled: z.boolean() });

export async function PUT(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Data tidak valid" }, { status: 400 });
  }
  try {
    await selfApi.put(audienceFromRequest(req), "/api/v1/notification-preferences", parsed.data);
    return new NextResponse(null, { status: 204 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal menyimpan preferensi" }, { status: 502 });
  }
}
