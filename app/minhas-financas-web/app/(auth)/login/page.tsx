import Link from 'next/link';
import { LoginForm } from './login-form';
import { hasValidSession } from '@/lib/auth/session';
import { redirect } from 'next/navigation';

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ signup?: string; redirect?: string }>;
}) {
  const { signup, redirect: redirectTo } = await searchParams;


  if (await hasValidSession()) {
    redirect(redirectTo ?? '/dashboard');
  }

  return (
    <div className="w-full max-w-md">
      <div className="rounded-2xl border border-slate-200 bg-white p-8 shadow-xl">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-400 to-teal-500 shadow-md">
            <span className="text-xl font-bold text-white">MF</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-slate-900">
            Minhas Finanças
          </h1>
          <p className="mt-1 text-sm text-slate-500">
            Entre para gerenciar suas finanças
          </p>
        </div>

        {signup === 'ok' && (
          <div
            role="status"
            className="mb-6 flex items-start gap-2 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2.5 text-sm text-emerald-700"
          >
            <span aria-hidden className="shrink-0">
              ✓
            </span>
            <span>Conta criada com sucesso. Faça login para continuar.</span>
          </div>
        )}

        <LoginForm redirectTo={redirectTo} />
      </div>

      <p className="mt-6 text-center text-xs text-slate-400">
        Não tem conta?{' '}
        <Link
          href="/signup"
          className="font-medium text-emerald-600 hover:text-emerald-700 hover:underline"
        >
          Cadastre-se
        </Link>
      </p>
    </div>
  );
}