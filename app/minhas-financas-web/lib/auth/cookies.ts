import { ResponseCookie } from "next/dist/compiled/@edge-runtime/cookies";

export const ACCESS_COOKIE = 'access_token';
export const REFRESH_COOKIE = 'refresh_token';

const isProd = process.env.NODE_ENV === 'production';

const baseCookieOptions: Partial<ResponseCookie> = {
    httpOnly: true,
    secure: isProd,
    sameSite: 'lax',
    path: '/',
}

export function accessCookieOptions(maxAgeSeconds: number) {
    return { ...baseCookieOptions, maxAge: maxAgeSeconds };
}

export function refreshCookieOptions(maxAgeSeconds: number) {
    return { ...baseCookieOptions, maxAge: maxAgeSeconds };
}