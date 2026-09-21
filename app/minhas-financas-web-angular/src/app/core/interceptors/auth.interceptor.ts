import { HttpErrorResponse, HttpInterceptorFn } from "@angular/common/http";
import { inject } from "@angular/core";
import { catchError, switchMap, throwError } from "rxjs";
import { AuthService } from "../services/auth.service";

function isAuthEndpoint(url: string): boolean {
  return (
    url.includes('/api/auth') ||        // login
    url.includes('/api/auth/refresh') ||
    url.includes('/api/auth/logout') ||
    url.includes('/api/auth/session')
  );
}

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService)
  const cloned = req.clone({ withCredentials: true })

  return next(cloned).pipe(
    catchError((err: HttpErrorResponse) => {
      if (err.status === 401 && !isAuthEndpoint(req.url)) {
        return auth.refreshToken().pipe(
          switchMap(() => next(cloned)),
          catchError(() => {
            auth.logout();
            return throwError(() => err);
          })
        )
      }
      return throwError(() => err);
    })
  );
};
