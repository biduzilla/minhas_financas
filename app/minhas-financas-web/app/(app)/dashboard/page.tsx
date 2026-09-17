import { ACCESS_COOKIE } from "@/lib/auth/cookies";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { LogoutButton } from "./logout-button";

export default async function DashboardPage() {
    const jar = await cookies()
    const token = jar.get(ACCESS_COOKIE)?.value

    if (!token) {
        redirect('/login')
    }

    return (
        <main className="p-8">
            <h1 className="text-2xl font-bold">Dashboard</h1>
            <p className="mt-2 text-sm text-gray-600">
                Você está autenticado. 🎉
            </p>
            <div className="mt-6">
                <LogoutButton />
            </div>
        </main>
    );
}