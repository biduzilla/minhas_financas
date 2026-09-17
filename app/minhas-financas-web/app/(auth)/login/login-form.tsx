// app/(auth)/login/login-form.tsx
'use client';

import { useState, useTransition } from 'react';
import { useRouter } from 'next/navigation';
import { HttpErrorResponse, ValidationErrorResponse } from '@/type/api';


type ApiError = HttpErrorResponse | ValidationErrorResponse;

export function LoginForm() {
    const router = useRouter();
    const [isPending, startTransition] = useTransition();
    const [error, setError] = useState<string | null>(null);
    const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
    const [showPassword, setShowPassword] = useState(false)

    async function handleSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
        e.preventDefault();
        setError(null);
        setFieldErrors({});

        const formData = new FormData(e.currentTarget)
        const payload = {
            email: String(formData.get('email') ?? ''),
            password: String(formData.get('password') ?? '')
        }

        startTransition(async () => {
            const res = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });

            if (res.ok) {
                router.push('/dashboard');
                router.refresh();
                return;
            }

            const body = (await res.json()) as ApiError;
            if ('errors' in body && body.errors) {
                setFieldErrors(body.errors);
                setError(body.message ?? 'Erro de validação');
            } else {
                setError(body.message ?? 'Falha no login');
            };
        });
    }

    return (
        <form onSubmit={handleSubmit} className="space-y-5">
            {/* EMAIL */}
            <div>
                <label
                    htmlFor="email"
                    className="block text-sm font-medium text-slate-700 mb-1.5"
                >
                    Email
                </label>
                <input
                    id="email"
                    name="email"
                    type="email"
                    required
                    autoComplete="email"
                    disabled={isPending}
                    placeholder="seu@email.com"
                    className="
          w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5
          text-slate-900 placeholder:text-slate-400
          transition-all duration-150
          focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20
          disabled:opacity-60 disabled:cursor-not-allowed
        "
                />
                {fieldErrors.email && (
                    <p className="mt-1.5 text-xs text-red-600 flex items-center gap-1">
                        <span aria-hidden>⚠</span>
                        {fieldErrors.email}
                    </p>
                )}
            </div>

            {/* SENHA */}
            <div>
                <div className="flex items-center justify-between mb-1.5">
                    <label
                        htmlFor="password"
                        className="block text-sm font-medium text-slate-700"
                    >
                        Senha
                    </label>
                    <a
                        href="#"
                        className="text-xs text-emerald-600 hover:text-emerald-700 hover:underline"
                    >
                        Esqueceu?
                    </a>
                </div>

                <div className="relative">
                    <input
                        id="password"
                        name="password"
                        type={showPassword ? 'text' : 'password'}
                        required
                        autoComplete="current-password"
                        disabled={isPending}
                        placeholder="••••••••"
                        className="
        w-full rounded-lg border border-slate-300 bg-white
        pl-3 pr-10 py-2.5
        text-slate-900 placeholder:text-slate-400
        transition-all duration-150
        focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20
        disabled:opacity-60 disabled:cursor-not-allowed
      "
                    />

                    {/* Botão de mostrar/ocultar senha */}
                    <button
                        type="button"
                        onClick={() => setShowPassword((v) => !v)}
                        disabled={isPending}
                        aria-label={showPassword ? 'Ocultar senha' : 'Mostrar senha'}
                        aria-pressed={showPassword}
                        title={showPassword ? 'Ocultar senha' : 'Mostrar senha'}
                        className="
        absolute right-2 top-1/2 -translate-y-1/2
        flex h-7 w-7 items-center justify-center
        rounded-md
        text-slate-400 hover:text-slate-600 hover:bg-slate-100
        focus:outline-none focus:ring-2 focus:ring-emerald-500/40
        transition-colors
        disabled:opacity-40 disabled:cursor-not-allowed
      "
                    >
                        {showPassword ? (
                            /* Ícone "olho cortado" — senha visível, clique pra esconder */
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                strokeWidth="2"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                aria-hidden="true"
                            >
                                <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
                                <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
                                <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61" />
                                <line x1="2" y1="2" x2="22" y2="22" />
                            </svg>
                        ) : (
                            /* Ícone "olho" — senha oculta, clique pra mostrar */
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                strokeWidth="2"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                aria-hidden="true"
                            >
                                <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
                                <circle cx="12" cy="12" r="3" />
                            </svg>
                        )}
                    </button>
                </div>

                {fieldErrors.password && (
                    <p className="mt-1.5 text-xs text-red-600 flex items-center gap-1">
                        <span aria-hidden>⚠</span>
                        {fieldErrors.password}
                    </p>
                )}
            </div>

            {/* ERRO GERAL */}
            {error && (
                <div
                    role="alert"
                    className="
          rounded-lg border border-red-200 bg-red-50 px-3 py-2.5
          text-sm text-red-700 flex items-start gap-2
        "
                >
                    <span aria-hidden className="shrink-0">⚠</span>
                    <span>{error}</span>
                </div>
            )}

            {/* BOTÃO */}
            <button
                type="submit"
                disabled={isPending}
                className="
        w-full rounded-lg px-4 py-2.5
        bg-emerald-600 text-white font-medium
        shadow-sm
        transition-all duration-150
        hover:bg-emerald-700
        focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2
        active:scale-[0.98]
        disabled:opacity-60 disabled:cursor-not-allowed disabled:active:scale-100
      "
            >
                {isPending ? 'Entrando…' : 'Entrar'}
            </button>
        </form>
    );
}