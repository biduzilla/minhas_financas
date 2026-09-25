import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { catchError, throwError } from 'rxjs';

const FALLBACK_BY_STATUS: Record<number, string> = {
  0: 'Sem conexão com o servidor. Verifique sua internet.',
  400: 'Requisição inválida.',
  401: 'Sua sessão expirou. Faça login novamente.',
  403: 'Você não tem permissão para essa ação.',
  404: 'Recurso não encontrado.',
  409: 'Os dados foram alterados por outra pessoa. Recarregue a página.',
  422: 'Dados inválidos. Verifique os campos.',
  429: 'Muitas requisições. Aguarde um momento.',
  500: 'Erro interno do servidor. Tente novamente.',
  502: 'Servidor temporariamente indisponível.',
  503: 'Serviço indisponível. Tente novamente em instantes.',
  504: 'O servidor demorou para responder. Tente novamente.',
};

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  return next(req).pipe(
    catchError((err: HttpErrorResponse) => {
      if (err.error?.message) {
        return throwError(() => err);
      }

      const fallback =
        FALLBACK_BY_STATUS[err.status] ??
        (err.status >= 500 ? 'Erro no servidor.' : 'Erro inesperado.');

      const originalBody =
        err.error && typeof err.error === 'object' ? err.error : {};

      Object.assign(err, {
        error: {
          ...originalBody,
          message: fallback,
        },
      });

      return throwError(() => err);
    }),
  );
};
