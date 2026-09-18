'use client'

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState, useTransition } from "react";

const NAV_ITEMS = [
    { href: '/dashboard', label: 'Dashboard', icon: '◫' },
    { href: '/transactions', label: 'Transações', icon: '⇄' },
    { href: '/categories', label: 'Categorias', icon: '⊞' },
    { href: '/goals', label: 'Metas', icon: '◎' },
] as const;

export function Sidebar() {
    const pathname = usePathname()
    const router = useRouter()
    const [isPending, startTransition] = useTransition()
    const [showMobile, setShowMobile] = useState(false)

    function handleLogout() {
        startTransition(async () => {
            try {
                await fetch('/api/auth/logout', { method: 'POST' });
            } catch {
            }
            router.push('/login');
            router.refresh();
        })
    }

    const isActive = (href: string) =>
        pathname === href || pathname.startsWith(href + '/');

    const content = (
        <>
            {/* LOGO */}
            <div className="mb-8 flex items-center gap-3 px-2">
                <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-400 to-teal-500 shadow-sm">
                    <span className="text-sm font-bold text-white">MF</span>
                </div>
                <span className="font-semibold tracking-tight text-slate-900">
                    Minhas Finanças
                </span>
            </div>

            {/* NAV */}
            <nav className="flex-1 space-y-1">
                {NAV_ITEMS.map((item) => (
                    <Link
                        key={item.href}
                        href={item.href}
                        onClick={() => setShowMobile(false)}
                        className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${isActive(item.href)
                            ? 'bg-emerald-50 text-emerald-700'
                            : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
                            }`}
                    >
                        <span aria-hidden className="text-base">{item.icon}</span>
                        {item.label}
                    </Link>
                ))}
            </nav>

            {/* LOGOUT */}
            <button
                type="button"
                onClick={handleLogout}
                disabled={isPending}
                className="mt-auto flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-rose-50 hover:text-rose-700 disabled:opacity-60"
            >
                <span aria-hidden className="text-base">⏻</span>
                {isPending ? 'Saindo…' : 'Sair'}
            </button>
        </>
    );

    return (
        <>
            {/* DESKTOP */}
            <aside className="hidden w-56 shrink-0 border-r border-slate-200 bg-white p-4 md:flex md:flex-col">
                {content}
            </aside>

            {/* MOBILE HEADER */}
            <div className="fixed inset-x-0 top-0 z-40 flex items-center justify-between border-b border-slate-200 bg-white px-4 py-2 md:hidden">
                <button
                    type="button"
                    onClick={() => setShowMobile(true)}
                    aria-label="Abrir menu"
                    className="flex h-9 w-9 items-center justify-center rounded-lg hover:bg-slate-100"
                >
                    ☰
                </button>
                <span className="font-semibold tracking-tight text-slate-900">
                    Minhas Finanças
                </span>
                <div className="w-9" />
            </div>

            {/* MOBILE DRAWER */}
            {showMobile && (
                <div
                    className="fixed inset-0 z-50 bg-slate-900/40 backdrop-blur-sm md:hidden"
                    onClick={() => setShowMobile(false)}
                >
                    <aside
                        className="flex h-full w-64 flex-col bg-white p-4 shadow-xl"
                        onClick={(e) => e.stopPropagation()}
                    >
                        {content}
                    </aside>
                </div>
            )}
        </>
    );
}