import { login } from "@/lib/api/endpoints/auth";
import { LoginInput } from "@/type/api";
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

    const { access_token, refresh_token, expires_in } = result.data
    const jar = await cookies()

    jar.set(ACCESS_COOKIE, access_token, accessCookieOptions(expires_in));
    jar.set(REFRESH_COOKIE, refresh_token, refreshCookieOptions());

    return Response.json({ ok: true });
}