import { inject, Injectable } from "@angular/core";
import { GoalStatusFilter, CreateGoalInput, Goal, GoalReport, Paginated } from "../models/api.types";
import { HttpClient, HttpParams } from "@angular/common/http";

export interface GoalFilters {
  page?: number;
  page_size?: number;
  sort?: 'id' | 'name' | '-id' | '-name';
  status?: GoalStatusFilter;
}

export interface UpdateGoalInput extends Partial<CreateGoalInput> {
  version?: number;
}

@Injectable({ providedIn: 'root' })
export class GoalService {
  private http = inject(HttpClient)

  list(filters: GoalFilters = {}) {
    return this.http.get<Paginated<Goal>>('/api/goals', {
      params: this.buildParams(filters),
    });
  }

  getById(id: string) {
    return this.http.get<Goal>(`/api/goals/${id}`);
  }

  report(id: string) {
    return this.http.get<GoalReport>(`/api/goals/report/${id}`);
  }

  create(input: CreateGoalInput) {
    return this.http.post<Goal>('/api/goals', input);
  }

  update(id: string, input: UpdateGoalInput) {
    return this.http.put<Goal>(`/api/goals/${id}`, input);
  }

  delete(id: string) {
    return this.http.delete<void>(`/api/goals/${id}`);
  }

  private buildParams(f: GoalFilters): HttpParams {
    let p = new HttpParams();
    if (f.page) p = p.set('page', f.page);
    if (f.page_size) p = p.set('page_size', f.page_size);
    if (f.sort) p = p.set('sort', f.sort);
    if (f.status) p = p.set('status', f.status);
    return p;
  }
}
