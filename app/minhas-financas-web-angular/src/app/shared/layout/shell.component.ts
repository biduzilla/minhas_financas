import { Component, inject } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [RouterLink, RouterLinkActive, RouterOutlet],
  template: `
    <div class="min-h-screen bg-slate-50">
      <nav class="border-b border-slate-200 bg-white">
        <div class="mx-auto flex max-w-6xl items-center gap-6 px-6 py-3">
          <a
            routerLink="/dashboard"
            routerLinkActive="text-emerald-600"
            class="text-sm font-medium text-slate-600 transition-colors hover:text-slate-900"
          >
            Dashboard
          </a>
          <a
            routerLink="/transactions"
            routerLinkActive="text-emerald-600"
            class="text-sm font-medium text-slate-600 transition-colors hover:text-slate-900"
          >
            Transações
          </a>
          <a
            routerLink="/categories"
            routerLinkActive="text-emerald-600"
            class="text-sm font-medium text-slate-600 transition-colors hover:text-slate-900"
          >
            Categorias
          </a>
          <a
            routerLink="/goals"
            routerLinkActive="text-emerald-600"
            class="text-sm font-medium text-slate-600 transition-colors hover:text-slate-900"
          >
            Metas
          </a>
          <button
      type="button"
      (click)="logout()"
      class="ml-auto cursor-pointer text-sm font-medium text-slate-500 hover:text-slate-900"
    >
      Sair
    </button>
        </div>
      </nav>

      <main>
        <router-outlet />
      </main>
    </div>
  `,
})
export class ShellComponent {
  private auth = inject(AuthService);
  isAuthenticated = this.auth.isAuthenticated;

  logout() {
    this.auth.logout();
  }
}
