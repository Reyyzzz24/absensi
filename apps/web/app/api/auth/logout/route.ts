import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import {
  ADMIN_ACCESS_COOKIE,
  ADMIN_REFRESH_COOKIE,
  EMPLOYEE_ACCESS_COOKIE,
  EMPLOYEE_REFRESH_COOKIE,
} from "@/lib/session";

const API_INTERNAL_URL = process.env.API_INTERNAL_URL ?? "http://localhost:8080";

// Revokes the refresh token server-side (D-24) before clearing cookies, so
// the session can't be resurrected via /auth/refresh even though the JWTs
// themselves remain valid until they expire naturally. Best-effort: if the
// API call fails, cookies are still cleared -- a client that can't reach the
// API to revoke shouldn't be stuck unable to log out locally.
export async function POST(req: NextRequest) {
  const audience = req.nextUrl.searchParams.get("aud");
  const isAdmin = audience === "admin";
  const refreshCookie = isAdmin ? ADMIN_REFRESH_COOKIE : EMPLOYEE_REFRESH_COOKIE;

  const store = await cookies();
  const refreshToken = store.get(refreshCookie)?.value;

  if (refreshToken) {
    try {
      await fetch(`${API_INTERNAL_URL}/api/v1/auth/logout`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
    } catch {
      // Best-effort -- see comment above.
    }
  }

  const res = NextResponse.json({ ok: true });

  if (isAdmin) {
    res.cookies.delete(ADMIN_ACCESS_COOKIE);
    res.cookies.delete(ADMIN_REFRESH_COOKIE);
  } else {
    res.cookies.delete(EMPLOYEE_ACCESS_COOKIE);
    res.cookies.delete(EMPLOYEE_REFRESH_COOKIE);
  }

  return res;
}
