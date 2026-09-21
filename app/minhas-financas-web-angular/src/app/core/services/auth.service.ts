import { HttpClient, HttpErrorResponse } from "@angular/common/http";
import { computed, inject, Injectable, signal } from "@angular/core";
import { Router } from "@angular/router";
import { catchError, finalize, Observable, of, shareReplay, tap, throwError } from "rxjs";

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

  constructor() {
    this.http
      .get<{ authenticated: boolean }>('/api/auth/session')
      .pipe(catchError(() => of({ authenticated: false })))
      .subscribe((r) => this._isAuthenticated.set(r.authenticated))
  }

  login(email: string, password: string) {
    return this.http
      .post<OkResponse>('/api/auth', { email, password })
      .pipe(tap(() => this._isAuthenticated.set(true)));
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
