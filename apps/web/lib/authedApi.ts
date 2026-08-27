// Server-only authenticated API client. Reads the employee's httpOnly
// access-token cookie (see lib/session.ts) and forwards it as a Bearer
// token to the Go API. Used two ways:
//   1. Directly from Server Components for read-only data on page load
//      (no network round-trip through a Route Handler needed).
//   2. From Route Handlers (app/api/**) that proxy client-side form
//      submissions -- a client component can't read an httpOnly cookie
//      itself, so writes always go through a same-origin Route Handler.
import "server-only";
import { cookies } from "next/headers";
import { ApiError } from "@/lib/api";
import { ADMIN_ACCESS_COOKIE, EMPLOYEE_ACCESS_COOKIE } from "@/lib/session";

const API_INTERNAL_URL = process.env.API_INTERNAL_URL ?? "http://localhost:8080";

async function authedRequest<T>(
  cookieName: string,
  path: string,
  init?: RequestInit,
): Promise<T> {
  const store = await cookies();
  const token = store.get(cookieName)?.value;
  if (!token) {
    throw new ApiError(401, "Not logged in");
  }

  const res = await fetch(`${API_INTERNAL_URL}${path}`, {
    ...init,
    headers: {
      ...(init?.headers ?? {}),
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    cache: "no-store",
  });

  if (!res.ok) {
    const data = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, data.error ?? "Request failed");
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const employeeApi = {
  get: <T>(path: string) => authedRequest<T>(EMPLOYEE_ACCESS_COOKIE, path),
  post: <T>(path: string, body: unknown) =>
    authedRequest<T>(EMPLOYEE_ACCESS_COOKIE, path, { method: "POST", body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    authedRequest<T>(EMPLOYEE_ACCESS_COOKIE, path, { method: "PUT", body: JSON.stringify(body) }),
};

export const adminApi = {
  get: <T>(path: string) => authedRequest<T>(ADMIN_ACCESS_COOKIE, path),
  post: <T>(path: string, body?: unknown) =>
    authedRequest<T>(ADMIN_ACCESS_COOKIE, path, { method: "POST", body: body !== undefined ? JSON.stringify(body) : undefined }),
  put: <T>(path: string, body: unknown) =>
    authedRequest<T>(ADMIN_ACCESS_COOKIE, path, { method: "PUT", body: JSON.stringify(body) }),
  delete: <T>(path: string) => authedRequest<T>(ADMIN_ACCESS_COOKIE, path, { method: "DELETE" }),
};

export type SelfAudience = "employee" | "admin";

// A handful of Go endpoints (self profile, notifications) genuinely accept
// either audience's token (RequireAnyAuth server-side). Both sessions can
// be logged in simultaneously in the same browser (separate cookie
// namespaces, lib/session.ts) -- so the caller MUST say which one it means
// explicitly. Guessing (e.g. "prefer whichever cookie exists") silently
// leaks the wrong account's data whenever both are logged in at once: an
// admin browsing /admin/settings would see their OWN employee notifications
// instead of an empty admin inbox. audience always comes from data the
// route handler already knows for certain (e.g. a `?aud=` query param set
// by the client component that knows which shell it's rendered in).
function cookieFor(audience: SelfAudience) {
  return audience === "employee" ? EMPLOYEE_ACCESS_COOKIE : ADMIN_ACCESS_COOKIE;
}

async function selfRequest<T>(audience: SelfAudience, path: string, init?: RequestInit): Promise<T> {
  const store = await cookies();
  const token = store.get(cookieFor(audience))?.value;
  if (!token) {
    throw new ApiError(401, "Not logged in");
  }

  const res = await fetch(`${API_INTERNAL_URL}${path}`, {
    ...init,
    headers: { ...(init?.headers ?? {}), Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    cache: "no-store",
  });

  if (!res.ok) {
    const data = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, data.error ?? "Request failed");
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const selfApi = {
  get: <T>(audience: SelfAudience, path: string) => selfRequest<T>(audience, path),
  patch: <T>(audience: SelfAudience, path: string, body: unknown) =>
    selfRequest<T>(audience, path, { method: "PATCH", body: JSON.stringify(body) }),
  post: <T>(audience: SelfAudience, path: string, body: unknown) =>
    selfRequest<T>(audience, path, { method: "POST", body: JSON.stringify(body) }),
  put: <T>(audience: SelfAudience, path: string, body: unknown) =>
    selfRequest<T>(audience, path, { method: "PUT", body: JSON.stringify(body) }),
};

// Reads the ?aud= query param a client component sets on every self-* proxy
// request (it always knows which shell it's rendered in), rejecting
// anything else so a missing/garbled param fails loudly instead of
// silently falling back to a guess.
export function audienceFromRequest(req: { nextUrl: URL } | { url: string }): SelfAudience {
  const url = "nextUrl" in req ? req.nextUrl : new URL(req.url);
  const aud = url.searchParams.get("aud");
  if (aud === "employee" || aud === "admin") return aud;
  throw new ApiError(400, "Missing or invalid aud query parameter");
}

export { ApiError };
