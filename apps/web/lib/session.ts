import "server-only";
import { cookies } from "next/headers";
import { decodeJwt } from "jose";

// Separate cookie namespaces per audience so an employee session and an
// admin session can coexist in the same browser, matching the legacy
// system's two independent guards (auth:karyawan / auth:user,
// config/auth.php) rather than collapsing them into one.
export const EMPLOYEE_ACCESS_COOKIE = "emp_access_token";
export const EMPLOYEE_REFRESH_COOKIE = "emp_refresh_token";
export const ADMIN_ACCESS_COOKIE = "adm_access_token";
export const ADMIN_REFRESH_COOKIE = "adm_refresh_token";

export type SessionClaims = {
  sub: string;
  aud_type: "employee" | "admin";
  role?: string;
  exp: number;
};

export async function getEmployeeSession(): Promise<SessionClaims | null> {
  return readSession(EMPLOYEE_ACCESS_COOKIE);
}

export async function getAdminSession(): Promise<SessionClaims | null> {
  return readSession(ADMIN_ACCESS_COOKIE);
}

async function readSession(cookieName: string): Promise<SessionClaims | null> {
  const store = await cookies();
  const token = store.get(cookieName)?.value;
  if (!token) return null;

  try {
    // Decoded, not verified -- the API is the source of truth for
    // authorization on every request; this is only used here to render
    // "logged in as X" and to redirect already-logged-in users away from
    // the login form. Never trust this for access control decisions.
    return decodeJwt(token) as SessionClaims;
  } catch {
    return null;
  }
}
