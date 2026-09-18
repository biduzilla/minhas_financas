import { login } from "@/lib/api/endpoints/auth";
import { LoginInput } from "@/types/api";
import { cookies } from "next/headers";
import {
    ACCESS_COOKIE,
    REFRESH_COOKIE,
    accessCookieOptions,
    refreshCookieOptions,
} from '@/lib/auth/cookies';

export async function POST(req: Request) {
    let input: LoginInput
    try {
        input = await req.json();
    } catch {
        return Response.json(
            { message: 'invalid JSON body' },
            { status: 400 },
        );
    }

    const result = await login(input)
    if (!result.ok) {
        return Response.json(result.error.body, { status: result.error.status });
    }

    const data = result.data
    const jar = await cookies()

    jar.set(ACCESS_COOKIE, data.access_token, accessCookieOptions(data.expires_in));
    jar.set(REFRESH_COOKIE, data.refresh_token, refreshCookieOptions(data.refresh_expires_in));

    return Response.json({ ok: true });
}