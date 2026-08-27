import { redirect } from "next/navigation";
import { getEmployeeSession } from "@/lib/session";
import { employeeApi, ApiError } from "@/lib/authedApi";
import type { Task } from "@/lib/types";
import { TaskForm } from "@/components/TaskForm";

function formatDateTime(iso: string) {
  return new Date(iso).toLocaleString("id-ID", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default async function TasksPage() {
  const session = await getEmployeeSession();
  if (!session) redirect("/");

  let tasks: Task[] = [];
  try {
    tasks = await employeeApi.get<Task[]>("/api/v1/tasks");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) redirect("/");
  }

  return (
    <div className="mx-auto max-w-lg space-y-4">
      <h1 className="text-xl font-semibold text-foreground">Task</h1>

      <div className="rounded-2xl border border-border bg-white p-4 shadow-sm">
        <h2 className="text-sm font-medium text-foreground">Tambah task</h2>
        <div className="mt-3">
          <TaskForm />
        </div>
      </div>

      <div className="rounded-2xl border border-border bg-white p-4 shadow-sm">
        <h2 className="text-sm font-medium text-foreground">Daftar task</h2>
        {tasks.length === 0 && <p className="mt-2 text-sm text-muted-foreground">Belum ada task.</p>}
        <ul className="mt-3 divide-y divide-border text-sm">
          {tasks.map((t) => (
            <li key={t.id} className="py-2">
              <div className="flex items-center justify-between">
                <p className="font-medium text-foreground">{t.title}</p>
                <span className="text-xs font-mono text-muted-foreground">{t.status}</span>
              </div>
              <p className="text-muted-foreground">
                {formatDateTime(t.starts_at)}
                {t.ends_at ? ` – ${formatDateTime(t.ends_at)}` : ""}
              </p>
              {t.detail && <p className="mt-1 text-foreground">{t.detail}</p>}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
