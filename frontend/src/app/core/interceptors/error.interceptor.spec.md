# SPEC: Error Interceptor

## Overview
Manejo centralizado de errores HTTP en toda la aplicación. Interceptor de Angular que captura respuestas de error y toma acciones apropiadas.

## Ubicación
`frontend/src/app/core/interceptors/error.interceptor.ts`

## Responsabilidades
- Capturar todos los errores HTTP de respuestas de API
- Mostrar notificaciones toast según el tipo de error
- Redirigir a login en caso de 401 (sesión expirada)
- Mapear códigos de error HTTP a mensajes amigables

## Comportamiento por Código HTTP

| Status | Tipo Notificación | Acción Adicional |
|--------|-------------------|------------------|
| 0 (network error) | warning | "No se puede conectar al servidor" |
| 400 | error | "Solicitud inválida" |
| 401 | error | Redirect a /login, limpiar tokens |
| 403 | error | "No tiene permisos para realizar esta acción" |
| 404 | error | "Recurso no encontrado" |
| 409 | error | "Conflicto de datos" |
| 422 | error | "Datos inválidos" |
| 429 | warning | "Demasiadas solicitudes. Intente más tarde" |
| 500 | error | "Error interno del servidor" |
| 502, 503, 504 | warning | "Servicio no disponible. Intente más tarde" |
| Default | error | "Error {status}" |

## Integración con UiStore
- Usa `uiStore.showNotification()` para mostrar errores
- El tipo de notificación varía según el código (error/warning)

## Ejemplo de Uso
```typescript
// No requiere uso directo - funciona automáticamente
// Cualquier error HTTP en cualquier servicio mostrará un toast
const response = await api.get('/users'); // si falla, muestra toast automáticamente
```

## Errores del Cliente (ErrorEvent)
- Captura errores de red, timeouts, etc.
- Muestra el mensaje del error original

## Dependencias
- `UiStore` - para mostrar notificaciones
- `HttpInterceptorFn` - interfaz de interceptor de Angular

## Tests Esperados
- [ ] Código 401 redirige a /login
- [ ] Código 500 muestra toast de error
- [ ] Error de red muestra toast de warning
- [ ] Otros códigos muestran mensajes apropiados