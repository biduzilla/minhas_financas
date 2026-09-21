'use client';

import { useState, useTransition } from 'react';
import { useRouter } from 'next/navigation';
import { PasswordInput } from '@/components/ui/password-input';
import type {
  HttpErrorResponse,
  ValidationErrorResponse,
} from '@/types/api';

type ApiError = HttpErrorResponse | ValidationErrorResponse;

export function LoginForm({ redirectTo }: { redirectTo?: string }) {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  async function handleSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setFieldErrors({});

    const formData = new FormData(e.currentTarget);
    const payload = {
      email: String(formData.get('email') ?? ''),
      password: String(formData.get('password') ?? ''),
    };

    startTransition(async () => {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        router.push(redirectTo ?? '/dashboard');
        router.refresh();
        return;
      }

      const body = (await res.json()) as ApiError;
      if ('errors' in body && body.errors) {
        setFieldErrors(body.errors);
        setError(body.message ?? 'Erro de validação');
      } else {
        setError(body.message ?? 'Falha no login');
      }
    });
  }

  const inputCls =
    'w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-slate-900 placeholder:text-slate-400 transition-all duration-150 focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 disabled:opacity-60 disabled:cursor-not-allowed';
  const labelCls = 'mb-1.5 block text-sm font-medium text-slate-700';
  const errorCls = 'mt-1.5 flex items-center gap-1 text-xs text-red-600';

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      {/* EMAIL */}
      <div>
        <label htmlFor="email" className={labelCls}>
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
          className={inputCls}
        />
        {fieldErrors.email && (
          <p className={errorCls}>
            <span aria-hidden>⚠</span>
            {fieldErrors.email}
          </p>
        )}
      </div>

      {/* SENHA */}
      <div>
        <div className="mb-1.5 flex items-center justify-between">
          <label htmlFor="password" className={labelCls}>
            Senha
          </label>
          {/* <a
            href="#"
            className="text-xs text-emerald-600 hover:text-emerald-700 hover:underline"
          >
            Esqueceu?
          </a> */}
        </div>
        <PasswordInput
          id="password"
          name="password"
          autoComplete="current-password"
          disabled={isPending}
        />
        {fieldErrors.password && (
          <p className={errorCls}>
            <span aria-hidden>⚠</span>
            {fieldErrors.password}
          </p>
        )}
      </div>

      {/* ERRO GERAL */}
      {error && (
        <div
          role="alert"
          className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-sm text-red-700"
        >
          <span aria-hidden className="shrink-0">
            ⚠
          </span>
          <span>{error}</span>
        </div>
      )}

      {/* BOTÃO */}
      <button
        type="submit"
        disabled={isPending}
        className="w-full rounded-lg bg-emerald-600 px-4 py-2.5 font-medium text-white shadow-sm transition-all duration-150 hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed disabled:active:scale-100"
      >
        {isPending ? 'Entrando…' : 'Entrar'}
      </button>
    </form>
  );
}