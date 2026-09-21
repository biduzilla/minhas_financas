// app/(app)/transactions/page.tsx
import Link from 'next/link';
import { listTransactions } from '@/lib/api/endpoints/transactions';
import { getCategoryMap } from '@/lib/api/endpoints/categories';
import type { TransactionFilters as TFilters, CategoryType } from '@/types/api';
import { ApiError } from '@/lib/api/client';
import { Pagination } from '@/components/paginations';
import { TransactionFilters } from './filters';

const VALID_SORTS = ['id', 'amount', '-id', '-amount'] as const;
const VALID_TYPES = ['input', 'output'] as const;

type ValidSort = (typeof VALID_SORTS)[number];
type ValidType = (typeof VALID_TYPES)[number];

function isOneOf<T extends string>(v: unknown, list: readonly T[]): v is T {
    return typeof v === 'string' && (list as readonly string[]).includes(v);
}

function parseFilters(sp: Record<string, string | string[] | undefined>): TFilters {
    const one = (k: string) => {
        const v = sp[k];
        return Array.isArray(v) ? v[0] : v;
    };

    const page = Number(one('page') ?? 1) || 1;
    const page_size = Number(one('page_size') ?? 20) || 20;
    const sort = one('sort');
    const type = one('type');

    return {
        page: page > 0 ? page : 1,
        page_size: Math.min(Math.max(page_size, 1), 100),
        sort: isOneOf(sort, VALID_SORTS) ? sort : '-id',
        type: isOneOf(type, VALID_TYPES) ? type : undefined,
        start_date: one('start_date'),
        end_date: one('end_date'),
        category_id: one('category_id'),
    };
}

export default async function TransactionsPage({
    searchParams,
}: {
    searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
    const sp = await searchParams;
    const filters = parseFilters(sp);

    // Fetch paralelo: transações + categorias (pra resolver nomes e popular dropdown)
    let data, categoryMap;
    try {
        [data, categoryMap] = await Promise.all([
            listTransactions(filters),
            getCategoryMap(),
        ]);
    } catch (e) {
        return <ErrorState error={e} />;
    }

    const { content, metadata } = data;
    const total = metadata.total_records ?? 0;

    // Query string sem `page`, pra passar ao Pagination
    const baseParams: Record<string, string> = {};
    if (filters.type) baseParams.type = filters.type;
    if (filters.start_date) baseParams.start_date = filters.start_date;
    if (filters.end_date) baseParams.end_date = filters.end_date;
    if (filters.category_id) baseParams.category_id = filters.category_id;
    if (filters.sort && filters.sort !== '-id') baseParams.sort = filters.sort;

    // Lista de categorias pro dropdown de filtro
    const categories = Array.from(categoryMap.values());

    return (
        <div className="mx-auto max-w-6xl space-y-6 p-6 md:p-8">
            <header className="flex flex-wrap items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold tracking-tight text-slate-900">
                        Transações
                    </h1>
                    <p className="mt-1 text-sm text-slate-500">
                        {total} {total === 1 ? 'registro' : 'registros'}
                    </p>
                </div>
                <Link
                    href="/transactions/new"
                    className="rounded-lg bg-emerald-600 px-4 py-2 font-medium text-white shadow-sm transition-colors hover:bg-emerald-700"
                >
                    + Nova transação
                </Link>
            </header>

            <TransactionFilters categories={categories} />

            {content.length === 0 ? (
                <EmptyState hasFilters={Object.keys(baseParams).length > 0} />
            ) : (
                <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
                    <div className="overflow-x-auto">
                        <table className="w-full">
                            <thead className="bg-slate-50 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
                                <tr>
                                    <th className="px-5 py-3">Descrição</th>
                                    <th className="px-5 py-3">Categoria</th>
                                    <th className="px-5 py-3">Data</th>
                                    <th className="px-5 py-3 text-right">Valor</th>
                                    <th className="px-5 py-3"></th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-slate-100">
                                {content.map((t) => {
                                    const cat = categoryMap.get(t.category_id);
                                    const isInput = cat?.type === 'input';
                                    return (
                                        <tr key={t.id} className="transition-colors hover:bg-slate-50">
                                            <td className="px-5 py-4 font-medium text-slate-900">
                                                {t.description}
                                            </td>
                                            <td className="px-5 py-4">
                                                <span className="inline-flex items-center gap-2 text-sm text-slate-600">
                                                    <span
                                                        className={`inline-flex h-6 w-6 items-center justify-center rounded-full text-xs font-semibold ${isInput
                                                                ? 'bg-emerald-100 text-emerald-700'
                                                                : 'bg-rose-100 text-rose-700'
                                                            }`}
                                                        aria-hidden
                                                    >
                                                        {isInput ? '↑' : '↓'}
                                                    </span>
                                                    {cat?.name ?? '—'}
                                                </span>
                                            </td>
                                            <td className="px-5 py-4 text-sm text-slate-500">
                                                {new Date(t.created_at).toLocaleDateString('pt-BR')}
                                            </td>
                                            <td
                                                className={`px-5 py-4 text-right font-semibold ${isInput ? 'text-emerald-600' : 'text-rose-600'
                                                    }`}
                                            >
                                                {new Intl.NumberFormat('pt-BR', {
                                                    style: 'currency',
                                                    currency: 'BRL',
                                                }).format(t.amount)}
                                            </td>
                                            <td className="px-5 py-4 text-right">
                                                <Link
                                                    href={`/transactions/${t.id}`}
                                                    className="text-sm font-medium text-emerald-600 hover:text-emerald-700 hover:underline"
                                                >
                                                    Editar
                                                </Link>
                                            </td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    </div>
                </div>
            )}

            <Pagination
                meta={metadata}
                baseParams={baseParams}
                basePath="/transactions"
            />
        </div>
    );
}

/* ==================== Estados auxiliares ==================== */

function EmptyState({ hasFilters }: { hasFilters: boolean }) {
    return (
        <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-12 text-center">
            <p className="text-sm text-slate-500">
                {hasFilters
                    ? 'Nenhuma transação encontrada com esses filtros.'
                    : 'Você ainda não cadastrou nenhuma transação.'}
            </p>
            {!hasFilters && (
                <Link
                    href="/transactions/new"
                    className="mt-4 inline-block text-sm font-medium text-emerald-600 hover:underline"
                >
                    Cadastrar primeira transação →
                </Link>
            )}
        </div>
    );
}

function ErrorState({ error }: { error: unknown }) {
    const message =
        error instanceof ApiError
            ? error.status === 422
                ? 'Filtros inválidos. Verifique os campos e tente novamente.'
                : `Erro ao carregar transações (${error.status}).`
            : 'Erro inesperado ao carregar transações.';

    return (
        <div className="mx-auto max-w-6xl p-6 md:p-8">
            <div className="rounded-2xl border border-red-200 bg-red-50 p-8 text-center">
                <p className="font-medium text-red-700">{message}</p>
                <Link
                    href="/transactions"
                    className="mt-4 inline-block text-sm font-medium text-red-600 hover:underline"
                >
                    Limpar filtros
                </Link>
            </div>
        </div>
    );
}