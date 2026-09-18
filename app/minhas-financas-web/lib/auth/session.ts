import { decodeJwt } from 'jose';
import { cookies } from 'next/headers';
import 'server-only';
import { ACCESS_COOKIE, REFRESH_COOKIE } from './cookies';

export interface SessionPayload {
    sub: string;
    username: string;
    type: 'access' | 'refresh';
    exp: number;
    iat: number;
    [key: string]: unknown;
}

export function decodeSession(token: string): SessionPayload | null {
    try {
        return decodeJwt(token) as SessionPayload
    } catch {
        return null
    }
}

export function isExpiringSoon(
    payload: SessionPayload | null,
    skewSeconds = 30,
): boolean {
    if (!payload) return true
    const now = Math.floor(Date.now() / 1000)
    return payload.exp - now <= skewSeconds
}

export async function getSession(): Promise<SessionPayload | null> {
    const jar = await cookies()
    const token = jar.get(ACCESS_COOKIE)?.value;
    if (!token) return null;
    return decodeSession(token);
}

export async function hasValidSession(): Promise<boolean> {
    const session = await getSession();
    return session !== null && !isExpiringSoon(session, 0);
}

export { ACCESS_COOKIE, REFRESH_COOKIE };