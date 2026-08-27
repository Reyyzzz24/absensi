import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { employeeApi, ApiError } from "@/lib/authedApi";

const bodySchema = z.object({
  title: z.string().min(1),
  detail: z.string().optional(),
  starts_at: z.string().min(1),
  ends_at: z.string().optional(),
});

export async function POST(req: NextRequest) {
  const parsed = bodySchema.safeParse(await req.json().catch(() => null));
  if (!parsed.success) {
    return NextResponse.json({ error: "Data tidak valid" }, { status: 400 });
  }

  try {
    const created = await employeeApi.post("/api/v1/tasks", parsed.data);
    return NextResponse.json(created, { status: 201 });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json({ error: err.message }, { status: err.status });
    }
    return NextResponse.json({ error: "Gagal membuat task" }, { status: 502 });
  }
}
