import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { z } from "zod";
import { adminApi, ApiError } from "@/lib/authedApi";
import { ADMIN_ACCESS_COOKIE } from "@/lib/session";

const API_INTERNAL_URL = process.env.API_INTERNAL_URL ?? "http://localhost:8080";

const bodySchema = z.object({ photo: z.string().min(1) });

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Data tidak valid" }, { status: 400 });
  }
  try {
    const updated = await adminApi.post("/api/v1/admin/config/company/logo", parsed.data);
    return NextResponse.json(updated);
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal mengunggah logo" }, { status: 502 });
  }
}

// <img>-can't-send-Authorization proxy, same pattern as the attendance photo route.
export async function GET() {
  const store = await cookies();
  const token = store.get(ADMIN_ACCESS_COOKIE)?.value;
  if (!token) {
    return NextResponse.json({ error: "Not logged in" }, { status: 401 });
  }
  const apiRes = await fetch(`${API_INTERNAL_URL}/api/v1/admin/config/company/logo`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });
  if (!apiRes.ok) {
    return NextResponse.json({ error: "Logo not found" }, { status: apiRes.status });
  }
  const buffer = await apiRes.arrayBuffer();
  return new NextResponse(buffer, {
    headers: { "Content-Type": apiRes.headers.get("Content-Type") ?? "image/png" },
  });
}
