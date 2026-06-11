import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, throwError } from 'rxjs';
import { UiStore } from '../stores/ui.store';

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const uiStore = inject(UiStore);

  return next(req).pipe(
    catchError((error: HttpErrorResponse) => {
      let errorMessage = 'Error inesperado';
      let errorType: 'error' | 'warning' | 'info' = 'error';

      // Extract server message supporting both formats:
      // - flat:  {code: "...", message: "..."}        (AppError)
      // - wrapped: {success: false, error: {code: "...", message: "..."}} (customErrorHandler)
      const serverData = error.error;
      let serverMsg = '';

      // Handle wrapped format: {success: false, error: {code: "...", message: "..."}}
      if (serverData && typeof serverData === 'object' && 'error' in serverData) {
        const inner = (serverData as any).error;
        serverMsg = inner?.message || '';
      }
      // Handle flat format: {code: "...", message: "..."}
      if (!serverMsg && serverData && typeof serverData === 'object') {
        serverMsg = (serverData as any).message || '';
      }

      if (error.error instanceof ErrorEvent) {
        // Error del cliente (network, etc)
        errorMessage = error.error.message;
      } else {
        // Errores del servidor
        switch (error.status) {
          case 0:
            errorMessage = 'No se puede conectar al servidor';
            errorType = 'warning';
            break;
          case 400:
            errorMessage = serverMsg || 'Solicitud inválida';
            break;
          case 401:
            errorMessage = serverMsg || 'Sesión expirada. Por favor inicie sesión nuevamente';
            uiStore.logout();
            break;
          case 403:
            errorMessage = serverMsg || 'No tiene permisos para realizar esta acción';
            break;
          case 404:
            errorMessage = serverMsg || 'Recurso no encontrado';
            break;
          case 409:
            errorMessage = serverMsg || 'Conflicto de datos';
            break;
          case 422:
            errorMessage = serverMsg || 'Datos inválidos';
            break;
          case 429:
            errorMessage = serverMsg || 'Demasiadas solicitudes. Intente más tarde';
            errorType = 'warning';
            break;
          case 500:
            errorMessage = serverMsg || 'Error interno del servidor';
            break;
          case 502:
          case 503:
          case 504:
            errorMessage = serverMsg || 'Servicio no disponible. Intente más tarde';
            errorType = 'warning';
            break;
          default:
            errorMessage = serverMsg || `Error ${error.status}`;
        }
      }

      // Mostrar toast de error
      uiStore.showNotification({
        type: errorType,
        title: `Error ${error.status || ''}`,
        message: errorMessage,
      });

      return throwError(() => error);
    })
  );
};
