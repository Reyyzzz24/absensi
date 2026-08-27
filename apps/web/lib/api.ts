// Server-only helper for calling the Go API. Used exclusively from Route
// Handlers (app/api/**) and Server Components -- never imported into a
// "use client" file. Uses the internal Docker service name (API_INTERNAL_URL,
// e.g. http://api:8080), not NEXT_PUBLIC_API_URL (that one is for the
// browser, which can't resolve Docker service names).
import "server-only";

const API_INTERNAL_URL = process.env.API_INTERNAL_URL ?? "http://localhost:8080";

export type Tokens = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
};

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function postJSON<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${API_INTERNAL_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });

  if (!res.ok) {
    const data = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, data.error ?? "Request failed");
  }

  return res.json() as Promise<T>;
}

export function loginEmployee(nik: string, password: string) {
  return postJSON<Tokens>("/api/v1/auth/employee/login", { nik, password });
}

export function loginAdmin(username: string, password: string) {
  return postJSON<Tokens>("/api/v1/auth/admin/login", { username, password });
}
