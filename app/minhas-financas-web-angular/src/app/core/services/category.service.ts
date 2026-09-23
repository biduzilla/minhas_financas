import { HttpClient, HttpParams } from "@angular/common/http";
import { inject, Injectable } from "@angular/core";
import { Paginated, Category, CategoryType, CreateCategoryInput } from "../models/api.types";

export interface CategoryFilters {
  page?: number;
  page_size?: number;
  sort?: 'id' | 'name' | '-id' | '-name';
  type?: CategoryType;
}

@Injectable({ providedIn: 'root' })
export class CategoryService {
  private http = inject(HttpClient)

  list(filters: CategoryFilters) {
    return this.http.get<Paginated<Category>>('/api/categories',
      { params: this.buildParams(filters) });
  }

  getById(id: string) {
    return this.http.get<Category>(`/api/categories/${id}`)
  }

  create(input: CreateCategoryInput) {
    return this.http.post<Category>('/api/categories', input);
  }

  update(id: string, input: CreateCategoryInput) {
    return this.http.put<Category>(`/api/categories/${id}`, input);
  }

  delete(id: string) {
    return this.http.delete<void>(`/api/categories/${id}`);
  }

  private buildParams(f: CategoryFilters): HttpParams {
    let p = new HttpParams();
    if (f.page) p = p.set('page', f.page);
    if (f.page_size) p = p.set('page_size', f.page_size);
    if (f.sort) p = p.set('sort', f.sort);
    if (f.type) p = p.set('type', f.type);
    return p;
  }
}
