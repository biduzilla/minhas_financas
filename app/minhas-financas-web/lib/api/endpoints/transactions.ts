import { Paginated, Transaction, TransactionFilters } from '@/types/api';
import 'server-only';
import { ApiError, apiFetch } from '../client';

function buildQuery(filters: TransactionFilters): string {
    const qs = new URLSearchParams();

    if (filters.page) qs.set('page', String(filters.page));
    if (filters.page_size) qs.set('page_size', String(filters.page_size));
    if (filters.sort) qs.set('sort', filters.sort);
    if (filters.start_date) qs.set('start_date', filters.start_date);
    if (filters.end_date) qs.set('end_date', filters.end_date);
    if (filters.category_id) qs.set('category_id', filters.category_id);
    if (filters.type) qs.set('type', filters.type);
    const s = qs.toString();
    return s ? `?${s}` : '';
}

export async function listTransactions(
    filters: TransactionFilters = {},
): Promise<Paginated<Transaction>> {
    return apiFetch<Paginated<Transaction>>({
        service: 'transaction',
        path: `/v1/transactions${buildQuery(filters)}`,
        next: { tags: ['transactions'], revalidate: 30 },
    });
}

export async function getTransaction(id: string): Promise<Transaction> {
    return apiFetch<Transaction>({
        service: 'transaction',
        path: `/v1/transactions/${id}`,
        next: { tags: ['transactions', `transaction:${id}`] },
    });
}

export { ApiError };