import { NextRequest, NextResponse } from 'next/server';
import { decodeJwt } from 'jose';
import { accessCookieOptions, refreshCookieOptions } from './lib/auth/cookies';

const ACCESS_COOKIE = 'access_token';
const REFRESH_COOKIE = 'refresh_token';
const SKEW_SECONDS = 30

interface SessionPayload {
    exp: number;
    [key: string]: unknown;
}

function isExpired(token: string, skew = SKEW_SECONDS): boolean {
    try {
        const payload = decodeJwt(token) as SessionPayload;
        const now = Math.floor(Date.now() / 1000);
        return payload.exp - now <= skew;
    } catch {
        return true;
    }
}

export async function proxy(req: NextRequest) {
    const access = req.cookies.get(ACCESS_COOKIE)?.value
    const refresh = req.cookies.get(REFRESH_COOKIE)?.value

    if (!access || isExpired(access)) {
        if (!refresh) {
            return redirectToLogin(req)
        }


        const refreshRes = await fetch(`${process.env.AUTH_URL}/v1/auth/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refresh }),
            cache: 'no-store',
        })

        if (!refreshRes.ok) {
            const res = redirectToLogin(req);
            res.cookies.delete(ACCESS_COOKIE)
            res.cookies.delete(REFRESH_COOKIE)
            return res
        }

        const data = await refreshRes.json()
        const res = NextResponse.next()

        res.cookies.set(
            ACCESS_COOKIE,
            data.access_token,
            accessCookieOptions(data.expires_in),
        );
        res.cookies.set(
            REFRESH_COOKIE,
            data.refresh_token,
            refreshCookieOptions(data.refresh_expires_in),
        );

        return res;
    }
    return NextResponse.next()
}

function redirectToLogin(req: NextRequest) {
    const loginUrl = new URL('/login', req.url);
    loginUrl.searchParams.set('redirect', req.nextUrl.pathname)
    return NextResponse.redirect(loginUrl)
}

export const config = {
    matcher: [
        '/dashboard/:path*',
        '/transactions/:path*',
        '/categories/:path*',
        '/goals/:path*',
    ],
};