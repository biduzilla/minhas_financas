import { Component, inject, signal } from "@angular/core";
import { email, form, FormField, maxLength, minLength, required } from "@angular/forms/signals";
import { Router, RouterLink } from "@angular/router";
import { AuthService } from "../../../core/services/auth.service";
import { HttpErrorResponse } from "@angular/common/http";

interface SignupData {
  name: string;
  email: string;
  password: string;
}

@Component({
  selector: 'app-signup',
  standalone: true,
  imports: [FormField, RouterLink],
  templateUrl: './signup.component.html',
})
export class SignupComponent {
  private auth = inject(AuthService)
  private router = inject(Router)

  signupModel = signal<SignupData>({
    name: '',
    email: '',
    password: '',
  })

  signupForm = form(this.signupModel, (schemaPath) => {
    required(schemaPath.name, { message: 'Nome é obrigatório' });
    minLength(schemaPath.name, 3, { message: 'Nome deve ter pelo menos 3 caracteres' });
    maxLength(schemaPath.name, 100, { message: 'Nome deve ter no máximo 100 caracteres' });

    required(schemaPath.email, { message: 'Email é obrigatório' });
    email(schemaPath.email, { message: 'Email inválido' });

    required(schemaPath.password, { message: 'Senha é obrigatória' });
    minLength(schemaPath.password, 8, { message: 'Senha deve ter pelo menos 8 caracteres' })
  })

  isPending = signal(false)
  error = signal<string | null>(null)
  fieldErrors = signal<Record<string, string>>({})

  onSubmit() {
    if (this.signupForm().invalid()) return

    const data = this.signupModel()
    this.isPending.set(true);
    this.fieldErrors.set({});
    this.error.set(null);

    this.auth.signup(data).subscribe({
      next: () => {
        this.router.navigate(['/login'], {
          queryParams: { registed: '1' },
        });
      },
      error: (err: HttpErrorResponse) => {
        this.isPending.set(false);
        if (err.status === 422 && err.error?.errors) {
          this.fieldErrors.set(err.error.errors);
          this.error.set(err.error.message ?? 'Erro de validação');
        } else {
          this.error.set(err.error?.message ?? 'Falha ao criar conta');
        }
      },
      complete: () => this.isPending.set(false)
    })
  }

  fieldError(field: keyof SignupData): string | null {
    const backend = this.fieldErrors()[field]
    if (backend) return backend
    return null
  }
}
