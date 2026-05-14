# SPEC: Toast Component

## Overview
Componente visual para mostrar notificaciones/toasts en la aplicación. Renderiza la cola de notificaciones del UiStore.

## Ubicación
`frontend/src/app/shared/components/toast/toast.component.ts`

## Responsabilidades
- Renderizar todas las notificaciones activas del UiStore
- Mostrar ícono según tipo de notificación
- Permitir dismiss manual por click
- Auto-remover notificaciones después del duration
- Animación de entrada (slide from right)

## Inputs
No tiene inputs - lee directamente del UiStore:
```typescript
readonly uiStore = inject(UiStore);
```

## Tipos de Notificación y Estilos

| Tipo | Color Fondo | Color Borde | Ícono |
|------|-------------|-------------|-------|
| success | #ecfdf5 (green-50) | #10b981 | ✓ |
| error | #fef2f2 (red-50) | #ef4444 | ✕ |
| warning | #fffbeb (amber-50) | #f59e0b | ⚠ |
| info | #eff6ff (blue-50) | #3b82f6 | ℹ |

## Estructura del Template
```
toast-container (fixed, top-right)
  └── toast (por cada notification)
      ├── toast__icon (emoji según tipo)
      ├── toast__content
      │   ├── toast__title (bold)
      │   └── toast__message
      └── toast__close (× button)
```

## Comportamientos

### Renderizado
- Usa `@for` de Angular 17+ para iterar notifications
- `track notification.id` para eficiencia

### Click para Dismiss
- Click en todo el toast → dismiss
- Click en botón × → dismiss (con stopPropagation)

### Animación
- Slide-in desde la derecha: `transform: translateX(100%)` → `translateX(0)`
- Duración: 300ms, ease-out

### Posicionamiento
- `position: fixed; top: 1rem; right: 1rem`
- `z-index: 9999` (por sobre todo)
- `max-width: 400px`
- Flex column con gap de 0.5rem

## Integración con App
Se integra en `app.component.ts`:
```typescript
template: `
  <router-outlet />
  <app-toast />
  <app-loading />
`
```

El ToastComponent se suscribe automáticamente a `uiStore.notifications()`.

## Dependencias
- `UiStore` - para leer notificaciones
- `CommonModule` - para directivas Angular
- `@angular/core` - Component, inject

## Tests Esperados
- [ ] Muestra todas las notificaciones de la cola
- [ ] Estilos correctos según tipo (success/error/warning/info)
- [ ] Click en toast lo dismiss
- [ ] Click en × lo dismiss
- [ ] Animación slide-in funciona
- [ ] Posicionamiento fixed top-right
- [ ] Auto-dismiss después de duration (verificado en UiStore)