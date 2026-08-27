import Link from "next/link";
import { redirect } from "next/navigation";
import { LoginForm } from "@/components/LoginForm";
import { getEmployeeSession } from "@/lib/session";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";

export default async function EmployeeLoginPage() {
  const session = await getEmployeeSession();
  if (session) redirect("/dashboard");

  return (
    <main className="flex min-h-screen items-center justify-center px-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-xl">Absensi — Login Karyawan</CardTitle>
          <CardDescription>Masuk dengan NIK dan password Anda.</CardDescription>
        </CardHeader>
        <CardContent>
          <LoginForm audience="employee" />

          <div className="mt-6 flex justify-between text-sm text-muted-foreground">
            <Link href="/admin/login" className="underline">
              Login sebagai Admin
            </Link>
            <Link href="/status" className="underline">
              Cek status server
            </Link>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
