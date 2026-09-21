/** Botão principal — ação primária (salvar, entrar, criar) */
export const btnPrimary =
  'inline-flex items-center justify-center rounded-lg bg-emerald-600 px-4 py-2.5 font-medium text-white shadow-sm transition-all duration-150 hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:ring-offset-2 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed disabled:active:scale-100';

/** Botão secundário — cancelar, voltar */
export const btnSecondary =
  'inline-flex items-center justify-center rounded-lg px-4 py-2.5 font-medium text-slate-700 transition-colors hover:bg-slate-100 focus:outline-none focus:ring-2 focus:ring-slate-300 focus:ring-offset-2 disabled:opacity-60 disabled:cursor-not-allowed';

/** Input padrão (text, email, password, number, select) */
export const inputBase =
  'w-full rounded-lg border border-slate-300 bg-white px-3 py-2.5 text-slate-900 placeholder:text-slate-400 transition-all duration-150 focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 disabled:opacity-60 disabled:cursor-not-allowed';

/** Label padrão */
export const labelBase = 'mb-1.5 block text-sm font-medium text-slate-700';

/** Mensagem de erro de campo (abaixo do input) */
export const fieldError = 'mt-1.5 flex items-center gap-1 text-xs text-red-600';

/** Banner de erro geral (role=alert) */
export const alertError =
  'flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5 text-sm text-red-700';

/** Banner de aviso (ex: conflito 409) */
export const alertWarning =
  'flex items-start gap-3 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800';

/** Card padrão */
export const card = 'rounded-2xl border border-slate-200 bg-white p-6 shadow-sm';

/** Card de listagem (linha clicável) */
export const cardHover =
  'block rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition-all hover:border-emerald-300 hover:shadow-md';

/** Título de seção (h1) */
export const h1 = 'text-2xl font-bold tracking-tight text-slate-900';

/** Subtítulo (muted) */
export const subtitle = 'mt-1 text-sm text-slate-500';
