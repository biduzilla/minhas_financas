import { CurrencyPipe, DatePipe } from "@angular/common";
import { Component, computed, DestroyRef, inject, signal } from "@angular/core";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
import { PaginationComponent } from "../../../shared/components/pagination/pagination.component";
import { TransactionService } from "../../../core/services/transaction.service";
import { CategoryService } from "../../../core/services/category.service";
import { Transaction, PaginationMetadata, Category, TransactionFilters, CategoryType, Paginated } from "../../../core/models/api.types";
import { finalize } from "rxjs";
import { HttpErrorResponse } from "@angular/common/http";
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

@Component({
  selector: 'app-transactions-list',
  standalone: true,
  imports: [RouterLink, CurrencyPipe, DatePipe, PaginationComponent],
  templateUrl: './transactions-list.component.html',
})
export class TransactionsListComponent {
  private txService = inject(TransactionService)
  private categoryService = inject(CategoryService)
  private route = inject(ActivatedRoute)
  private router = inject(Router)
  private destroyRef = inject(DestroyRef)

  readonly transactions = signal<Transaction[]>([]);
  readonly metadata = signal<Partial<PaginationMetadata>>({});
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);

  readonly categories = signal<Category[]>([]);

  readonly filters = signal<TransactionFilters>({});

  readonly total = computed(() => this.metadata().total_records ?? 0);
  readonly isEmpty = computed(() => !this.loading() && this.transactions().length === 0);

  private readonly categoryMap = computed(() => {
    const map = new Map<string, Category>();
    for (const c of this.categories()) {
      if (c.id) map.set(c.id, c);
    }
    return map;
  });

  constructor() {
    this.loadCategories()

    this.route.queryParams
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((params) => {
        const parsed = this.parseFilters(params);
        this.filters.set(parsed);
        this.loadTransactions(parsed);
      })
  }

  private parseFilters(params: Record<string, string | undefined>): TransactionFilters {
    return {
      page: Number(params['page']) || 1,
      page_size: Number(params['page_size']) || 20,
      sort: (params['sort'] as TransactionFilters['sort']) || '-id',
      type: (params['type'] as CategoryType) || undefined,
      start_date: params['start_date'] || undefined,
      end_date: params['end_date'] || undefined,
      category_id: params['category_id'] || undefined,
    };
  }

  private updateUrl(patch: Record<string, string | number | null>) {
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: patch,
      queryParamsHandling: 'merge',
    })
  }

  onTypeChange(value: string) {
    this.updateUrl({ type: value || null, page: 1 })
  }

  onCategoryChange(value: string) {
    this.updateUrl({ category_id: value || null, page: 1 });
  }

  onStartDateChange(value: string) {
    this.updateUrl({ start_date: value || null, page: 1 });
  }

  onEndDateChange(value: string) {
    this.updateUrl({ end_date: value || null, page: 1 });
  }

  onSortChange(value: string) {
    this.updateUrl({ sort: value, page: 1 });
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

  hasActiveFilters = computed(() => {
    const f = this.filters();
    return Boolean(f.type || f.category_id || f.start_date || f.end_date);
  });

  private loadTransactions(filters: TransactionFilters) {
    this.loading.set(true)
    this.error.set(null)

    this.txService
      .list(filters)
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (res: Paginated<Transaction>) => {
          this.transactions.set(res.content)
          this.metadata.set(res.metadata)
        },
        error: (err: HttpErrorResponse) => {
          this.transactions.set([]);
          this.metadata.set({});
          this.error.set(err.error?.message ?? 'Erro ao carregar transações');
        }
      })
  }

  private loadCategories() {
    this.categoryService.list({ page_size: 100 }).subscribe({
      next: (res) => this.categories.set(res.content),
      error: () => this.categories.set([])
    })
  }

  categoryName(id: string): string {
    return this.categoryMap().get(id)?.name ?? '—';
  }

  isInput(id: string): boolean {
    return this.categoryMap().get(id)?.type === 'input';
  }
}
