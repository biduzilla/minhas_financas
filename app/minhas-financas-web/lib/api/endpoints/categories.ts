import { Category, Paginated } from '@/types/api';
import 'server-only';
import { apiFetch } from '../client';

export async function listAllCategories(): Promise<Category[]> {
    const res = await apiFetch<Paginated<Category>>({
        service: 'category',
        path: '/v1/categories?page_size=100&sort=name',
        next: { tags: ['categories'], revalidate: 60 },
    });
    return res.content;
}

export async function getCategoryMap(): Promise<Map<string, Category>> {
    const list = await listAllCategories()
    const map = new Map<string, Category>()
    for (const c of list) {
        if (c.id) map.set(c.id, c)
    }

    return map
}