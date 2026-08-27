import { NextRequest, NextResponse } from "next/server";
import { decodeJwt } from "jose";

// Silently refreshes an about-to-expire (or already-expired-but-still-
// refreshable) access token using its paired refresh token, so a user
// mid-session never hits a surprise redirect to the login page just
// because their 15-minute access token ran out. Runs before every matched
// page request; the actual per-request API calls still send whatever
// access token is current by the time they run (see lib/authedApi.ts).
//
// Employee and admin sessions are independent (see lib/session.ts) and
// refreshed independently here too.
const API_INTERNAL_URL = process.env.API_INTERNAL_URL ?? "http://localhost:8080";
const REFRESH_MARGIN_MS = 60_000;

const SESSION_PAIRS = [
  { access: "emp_access_token", refresh: "emp_refresh_token" },
  { access: "adm_access_token", refresh: "adm_refresh_token" },
] as const;

export async function proxy(req: NextRequest) {
  const refreshedAccessTokens: { name: string; value: string; maxAge: number }[] = [];

  for (const pair of SESSION_PAIRS) {
    const accessToken = req.cookies.get(pair.access)?.value;
    const refreshToken = req.cookies.get(pair.refresh)?.value;
    if (!refreshToken) continue;
    if (accessToken && !isExpiringSoon(accessToken)) continue;

    const refreshed = await tryRefresh(refreshToken);
    if (!refreshed) continue;

    // Update the request's cookies too, so Server Components rendered
    // for *this* request see the fresh token via next/headers cookies(),
    // not just the browser on its next request.
    req.cookies.set(pair.access, refreshed.access_token);
    refreshedAccessTokens.push({
      name: pair.access,
      value: refreshed.access_token,
      maxAge: refreshed.expires_in,
    });
  }

  const res = NextResponse.next({ request: req });
  for (const t of refreshedAccessTokens) {
    res.cookies.set(t.name, t.value, {
      httpOnly: true,
      sameSite: "lax",
      path: "/",
      maxAge: t.maxAge,
    });
  }
  return res;
}

function isExpiringSoon(token: string): boolean {
  try {
    const claims = decodeJwt(token) as { exp?: number };
    if (!claims.exp) return true;
    return claims.exp * 1000 - Date.now() < REFRESH_MARGIN_MS;
  } catch {
    return true;
  }
}

async function tryRefresh(refreshToken: string): Promise<{ access_token: string; expires_in: number } | null> {
  try {
    const res = await fetch(`${API_INTERNAL_URL}/api/v1/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    if (!res.ok) return null;
    return res.json();
  } catch {
    // API unreachable -- let the page-level session check handle it
    // (redirects to login), rather than failing the whole request here.
    return null;
  }
}

export const config = {
  matcher: [
    "/dashboard/:path*",
    "/checkin/:path*",
    "/leave/:path*",
    "/tasks/:path*",
    "/admin/:path*",
  ],
};
