import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { finalize } from 'rxjs';
import { form, FormField, required, minLength, maxLength } from '@angular/forms/signals';

import { CategoryService } from '../../../core/services/category.service';
import { CategoryType } from '../../../core/models/api.types';

interface CategoryFormData {
  name: string;
  type: CategoryType;
  version?: number;
}

@Component({
  selector: 'app-category-form',
  standalone: true,
  imports: [FormField, RouterLink],
  templateUrl: './category-form.component.html',
})
export class CategoryFormComponent {
  private categoryService = inject(CategoryService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private destroyRef = inject(DestroyRef);

  readonly categoryId = signal<string | null>(null);
  readonly isEdit = computed(() => this.categoryId() !== null);

  readonly model = signal<CategoryFormData>({
    name: '',
    type: 'output',
  });

  readonly categoryForm = form(this.model, (path) => {
    required(path.name, { message: 'Nome é obrigatório' });
    minLength(path.name, 3, { message: 'Nome deve ter pelo menos 3 caracteres' });
    maxLength(path.name, 100, { message: 'Nome deve ter no máximo 100 caracteres' });
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
        this.categoryId.set(id);
        if (id) this.loadCategory(id);
      });
  }

  onSubmit() {
    if (this.categoryForm().invalid()) return;

    this.submitting.set(true);
    this.error.set(null);
    this.fieldErrors.set({});

    const data = this.model();
    const id = this.categoryId();

    const request$ = id
      ? this.categoryService.update(id, data)
      : this.categoryService.create(data);

    request$
      .pipe(finalize(() => this.submitting.set(false)))
      .subscribe({
        next: () => this.router.navigate(['/categories']),
        error: (err: HttpErrorResponse) => {
          if (err.status === 422 && err.error?.errors) {
            this.fieldErrors.set(err.error.errors);
            this.error.set(err.error.message ?? 'Erro de validação');
          } else {
            this.error.set(err.error?.message ?? 'Falha ao salvar categoria');
          }
        },
      });
  }

  backendError(field: keyof CategoryFormData): string | null {
    return this.fieldErrors()[field] ?? null;
  }

  private loadCategory(id: string) {
    this.loading.set(true);
    this.error.set(null);

    this.categoryService
      .getById(id)
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (c) => {
          this.model.set({
            name: c.name ?? '',
            type: (c.type ?? 'output') as CategoryType,
            version: c.version,
          });
        },
        error: (err: HttpErrorResponse) => {
          this.error.set(
            err.status === 404
              ? 'Categoria não encontrada'
              : err.error?.message ?? 'Erro ao carregar categoria',
          );
        },
      });
  }
}
