'use client';

import { useRouter } from "next/navigation";
import { useTransition } from "react";

export function LogoutButton() {
    const router = useRouter();
    const [isPending, startTransition] = useTransition();

    function handleLogout() {
        startTransition(async () => {
            await fetch('/api/auth/logout', { method: 'POST' });
            router.push('/login')
            router.refresh();
        })
    }

    return (
        <button
            onClick={handleLogout}
            disabled={isPending}
            className="rounded border px-4 py-2 disabled:opacity-50"
        >
            {isPending ? 'Saindo…' : 'Sair'}
        </button>
    )
}