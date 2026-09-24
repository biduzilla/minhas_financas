import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { finalize } from 'rxjs';
import { CurrencyPipe, DatePipe } from '@angular/common';

import { GoalService } from '../../../core/services/goal.service';
import { GoalReport, GoalStatus } from '../../../core/models/api.types';

const STATUS_META: Record<GoalStatus, { label: string; cls: string }> = {
  IN_PROGRESS: { label: 'Em andamento', cls: 'bg-sky-100 text-sky-700' },
  COMPLETED:   { label: 'Concluída',    cls: 'bg-emerald-100 text-emerald-700' },
  EXPIRED:     { label: 'Vencida',      cls: 'bg-amber-100 text-amber-700' },
  CANCELED:    { label: 'Cancelada',    cls: 'bg-slate-100 text-slate-600' },
};

@Component({
  selector: 'app-goal-detail',
  standalone: true,
  imports: [RouterLink, CurrencyPipe, DatePipe],
  templateUrl: './goal-detail.component.html',
})
export class GoalDetailComponent {
  private goalService = inject(GoalService);
  private route = inject(ActivatedRoute);
  private destroyRef = inject(DestroyRef);

  readonly report = signal<GoalReport | null>(null);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);

  /** progress já vem 0-100 do backend */
  readonly progressPct = computed(() => {
    const r = this.report();
    if (!r) return 0;
    return Math.min(100, r.progress);
  });

  constructor() {
    this.route.paramMap
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((params) => {
        const id = params.get('id');
        if (id) this.load(id);
      });
  }

  getStatusMeta(status: GoalStatus) {
    return STATUS_META[status] ?? { label: status, cls: '' };
  }

  private load(id: string) {
    this.loading.set(true);
    this.error.set(null);

    this.goalService
      .report(id)
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (res) => this.report.set(res),
        error: (err: HttpErrorResponse) => {
          this.error.set(
            err.status === 404
              ? 'Meta não encontrada'
              : err.error?.message ?? 'Erro ao carregar relatório',
          );
        },
      });
  }
}
