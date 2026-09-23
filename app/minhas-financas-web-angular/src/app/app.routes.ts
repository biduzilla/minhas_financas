import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () =>
      import('./features/auth/login/login.component').then((m) => m.LoginComponent),
  },
  {
    path: 'signup',
    loadComponent: () =>
      import('./features/auth/signup/signup.component').then((m) => m.SignupComponent),
  },

  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./shared/layout/shell.component').then((m) => m.ShellComponent),
    children: [
      { path: '', pathMatch: 'full', redirectTo: 'dashboard' },
      {
        path: 'dashboard',
        loadComponent: () =>
          import('./features/dashboard/dashboard.component').then((m) => m.DashboardComponent),
      },
      {
        path: 'transactions',
        loadComponent: () =>
          import(
            './features/transactions/transactions-list/transactions-list.component'
          ).then((m) => m.TransactionsListComponent),
      },
      {
        path: 'transactions/new',                 // 👈 antes do :id
        loadComponent: () =>
          import(
            './features/transactions/transaction-form/transaction-form.component'
          ).then((m) => m.TransactionFormComponent),
      },
      {
        path: 'transactions/:id',
        loadComponent: () =>
          import(
            './features/transactions/transaction-form/transaction-form.component'
          ).then((m) => m.TransactionFormComponent),
      },
      {
        path: 'categories',
        loadComponent: () =>
          import(
            './features/categories/categories-list/categories-list.component'
          ).then((m) => m.CategoriesListComponent),
      },
      {
        path: 'categories/new',
        loadComponent: () =>
          import(
            './features/categories/category-form/category-form.component'
          ).then((m) => m.CategoryFormComponent),
      },
      {
        path: 'categories/:id',
        loadComponent: () =>
          import(
            './features/categories/category-form/category-form.component'
          ).then((m) => m.CategoryFormComponent),
      },
    ],
  },

  { path: '**', redirectTo: 'dashboard' },
];
