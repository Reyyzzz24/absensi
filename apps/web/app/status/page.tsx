"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

type HealthResponse = {
  status: string;
  db: string;
};

type LoadState =
  | { phase: "loading" }
  | { phase: "ok"; data: HealthResponse }
  | { phase: "error"; message: string };

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export default function StatusPage() {
  const [state, setState] = useState<LoadState>({ phase: "loading" });

  useEffect(() => {
    let cancelled = false;

    fetch(`${API_URL}/health`)
      .then(async (res) => {
        const data = (await res.json()) as HealthResponse;
        if (!cancelled) setState({ phase: "ok", data });
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setState({
            phase: "error",
            message: err instanceof Error ? err.message : String(err),
          });
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <main className="mx-auto max-w-md px-6 py-16">
      <h1 className="text-xl font-semibold">Absensi Next — Dev Stack Health</h1>
      <p className="mt-2 text-sm text-slate-500">
        Checking <code className="rounded bg-slate-100 px-1 py-0.5">{API_URL}/health</code>
      </p>

      {state.phase === "loading" && <p className="mt-4 text-slate-500">Checking connection…</p>}

      {state.phase === "ok" && (
        <div className={`mt-4 ${state.data.db === "ok" ? "text-green-600" : "text-orange-600"}`}>
          <p>API status: {state.data.status}</p>
          <p>Database: {state.data.db}</p>
        </div>
      )}

      {state.phase === "error" && (
        <div className="mt-4 text-red-600">
          <p>Could not reach the API.</p>
          <p>{state.message}</p>
        </div>
      )}

      <Link href="/" className="mt-8 inline-block text-sm text-slate-500 underline">
        &larr; Back to login
      </Link>
    </main>
  );
}
