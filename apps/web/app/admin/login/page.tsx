import Link from "next/link";
import { redirect } from "next/navigation";
import { LoginForm } from "@/components/LoginForm";
import { getAdminSession } from "@/lib/session";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";

export default async function AdminLoginPage() {
  const session = await getAdminSession();
  if (session) redirect("/admin/dashboard");

  return (
    <main className="flex min-h-screen items-center justify-center bg-muted/30 px-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-xl">Absensi — Panel Admin</CardTitle>
          <CardDescription>Masuk dengan akun admin/superadmin.</CardDescription>
        </CardHeader>
        <CardContent>
          <LoginForm audience="admin" />

          <div className="mt-6 text-sm text-muted-foreground">
            <Link href="/" className="underline">
              &larr; Login karyawan
            </Link>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
