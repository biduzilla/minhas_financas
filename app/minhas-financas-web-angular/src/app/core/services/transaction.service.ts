import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import {
  CreateTransactionInput,
  Paginated,
  Summary,
  Transaction,
  TransactionFilters,
  UpdateTransactionInput,
} from '../models/api.types';

@Injectable({ providedIn: 'root' })
export class TransactionService {
  private http = inject(HttpClient);

  list(filters: TransactionFilters = {}) {
    return this.http.get<Paginated<Transaction>>('/api/transactions', {
      params: this.buildParams(filters),
    });
  }

  getById(id: string) {
    return this.http.get<Transaction>(`/api/transactions/${id}`);
  }

  summary(filters: Pick<TransactionFilters, 'start_date' | 'end_date' | 'category_id' | 'type'> = {}) {
    return this.http.get<Summary>('/api/transactions/summary', {
      params: this.buildParams(filters),
    });
  }

  create(input: CreateTransactionInput) {
    return this.http.post<Transaction>('/api/transactions', input);
  }

  update(id: string, input: UpdateTransactionInput) {
    return this.http.put<Transaction>(`/api/transactions/${id}`, input);
  }

  delete(id: string) {
    return this.http.delete<void>(`/api/transactions/${id}`);
  }

  private buildParams(f: TransactionFilters): HttpParams {
    let p = new HttpParams();
    if (f.page) p = p.set('page', f.page);
    if (f.page_size) p = p.set('page_size', f.page_size);
    if (f.sort) p = p.set('sort', f.sort);
    if (f.type) p = p.set('type', f.type);
    if (f.start_date) p = p.set('start_date', f.start_date);
    if (f.end_date) p = p.set('end_date', f.end_date);
    if (f.category_id) p = p.set('category_id', f.category_id);
    return p;
  }
}
