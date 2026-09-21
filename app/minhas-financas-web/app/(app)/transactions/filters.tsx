'use client';

import { useRouter, useSearchParams, usePathname } from 'next/navigation';
import { useState, useTransition, useEffect } from 'react';
import type { Category } from '@/types/api';

interface Props {
    categories: Category[];
}

export function TransactionFilters({ categories }: Props) {
    const router = useRouter();
    const pathname = usePathname();
    const searchParams = useSearchParams();
    const [isPending, startTransition] = useTransition();

    const type = searchParams.get('type') ?? '';
    const categoryId = searchParams.get('category_id') ?? '';
    const startDate = searchParams.get('start_date') ?? '';
    const endDate = searchParams.get('end_date') ?? '';

    function apply(overrides: Record<string, string | undefined>) {
        const next = new URLSearchParams(searchParams.toString());
        next.delete('page');

        for (const [key, val] of Object.entries(overrides)) {
            if (val) next.set(key, val);
            else next.delete(key);
        }

        startTransition(() => {
            router.push(`${pathname}?${next.toString()}`);
        });
    }

    function clearAll() {
        startTransition(() => {
            router.push(pathname);
        });
    }

    const hasFilters =
        !!type || !!categoryId || !!startDate || !!endDate;

    const selectCls =
        'w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 transition-all focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 disabled:opacity-60';

    return (
        <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
            <div className="grid gap-3 md:grid-cols-4">
                {/* TIPO */}
                <div>
                    <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                        Tipo
                    </label>
                    <select
                        value={type}
                        onChange={(e) => apply({ type: e.target.value || undefined })}
                        disabled={isPending}
                        className={selectCls}
                    >
                        <option value="">Todos</option>
                        <option value="input">Entradas</option>
                        <option value="output">Saídas</option>
                    </select>
                </div>

                {/* CATEGORIA */}
                <div>
                    <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                        Categoria
                    </label>
                    <select
                        value={categoryId}
                        onChange={(e) => apply({ category_id: e.target.value || undefined })}
                        disabled={isPending}
                        className={selectCls}
                    >
                        <option value="">Todas</option>
                        {categories.map((c) => (
                            <option key={c.id} value={c.id ?? ''}>
                                {c.name}
                            </option>
                        ))}
                    </select>
                </div>

                {/* DATA INICIAL */}
                <div>
                    <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                        De
                    </label>
                    <input
                        type="date"
                        value={startDate}
                        onChange={(e) => apply({ start_date: e.target.value || undefined })}
                        disabled={isPending}
                        className={selectCls}
                    />
                </div>

                {/* DATA FINAL */}
                <div>
                    <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                        Até
                    </label>
                    <input
                        type="date"
                        value={endDate}
                        onChange={(e) => apply({ end_date: e.target.value || undefined })}
                        disabled={isPending}
                        className={selectCls}
                    />
                </div>
            </div>

            {hasFilters && (
                <div className="mt-3 flex items-center justify-between border-t border-slate-100 pt-3">
                    <p className="text-xs text-slate-500">
                        Filtros ativos
                        {isPending && ' · atualizando…'}
                    </p>
                    <button
                        type="button"
                        onClick={clearAll}
                        disabled={isPending}
                        className="text-xs font-medium text-slate-600 hover:text-slate-900 hover:underline disabled:opacity-60"
                    >
                        Limpar filtros
                    </button>
                </div>
            )}
        </div>
    );
}