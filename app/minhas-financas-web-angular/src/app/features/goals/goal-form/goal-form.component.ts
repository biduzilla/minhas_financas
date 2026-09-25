import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { finalize } from 'rxjs';
import { form, FormField, required, min, minLength, maxLength } from '@angular/forms/signals';

import { GoalService } from '../../../core/services/goal.service';

interface GoalFormData {
  name: string;
  target_amount: number | null;
  deadline: string;         // YYYY-MM-DD (input date)
  description: string;
  version?: number;
}

@Component({
  selector: 'app-goal-form',
  standalone: true,
  imports: [FormField, RouterLink],
  templateUrl: './goal-form.component.html',
})
export class GoalFormComponent {
  private goalService = inject(GoalService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private destroyRef = inject(DestroyRef);

  readonly goalId = signal<string | null>(null);
  readonly isEdit = computed(() => this.goalId() !== null);

  readonly model = signal<GoalFormData>({
    name: '',
    target_amount: null,
    deadline: '',
    description: '',
  });

  readonly goalForm = form(this.model, (path) => {
    required(path.name, { message: 'Nome é obrigatório' });
    minLength(path.name, 3, { message: 'Nome deve ter pelo menos 3 caracteres' });
    maxLength(path.name, 100, { message: 'Nome deve ter no máximo 100 caracteres' });

    required(path.target_amount, { message: 'Valor-alvo é obrigatório' });
    min(path.target_amount, 0.01, { message: 'Valor-alvo deve ser maior que zero' });

    required(path.deadline, { message: 'Prazo é obrigatório' });

    maxLength(path.description, 500, { message: 'Descrição deve ter no máximo 500 caracteres' });
  });

  readonly loading = signal(false);
  readonly submitting = signal(false);
  readonly error = signal<string | null>(null);
  readonly fieldErrors = signal<Record<string, string>>({});

  constructor() {
    this.route.paramMap
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((params) => {
        const id = params.get('id');
        this.goalId.set(id);
        if (id) this.loadGoal(id);
      });
  }

  onSubmit() {
    if (this.goalForm().invalid()) return;

    this.submitting.set(true);
    this.error.set(null);
    this.fieldErrors.set({});

    const data = this.model();
    const id = this.goalId();

    // input date → ISO 8601 com hora
    // Ex: "2026-12-31" → "2026-12-31T00:00:00Z"
    const deadlineISO = new Date(`${data.deadline}T00:00:00Z`).toISOString();

    const payload = {
      name: data.name,
      target_amount: data.target_amount!,
      deadline: deadlineISO,
      description: data.description || undefined,
    };

    const request$ = id
      ? this.goalService.update(id, { ...payload, version: data.version })
      : this.goalService.create(payload);

    request$
      .pipe(finalize(() => this.submitting.set(false)))
      .subscribe({
        next: () => this.router.navigate(['/goals']),
        error: (err: HttpErrorResponse) => this.handleError(err),
      });
  }

  backendError(field: keyof GoalFormData): string | null {
    return this.fieldErrors()[field] ?? null;
  }

  private loadGoal(id: string) {
    this.loading.set(true);
    this.error.set(null);

    this.goalService
      .getById(id)
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (g) => {
          this.model.set({
            name: g.name,
            target_amount: g.target_amount,
            deadline: g.deadline ? g.deadline.substring(0, 10) : '',
            description: g.description ?? '',
            version: g.version,
          });
        },
        error: (err: HttpErrorResponse) => {
          this.error.set(err.error.message);
        },
      });
  }

  private handleError(err: HttpErrorResponse) {
    this.submitting.set(false);

    if (err.status === 409) {
      this.error.set(
        'Esta meta foi alterada em outro lugar. Recarregue a página para ver os dados atuais.',
      );
      return;
    }

    if (err.status === 422 && err.error?.errors) {
      this.fieldErrors.set(err.error.errors);
      this.error.set(err.error.message ?? 'Erro de validação');
      return;
    }

    this.error.set(err.error?.message ?? 'Falha ao salvar meta');
  }
}
