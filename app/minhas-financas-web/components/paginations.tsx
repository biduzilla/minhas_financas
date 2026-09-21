// components/pagination.tsx
import Link from 'next/link';
import type { PaginationMetadata } from '@/types/api';

interface Props {
  meta: Partial<PaginationMetadata>;
  baseParams?: Record<string, string>;
  basePath?: string;
}

export function Pagination({
  meta,
  baseParams = {},
  basePath = '',
}: Props) {
  const current = meta.current_page ?? 1;
  const last = meta.last_page ?? 1;

  if (last <= 1) return null;

  const visible = new Set<number>([1, last, current]);
  if (current > 1) visible.add(current - 1);
  if (current < last) visible.add(current + 1);
  const pages = Array.from(visible).sort((a, b) => a - b);

  function href(page: number) {
    const qs = new URLSearchParams({ ...baseParams, page: String(page) });
    return `${basePath}?${qs}`;
  }

  return (
    <nav
      aria-label="Paginação"
      className="flex items-center justify-center gap-1"
    >
      {/* Anterior */}
      {current > 1 && (
        <Link
          href={href(current - 1)}
          className="rounded-lg px-3 py-1.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-100"
        >
          ← Anterior
        </Link>
      )}

      {pages.map((p, i) => {
        const prev = pages[i - 1];
        const gap = prev != null && p - prev > 1;
        return (
          <span key={p} className="flex items-center gap-1">
            {gap && <span className="px-2 text-slate-400">…</span>}
            <Link
              href={href(p)}
              aria-current={p === current ? 'page' : undefined}
              className={`min-w-[2.25rem] rounded-lg px-3 py-1.5 text-center text-sm font-medium transition-colors ${
                p === current
                  ? 'bg-emerald-600 text-white'
                  : 'text-slate-700 hover:bg-slate-100'
              }`}
            >
              {p}
            </Link>
          </span>
        );
      })}

      {/* Próxima */}
      {current < last && (
        <Link
          href={href(current + 1)}
          className="rounded-lg px-3 py-1.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-100"
        >
          Próxima →
        </Link>
      )}
    </nav>
  );
}