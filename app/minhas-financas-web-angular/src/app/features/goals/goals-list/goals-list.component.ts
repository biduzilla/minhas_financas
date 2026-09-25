import { CurrencyPipe, DatePipe } from "@angular/common";
import { Component, computed, DestroyRef, inject, signal } from "@angular/core";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
import { Goal, GoalStatus, GoalStatusFilter, PaginationMetadata } from "../../../core/models/api.types";
import { PaginationComponent } from "../../../shared/components/pagination/pagination.component";
import { GoalFilters, GoalService } from "../../../core/services/goal.service";
import { HttpErrorResponse } from "@angular/common/http";
import { finalize } from "rxjs";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";

const STATUS_META: Record<GoalStatus, { label: string; cls: string }> = {
  IN_PROGRESS: { label: 'Em andamento', cls: 'bg-sky-100 text-sky-700' },
  COMPLETED: { label: 'Concluída', cls: 'bg-emerald-100 text-emerald-700' },
  EXPIRED: { label: 'Vencida', cls: 'bg-amber-100 text-amber-700' },
  CANCELED: { label: 'Cancelada', cls: 'bg-slate-100 text-slate-600' },
};

@Component({
  selector: 'app-goals-list',
  standalone: true,
  imports: [RouterLink, CurrencyPipe, DatePipe, PaginationComponent],
  templateUrl: './goals-list.component.html',
})
export class GoalsListComponent {
  private goalService = inject(GoalService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private destroyRef = inject(DestroyRef);

  readonly goals = signal<Goal[]>([])
  readonly metadata = signal<Partial<PaginationMetadata>>({})
  readonly loading = signal(true)
  readonly error = signal<string | null>(null);
  readonly filters = signal<GoalFilters>({});

  readonly pendingDeleteId = signal<string | null>(null);
  readonly deletingId = signal<string | null>(null);

  readonly total = computed(() => this.metadata().total_records ?? 0);
  readonly isEmpty = computed(() => !this.loading() && this.goals().length === 0);
  readonly hasActiveFilters = computed(() => Boolean(this.filters().status));


  constructor() {
    this.route.queryParams
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((params) => {
        const parsed = this.parseFilters(params)
        this.filters.set(parsed)
        this.loadGoals(parsed)
      })
  }

  private parseFilters(params: Record<string, string | undefined>): GoalFilters {
    return {
      page: Number(params['page']) || 1,
      page_size: Number(params['page_size']) || 20,
      sort: (params['sort'] as GoalFilters['sort']) || '-id',
      status: (params['status'] as GoalStatusFilter) || undefined,
    };
  }

  private updateUrl(patch: Record<string, string | number | null>) {
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: patch,
      queryParamsHandling: 'merge',
    });
  }

  private loadGoals(filters: GoalFilters) {
    this.loading.set(true);
    this.error.set(null);

    this.goalService
      .list(filters)
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (res) => {
          this.goals.set(res.content);
          this.metadata.set(res.metadata);
        },
        error: (err: HttpErrorResponse) => {
          this.goals.set([]);
          this.metadata.set({});
         this.error.set(err.error.message);
        },
      });
  }

  onStatusChange(value: string) {
    this.updateUrl({ status: value || null, page: 1 });
  }

  onPageChange(page: number) {
    this.updateUrl({ page });
  }

  clearFilters() {
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { page: 1 },
    });
  }

  askDelete(id: string) {
    this.pendingDeleteId.set(id);
  }

  cancelDelete() {
    this.pendingDeleteId.set(null);
  }

  confirmDelete(id: string) {
    this.deletingId.set(id);
    this.goalService
      .delete(id)
      .pipe(finalize(() => this.deletingId.set(null)))
      .subscribe({
        next: () => {
          this.pendingDeleteId.set(null);
          this.loadGoals(this.filters());
        },
        error: (err: HttpErrorResponse) => {
          this.pendingDeleteId.set(null);
          this.error.set(err.error.message);
        },
      });
  }

  getStatusMeta(status: GoalStatus) {
    return STATUS_META[status] ?? { label: status, cls: '' };
  }

  getProgress(g: Goal): number {
    if (!g.target_amount) return 0;
    return Math.min(100, (g.current_amount / g.target_amount) * 100);
  }
}
