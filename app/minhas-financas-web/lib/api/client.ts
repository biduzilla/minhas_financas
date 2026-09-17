import { ApiErrorResponse, ValidationErrorResponse } from '@/type/api';
import { cookies } from 'next/headers';
import 'server-only';

export class ApiError extends Error {
    readonly status: number;
    readonly body: ApiErrorResponse;
    readonly path: string;

    constructor(status: number, body: ApiErrorResponse, path: string) {
        super(body.message || `HTTP ${status}`);
        this.name = 'ApiError';
        this.status = status;
        this.body = body;
        this.path = path;
    }

    isValidation(): this is ApiError & { body: ValidationErrorResponse } {
        return (
            this.status === 422 &&
            typeof this.body === 'object' &&
            this.body !== null &&
            'errors' in this.body
        );
    }

    get fieldErrors(): Record<string, string> {
        return this.isValidation() ? this.body.errors : {};
    }

    get userMessage(): string {
        if (this.isValidation()) {
            const first = Object.values(this.fieldErrors)[0];
            return first ?? 'Verifique os campos e tente novamente.';
        }

        return this.message ?? 'Algo deu errado'
    }
}

export function isApiError(e: unknown): e is ApiError {
    return e instanceof ApiError;
}

export function isValidationError(
    e: unknown,
): e is ApiError & { body: ValidationErrorResponse } {
    return isApiError(e) && e.isValidation();
}

export function fieldErrors(e: unknown): Record<string, string> {
    return isValidationError(e) ? e.body.errors : {};
}

/** Mensagem amigável para o usuário, sem vazar stack/detalhe técnico */
export function userMessage(e: unknown): string {
    if (isApiError(e)) return e.userMessage;
    if (e instanceof Error) return e.message;
    return 'Algo deu errado.';
}

async function parseErrorBody(res: Response, path: string): Promise<ApiErrorResponse> {
    try {
        const body = (await res.json()) as ApiErrorResponse;
        return body;
    } catch {
        return {
            path,
            status: res.statusText || 'Error',
            message: `HTTP ${res.status} — resposta sem corpo JSON`,
        };
    }
}

export interface ApiFetchOptions extends Omit<RequestInit, 'body'> {
    body?: string;
    baseURL?: string;
    skipAuth?: boolean;
}

export async function apiFetch<T = unknown>(
    url: string,
    opts: ApiFetchOptions = {},
): Promise<T> {
    const { baseURL, skipAuth, headers, body, ...rest } = opts;

    const fullUrl = baseURL ? `${baseURL}${url}` : url;

    const finalHeaders = new Headers(headers);
    if (!finalHeaders.has('Content-Type') && body) {
        finalHeaders.set('Content-Type', 'application/json');
    }

    if (!skipAuth) {
        const jar = await cookies();
        const token = jar.get('access_token')?.value;
        if (token) finalHeaders.set('Authorization', `Bearer ${token}`);
    }

    const res = await fetch(fullUrl, {
        ...rest,
        body,
        headers: finalHeaders,
        cache: rest.cache ?? 'no-store',
    });

    if (!res.ok) {
        const errBody = await parseErrorBody(res, fullUrl);
        throw new ApiError(res.status, errBody, fullUrl);
    }

    if (res.status === 204) return undefined as T;

    return (await res.json()) as T;
}