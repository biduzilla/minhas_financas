'use client';

import { useState, useTransition } from 'react';
import { useRouter } from 'next/navigation';
import { PasswordInput } from '@/components/ui/password-input';
import type {
  HttpErrorResponse,
  ValidationErrorResponse,
} from '@/types/api';

type ApiError = HttpErrorResponse | ValidationErrorResponse;

export function SignUpForm() {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  async function handleSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setFieldErrors({});

    const fd = new FormData(e.currentTarget);
    const name = String(fd.get('name') ?? '').trim();
    const email = String(fd.get('email') ?? '').trim();
    const password = String(fd.get('password') ?? '');
    const confirm = String(fd.get('confirm_password') ?? '');

    if (password !== confirm) {
      setFieldErrors({ confirm_password: 'As senhas não conferem' });
      return;
    }

    startTransition(async () => {
      const res = await fetch('/api/auth/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, email, password }),
      });

      if (res.ok) {
        router.push('/login?signup=ok');
        return;
      }

      const body = (await res.json()) as ApiError;
      if ('errors' in body && body.errors) {
        setFieldErrors(body.errors);
        setError(body.message ?? 'Erro de validação');
      } else {
        setError(body.message ?? 'Falha no cadastro');
      }
    });
  }

  const inputCls =
    'w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-slate-900 placeholder:text-slate-400 transition-all duration-150 focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 disabled:opacity-60 disabled:cursor-not-allowed';
  const labelCls = 'mb-1.5 block text-sm font-medium text-slate-700';
  const errorCls = 'mt-1.5 flex items-center gap-1 text-xs text-red-600';

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      {/* NOME */}
      <div>
        <label htmlFor="name" className={labelCls}>
          Nome
        </label>
        <input
          id="name"
          name="name"
          type="text"
          required
          minLength={3}
          maxLength={100}
          autoComplete="name"
          disabled={isPending}
          placeholder="João Silva"
          className={inputCls}
        />
        {fieldErrors.name && (
          <p className={errorCls}>
            <span aria-hidden>⚠</span>
            {fieldErrors.name}
          </p>
        )}
      </div>

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
        <label htmlFor="password" className={labelCls}>
          Senha
        </label>
        <PasswordInput
          id="password"
          name="password"
          autoComplete="new-password"
          minLength={8}
          maxLength={72}
          disabled={isPending}
        />
        <p className="mt-1.5 text-xs text-slate-500">
          Mínimo 8 caracteres, com pelo menos uma letra e um número.
        </p>
        {fieldErrors.password && (
          <p className={errorCls}>
            <span aria-hidden>⚠</span>
            {fieldErrors.password}
          </p>
        )}
      </div>

      {/* CONFIRMAR SENHA */}
      <div>
        <label htmlFor="confirm_password" className={labelCls}>
          Confirmar senha
        </label>
        <PasswordInput
          id="confirm_password"
          name="confirm_password"
          autoComplete="new-password"
          disabled={isPending}
        />
        {fieldErrors.confirm_password && (
          <p className={errorCls}>
            <span aria-hidden>⚠</span>
            {fieldErrors.confirm_password}
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
        {isPending ? 'Criando…' : 'Criar conta'}
      </button>
    </form>
  );
}