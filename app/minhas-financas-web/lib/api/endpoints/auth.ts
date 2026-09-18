import { HttpErrorResponse, LoginInput, RefreshInput, SignUpInput, TokenResponse, User, ValidationErrorResponse } from '@/types/api';
import 'server-only';

const AUTH_URL = process.env.AUTH_URL!;

export interface BackendError {
    status: number;
    body: HttpErrorResponse | ValidationErrorResponse | string
}

async function parseError(res: Response): Promise<BackendError> {
    const contentType = res.headers.get('content-type') ?? '';
    const body = contentType.includes('json')
        ? await res.json().catch(() => res.statusText)
        : await res.text();
    return { status: res.status, body }
}

export async function login(input: LoginInput): Promise<
    { ok: true, data: TokenResponse } | { ok: false; error: BackendError }> {
    const res = await fetch(`${AUTH_URL}/v1/auth`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
        cache: 'no-store',
    });
    if (!res.ok) return { ok: false, error: await parseError(res) };
    return { ok: true, data: await res.json() };
}

export async function refresh(input: RefreshInput): Promise<
    { ok: true; data: TokenResponse } | { ok: false; error: BackendError }
> {
    const res = await fetch(`${AUTH_URL}/v1/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
        cache: 'no-store',
    });

    if (!res.ok) return { ok: false, error: await parseError(res) };
    return { ok: true, data: await res.json() };
}

export async function logout(input: RefreshInput): Promise<
    { ok: true } | { ok: false; error: BackendError }
> {
    const res = await fetch(`${AUTH_URL}/v1/auth/logout`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
        cache: 'no-store',
    });

    if (!res.ok) return { ok: false, error: await parseError(res) };
    return { ok: true };
}

export async function signup(input: SignUpInput): Promise<
    { ok: true; data: User } | { ok: false; error: BackendError }
> {
    const res = await fetch(`${AUTH_URL}/v1/users`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
        cache: 'no-store',
    });

    if (!res.ok) return { ok: false, error: await parseError(res) };
    return { ok: true, data: await res.json() };
}