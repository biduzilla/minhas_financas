import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { computed, inject, Injectable, signal } from "@angular/core";
import { Router } from "@angular/router";
import { catchError, finalize, Observable, of, shareReplay, tap, throwError } from "rxjs";
import { SignUpInput, User } from "../models/api.types";

interface OkResponse {
  ok: boolean;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private http = inject(HttpClient);
  private router = inject(Router);

  private readonly _isAuthenticated = signal(false)
  readonly isAuthenticated = computed(() => this._isAuthenticated());
  private refresh$?: Observable<OkResponse>

  init(): Observable<{ authenticated: boolean }> {
    return this.http
      .get<{ authenticated: boolean }>('/api/auth/session')
      .pipe(
        tap((r) => this._isAuthenticated.set(r.authenticated)),
        catchError(() => {
          this._isAuthenticated.set(false);
          return of({ authenticated: false });
        }),
      );
  }

  login(email: string, password: string) {
    return this.http
      .post<OkResponse>('/api/auth', { email, password })
      .pipe(tap(() => this._isAuthenticated.set(true)));
  }

  signup(input: SignUpInput) {
    return this.http.post<User>('/api/users', input)
  }

  refreshToken(): Observable<OkResponse> {
    this.refresh$ ??= this.http
      .post<OkResponse>('/api/auth/refresh', {})
      .pipe(
        finalize(() => (this.refresh$ = undefined)),
        shareReplay({ bufferSize: 1, refCount: false }),
      );
    return this.refresh$;
  }

  logout() {
    this.http.post('/api/auth/logout', {}).subscribe({
      complete: () => {
        this._isAuthenticated.set(false);
        this.router.navigate(['/login']);
      },
      error: () => {
        this._isAuthenticated.set(false);
        this.router.navigate(['/login']);
      },
    });
  }
}
