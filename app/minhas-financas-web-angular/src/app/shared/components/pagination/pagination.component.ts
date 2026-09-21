import { Component, input, computed, inject } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { PaginationMetadata } from '../../../core/models/api.types';

@Component({
  selector: 'app-pagination',
  standalone: true,
  template: `
    @if (totalRecords() > 0) {
      <div class="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 pt-4">
        <p class="text-xs text-slate-500">
          Página {{ currentPage() }} de {{ lastPage() }}
          · {{ totalRecords() }} {{ totalRecords() === 1 ? 'registro' : 'registros' }}
        </p>

        <div class="flex items-center gap-2">
          <button
            type="button"
            [disabled]="currentPage() <= 1"
            (click)="goTo(currentPage() - 1)"
            class="rounded-lg border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Anterior
          </button>

          <button
            type="button"
            [disabled]="currentPage() >= lastPage()"
            (click)="goTo(currentPage() + 1)"
            class="rounded-lg border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Próxima
          </button>
        </div>
      </div>
    }
  `,
})
export class PaginationComponent {
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  meta = input<Partial<PaginationMetadata>>({});

  protected currentPage = computed(() => this.meta().current_page ?? 1);
  protected lastPage = computed(() => this.meta().last_page ?? 1);
  protected totalRecords = computed(() => this.meta().total_records ?? 0);

  protected goTo(page: number) {
    if (page < 1 || page > this.lastPage()) return;
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { page },
      queryParamsHandling: 'merge',
    });
  }
}
