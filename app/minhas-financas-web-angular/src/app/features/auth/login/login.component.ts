import { Component, inject, signal, WritableSignal } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { AuthService } from '../../../core/services/auth.service';
import { form, FormField, required, email } from '@angular/forms/signals';

interface LoginData {
  email: string;
  password: string;
}

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormField,RouterLink],
  templateUrl: './login.component.html'
})
export class LoginComponent {
  private auth = inject(AuthService);
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  justRegistered = signal(false)

  constructor() {
    this.route.queryParams.subscribe(params => {
      this.justRegistered.set(params['registered'] === '1')
    })
  }

  loginModel = signal<LoginData>({
    email: '',
    password: '',
  });

  loginForm = form(this.loginModel, (schemaPath) => {
    required(schemaPath.email, { message: 'Email é obrigatório' });
    email(schemaPath.email, { message: 'Email inválido' });
    required(schemaPath.password, { message: 'Senha é obrigatória' });
  });

  isPending = signal(false);
  error = signal<string | null>(null);
  fieldErrors = signal<Record<string, string>>({});

  onSubmit() {
    if (this.loginForm().invalid()) return;

    const { email: userEmail, password } = this.loginModel();
    this.isPending.set(true);
    this.fieldErrors.set({});
    this.error.set(null);

    this.auth.login(userEmail, password)
      .subscribe({
        next: () => this.router.navigate(['/dashboard']),
        error: (err: HttpErrorResponse) => {
          this.isPending.set(false);
          if (err.status === 422 && err.error?.errors) {
            this.fieldErrors.set(err.error.errors);
            this.error.set(err.error.message);
          } else {
            this.error.set(err.error.message);
          }
        },
        complete: () => this.isPending.set(false),
      });
  }
}

