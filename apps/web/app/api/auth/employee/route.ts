import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { ApiError, loginEmployee } from "@/lib/api";
import { EMPLOYEE_ACCESS_COOKIE, EMPLOYEE_REFRESH_COOKIE } from "@/lib/session";

const bodySchema = z.object({
  nik: z.string().min(1),
  password: z.string().min(1),
});

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "nik and password are required" }, { status: 400 });
  }

  try {
    const tokens = await loginEmployee(parsed.data.nik, parsed.data.password);

    const res = NextResponse.json({ ok: true });
    // httpOnly cookies per docs/DECISIONS.md §2 (web auth: httpOnly cookie,
    // not localStorage) -- JS on the page can never read these tokens.
    res.cookies.set(EMPLOYEE_ACCESS_COOKIE, tokens.access_token, {
      httpOnly: true,
      sameSite: "lax",
      path: "/",
      maxAge: tokens.expires_in,
    });
    res.cookies.set(EMPLOYEE_REFRESH_COOKIE, tokens.refresh_token, {
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
