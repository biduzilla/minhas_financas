// app/(auth)/login/page.tsx
import { LoginForm } from './login-form';

export default function LoginPage() {
    return (
        <div className="w-full max-w-md">
            <div className="rounded-2xl border border-slate-200 bg-white p-8 shadow-xl">
                {/* Cabeçalho */}
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

                <LoginForm />
            </div>

            {/* Rodapé */}
            <p className="mt-6 text-center text-xs text-slate-400">
                Não tem conta?{' '}
                <a
                    href="/signup"
                    className="font-medium text-emerald-600 hover:text-emerald-700 hover:underline"
                >
                    Cadastre-se
                </a>
            </p>
        </div>
    );
}