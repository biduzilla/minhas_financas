import { Component, computed, DestroyRef, inject, signal } from "@angular/core";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
import { PaginationComponent } from "../../../shared/components/pagination/pagination.component";
import { CategoryFilters, CategoryService } from "../../../core/services/category.service";
import { Category, CategoryType, ApiErrorBody, PaginationMetadata } from "../../../core/models/api.types";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";
import { finalize } from "rxjs";
import { HttpErrorResponse } from "@angular/common/http";

@Component({
  selector: 'app-categories-list',
  standalone: true,
  imports: [RouterLink, PaginationComponent],
  templateUrl: './categories-list.component.html',
})
export class CategoriesListComponent {
  categoryService = inject(CategoryService)
  private route = inject(ActivatedRoute)
  private router = inject(Router)
  private destroyRef = inject(DestroyRef)

  readonly categories = signal<Category[]>([])
  readonly metadata = signal<Partial<PaginationMetadata>>({})
  readonly loading = signal(true)
  readonly error = signal<string | null>(null);
  readonly filters = signal<CategoryFilters>({})

  readonly pendingDeleteId = signal<string | null>(null)
  readonly deletingId = signal<string | null>(null)

  readonly total = computed(() => this.metadata().total_records ?? 0)
  readonly isEmpty = computed(() =>
    !this.loading() && this.categories().length === 0
  )
  readonly hasActiveFilters = computed(() => Boolean(this.filters().type))

  constructor() {
    this.route.queryParams
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((params) => {
        const parsed = this.parseFilters(params)
        this.filters.set(parsed)
        this.loadCategories(parsed)
      })
  }

  private parseFilters(params: Record<string, string | undefined>): CategoryFilters {
    return {
      page: Number(params['page']) || 1,
      page_size: Number(params['page_size']) || 20,
      sort: (params['sort'] as CategoryFilters['sort']) || 'name',
      type: (params['type'] as CategoryType) || undefined,
    };
  }

  private updateUrl(path: Record<string, string | number | null>) {
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: path,
      queryParamsHandling: 'merge'
    })
  }

  onTypeChange(value: string) {
    this.updateUrl({ type: value || null, page: 1 })
  }

  onPageChange(page: number) {
    this.updateUrl({ page })
  }

  clearFilters() {
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { page: 1 }
    })
  }

  askDelete(id: string) {
    this.pendingDeleteId.set(id)
  }

  cancelDelete() {
    this.pendingDeleteId.set(null)
  }

  confirmDelete(id: string) {
    this.deletingId.set(id)
    this.categoryService
      .delete(id)
      .pipe(finalize(() => this.deletingId.set(null)))
      .subscribe({
        next: () => {
          this.pendingDeleteId.set(null)
          this.loadCategories(this.filters());
        },
        error: (err: HttpErrorResponse) => {
          const body = err.error as ApiErrorBody | undefined;
          this.error.set(body?.message ?? 'Erro ao carregar categorias');
        }
      })
  }

  isLinkedToGoal(c: Category): boolean {
    return Boolean(c.goal_id)
  }

  private loadCategories(filters: CategoryFilters) {
    this.loading.set(true)
    this.error.set(null)

    this.categoryService
      .list(filters)
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (res) => {
          this.categories.set(res.content);
          this.metadata.set(res.metadata);
        },
        error: (err: HttpErrorResponse) => {
          this.categories.set([]);
          this.metadata.set({});
          this.error.set(err.message ?? 'Erro ao carregar categorias');
        },
      });
  }
}
