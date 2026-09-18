import { redirect } from 'next/navigation';
import { hasValidSession } from '@/lib/auth/session';
import { Sidebar } from './sidebar';

export default async function AppLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  if (!(await hasValidSession())) {
    redirect('/login');
  }

  return (
    <div className="flex min-h-screen bg-slate-50">
      <Sidebar />
      <main className="flex-1 overflow-x-hidden">{children}</main>
    </div>
  );
}