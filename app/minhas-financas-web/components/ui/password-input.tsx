'use client';

import { useState } from 'react';


function EyeIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

function EyeOffIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24" />
      <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68" />
      <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61" />
      <line x1="2" y1="2" x2="22" y2="22" />
    </svg>
  );
}

/* ============ Componente público ============ */

const inputBase =
  'w-full rounded-lg border border-slate-300 bg-white pl-3 pr-10 py-2.5 text-slate-900 placeholder:text-slate-400 transition-all duration-150 focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 disabled:opacity-60 disabled:cursor-not-allowed';

export function PasswordInput({
  id,
  name,
  placeholder = '••••••••',
  disabled,
  required = true,
  autoComplete = 'current-password',
  minLength,
  maxLength,
  className = '',
}: {
  id: string;
  name: string;
  placeholder?: string;
  disabled?: boolean;
  required?: boolean;
  autoComplete?: string;
  minLength?: number;
  maxLength?: number;
  className?: string;
}) {
  const [show, setShow] = useState(false);

  return (
    <div className="relative">
      <input
        id={id}
        name={name}
        type={show ? 'text' : 'password'}
        required={required}
        minLength={minLength}
        maxLength={maxLength}
        autoComplete={autoComplete}
        disabled={disabled}
        placeholder={placeholder}
        className={`${inputBase} ${className}`}
      />
      <button
        type="button"
        onClick={() => setShow((v) => !v)}
        disabled={disabled}
        aria-label={show ? 'Ocultar senha' : 'Mostrar senha'}
        aria-pressed={show}
        title={show ? 'Ocultar senha' : 'Mostrar senha'}
        className="absolute right-2 top-1/2 flex h-7 w-7 -translate-y-1/2 items-center justify-center rounded-md text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600 focus:outline-none focus:ring-2 focus:ring-emerald-500/40 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        {show ? <EyeOffIcon /> : <EyeIcon />}
      </button>
    </div>
  );
}