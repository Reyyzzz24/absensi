import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { ADMIN_ACCESS_COOKIE } from "@/lib/session";

const API_INTERNAL_URL = process.env.API_INTERNAL_URL ?? "http://localhost:8080";

// A plain <img src="..."> can't attach an Authorization header, and the
// httpOnly admin cookie lives on the Next.js origin, not the Go API's --
// so this proxies the image through the server, attaching the admin's
// token from the cookie the browser can't read directly.
export async function GET(req: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const type = req.nextUrl.searchParams.get("type") ?? "in";

  const store = await cookies();
  const token = store.get(ADMIN_ACCESS_COOKIE)?.value;
  if (!token) {
    return NextResponse.json({ error: "Not logged in" }, { status: 401 });
  }

  const apiRes = await fetch(`${API_INTERNAL_URL}/api/v1/admin/attendance/${id}/photo?type=${type}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });

  if (!apiRes.ok) {
    return NextResponse.json({ error: "Photo not found" }, { status: apiRes.status });
  }

  const buffer = await apiRes.arrayBuffer();
  return new NextResponse(buffer, {
    headers: { "Content-Type": apiRes.headers.get("Content-Type") ?? "image/png" },
  });
}
