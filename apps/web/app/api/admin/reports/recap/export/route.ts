import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { ADMIN_ACCESS_COOKIE } from "@/lib/session";

const API_INTERNAL_URL = process.env.API_INTERNAL_URL ?? "http://localhost:8080";

// Same "proxy through the server because the browser can't attach a Bearer
// token to a plain download" pattern as the attendance/avatar photo routes --
// except here the response is streamed back to a client fetch() (not an
// <img src>), so ExportRecapButton can show a loading/error state around it.
export async function GET(req: NextRequest) {
  const year = req.nextUrl.searchParams.get("year");
  const month = req.nextUrl.searchParams.get("month");
  const employeeId = req.nextUrl.searchParams.get("employee_id");
  if (!year || !month) {
    return NextResponse.json({ error: "year and month are required" }, { status: 400 });
  }

  const store = await cookies();
  const token = store.get(ADMIN_ACCESS_COOKIE)?.value;
  if (!token) {
    return NextResponse.json({ error: "Not logged in" }, { status: 401 });
  }

  const qs = new URLSearchParams({ year, month });
  if (employeeId) qs.set("employee_id", employeeId);

  const apiRes = await fetch(`${API_INTERNAL_URL}/api/v1/admin/reports/recap/export?${qs.toString()}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });

  if (!apiRes.ok) {
    const data = await apiRes.json().catch(() => ({ error: "Gagal mengekspor laporan" }));
    return NextResponse.json({ error: data.error ?? "Gagal mengekspor laporan" }, { status: apiRes.status });
  }

  const buffer = await apiRes.arrayBuffer();
  return new NextResponse(buffer, {
    headers: {
      "Content-Type":
        apiRes.headers.get("Content-Type") ?? "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
      "Content-Disposition": apiRes.headers.get("Content-Disposition") ?? "attachment",
    },
  });
}
