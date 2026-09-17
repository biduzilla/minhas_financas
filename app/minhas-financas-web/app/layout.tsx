import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Minhas Finanças',
  description: 'Controle suas finanças pessoais',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="pt-BR">
      <body className="antialiased">{children}</body>
    </html>
  );
}