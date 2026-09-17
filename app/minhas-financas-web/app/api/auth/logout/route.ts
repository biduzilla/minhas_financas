import { logout } from "@/lib/api/endpoints/auth";
import { ACCESS_COOKIE, REFRESH_COOKIE } from "@/lib/auth/cookies";
import { cookies } from "next/headers";

export async function POST() {
    const jar = await cookies()
    const refreshToken = jar.get(REFRESH_COOKIE)?.value

    if (refreshToken) {
        await logout({ refresh_token: refreshToken }).catch(() => { })
    }

    jar.delete(ACCESS_COOKIE)
    jar.delete(REFRESH_COOKIE)

    return new Response(null, { status: 204 })
}