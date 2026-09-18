export default function DashboardPage() {
    return (
        <div className="mx-auto max-w-6xl p-6 md:p-8">
            <header>
                <h1 className="text-2xl font-bold tracking-tight text-slate-900">
                    Dashboard
                </h1>
                <p className="mt-1 text-sm text-slate-500">
                    Bem-vindo de volta 👋
                </p>
            </header>

            <div className="mt-8 rounded-2xl border border-dashed border-slate-300 bg-white p-12 text-center">
                <p className="text-sm text-slate-500">
                    Em breve: cards de saldo, entradas, saídas e resumo por categoria.
                </p>
                <p className="mt-2 text-xs text-slate-400">
                    Próxima feature: Summary + Goal Report (Feature 5 do roadmap)
                </p>
            </div>
        </div>
    );
}