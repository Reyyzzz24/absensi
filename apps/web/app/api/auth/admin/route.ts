import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { ApiError, loginAdmin } from "@/lib/api";
import { ADMIN_ACCESS_COOKIE, ADMIN_REFRESH_COOKIE } from "@/lib/session";

const bodySchema = z.object({
  username: z.string().min(1),
  password: z.string().min(1),
});

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "username and password are required" }, { status: 400 });
  }

  try {
    const tokens = await loginAdmin(parsed.data.username, parsed.data.password);

    const res = NextResponse.json({ ok: true });
    res.cookies.set(ADMIN_ACCESS_COOKIE, tokens.access_token, {
      httpOnly: true,
      sameSite: "lax",
      path: "/",
      maxAge: tokens.expires_in,
    });
    res.cookies.set(ADMIN_REFRESH_COOKIE, tokens.refresh_token, {
      httpOnly: true,
      sameSite: "lax",
      path: "/",
      maxAge: 30 * 24 * 60 * 60,
    });
    return res;
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Login failed" }, { status: 502 });
  }
}
