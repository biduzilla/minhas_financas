import { Component, computed, DestroyRef, inject, signal } from "@angular/core";
import { form, FormField, maxLength, min, required } from "@angular/forms/signals";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
import { TransactionService } from "../../../core/services/transaction.service";
import { CategoryService } from "../../../core/services/category.service";
import { Category } from "../../../core/models/api.types";
import { HttpErrorResponse } from "@angular/common/http";
import { finalize } from "rxjs";
import { takeUntilDestroyed } from "@angular/core/rxjs-interop";

interface TransactionFormData {
  amount: number | null;
  category_id: string;
  description: string;
  version?: number;
}

@Component({
  selector: 'app-transaction-form',
  standalone: true,
  imports: [FormField, RouterLink],
  templateUrl: './transaction-form.component.html',
})
export class TransactionFormComponent {
  private txService = inject(TransactionService)
  private categoryService = inject(CategoryService)
  private route = inject(ActivatedRoute)
  private router = inject(Router)
  private destroyRef = inject(DestroyRef)

  readonly transactionId = signal<string | null>(null)
  readonly isEdit = computed(() => this.transactionId() !== null)

  readonly model = signal<TransactionFormData>({
    amount: null,
    category_id: '',
    description: '',
  })

  readonly txForm = form(this.model, (path) => {
    required(path.amount, { message: 'Valor é obrigatório' });
    min(path.amount, 0.01, { message: 'Valor deve ser maior que zero' });

    required(path.category_id, { message: 'Categoria é obrigatória' });

    maxLength(path.description, 100, {
      message: 'Descrição deve ter no máximo 100 caracteres',
    });
  });

  readonly categories = signal<Category[]>([]);
  readonly loading = signal(false);
  readonly submitting = signal(false);
  readonly error = signal<string | null>(null);
  readonly fieldErrors = signal<Record<string, string>>({});

  readonly inputCategories = computed(() =>
    this.categories().filter((c) => c.type === 'input'))
  readonly outputCategories = computed(() =>
    this.categories().filter((c) => c.type === 'output'))

  constructor() {
    this.loadCategories()

    this.route.paramMap
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((params) => {
        const id = params.get('id')
        this.transactionId.set(id)
        if (id) this.loadTransaction(id)
      })
  }

  onSubmit() {
    if (this.txForm().invalid()) return

    this.submitting.set(true)
    this.error.set(null)
    this.fieldErrors.set({})

    const data = this.model()
    const id = this.transactionId()

    const payload = {
      amount: data.amount!,
      category_id: data.category_id,
      description: data.description,
    };

    const request$ = id
      ? this.txService.update(id, { ...payload, version: data.version ?? 0 })
      : this.txService.create(payload)

    request$
      .pipe(finalize(() => this.submitting.set(false)))
      .subscribe({
        next: () => this.router.navigate(['/transactions']),
        error: (err: HttpErrorResponse) => this.handleError(err),
      })
  }

  backendError(field: keyof TransactionFormData): string | null {
    return this.fieldErrors()[field] ?? null
  }

  private loadTransaction(id: string) {
    this.loading.set(true)
    this.error.set(null)

    this.txService
      .getById(id)
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (t) => {
          this.model.set({
            amount: t.amount,
            category_id: t.category_id,
            description: t.description,
            version: t.version,
          });
        },
        error: (err: HttpErrorResponse) => {
          this.error.set(
            err.status === 404
              ? 'Transação não encontrada'
              : err.error?.message ?? 'Erro ao carregar transação',
          );
        },
      })
  }

  private loadCategories() {
    this.categoryService.list({ page_size: 100 }).subscribe({
      next: (res) => this.categories.set(res.content),
      error: () => {
        this.categories.set([])
      }
    })
  }

  private handleError(err: HttpErrorResponse) {
    this.submitting.set(false)

    if (err.status === 409) {
      this.error.set(
        'Esta transação foi alterada em outro lugar. Recarregue a página para ver os dados atuais.',
      );
      return;
    }

    if (err.status === 422 && err.error?.errors) {
      this.fieldErrors.set(err.error.errors);
      this.error.set(err.error.message ?? 'Erro de validação');
      return;
    }

    if (err.status === 404) {
      this.error.set('Categoria não encontrada. Escolha outra.');
      return;
    }

    this.error.set(err.error?.message ?? 'Falha ao salvar transação');
  }
}
