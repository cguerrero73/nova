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
            errorMessage = error.error?.error?.message || 'Solicitud inválida';
            break;
          case 401:
            errorMessage = 'Sesión expirada. Por favor inicie sesión nuevamente';
            uiStore.logout();
            break;
          case 403:
            errorMessage = 'No tiene permisos para realizar esta acción';
            break;
          case 404:
            errorMessage = 'Recurso no encontrado';
            break;
          case 409:
            errorMessage = error.error?.error?.message || 'Conflicto de datos';
            break;
          case 422:
            errorMessage = error.error?.error?.message || 'Datos inválidos';
            break;
          case 429:
            errorMessage = 'Demasiadas solicitudes. Intente más tarde';
            errorType = 'warning';
            break;
          case 500:
            errorMessage = 'Error interno del servidor';
            break;
          case 502:
          case 503:
          case 504:
            errorMessage = 'Servicio no disponible. Intente más tarde';
            errorType = 'warning';
            break;
          default:
            errorMessage = error.error?.error?.message || `Error ${error.status}`;
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