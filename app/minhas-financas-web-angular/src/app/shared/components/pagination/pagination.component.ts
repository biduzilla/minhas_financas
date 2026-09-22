import { Component, computed, input, output } from '@angular/core';
import { PaginationMetadata } from '../../../core/models/api.types';

@Component({
  selector: 'app-pagination',
  standalone: true,
  templateUrl: './pagination.component.html',
})
export class PaginationComponent {
  meta = input.required<Partial<PaginationMetadata>>();
  pageChange = output<number>();

  currentPage = computed(() => this.meta().current_page ?? 1);
  lastPage = computed(() => this.meta().last_page ?? 1);
  total = computed(() => this.meta().total_records ?? 0);
  pageSize = computed(() => this.meta().page_size ?? 20);

  hasPrev = computed(() => this.currentPage() > 1);
  hasNext = computed(() => this.currentPage() < this.lastPage());

  /** Páginas ao redor da atual para o seletor compacto. */
  pages = computed<number[]>(() => {
    const current = this.currentPage();
    const last = this.lastPage();
    if (last <= 1) return [];

    const window = 2;
    const from = Math.max(1, current - window);
    const to = Math.min(last, current + window);
    return Array.from({ length: to - from + 1 }, (_, i) => from + i);
  });

  goTo(page: number) {
    if (page < 1 || page > this.lastPage() || page === this.currentPage()) return;
    this.pageChange.emit(page);
  }
}
