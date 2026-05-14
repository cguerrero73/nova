# SPEC: Loading Component (Global)

## Overview
Componente de overlay que muestra un spinner cuando hay operaciones asincrónicas en progreso. Se muestra en toda la aplicación.

## Ubicación
`frontend/src/app/shared/components/loading/loading.component.ts`

## Responsabilidades
- Mostrar overlay con spinner cuando `uiStore.isLoading()` es true
- Ocultar completamente cuando no hay loading
- Animación de fade-in suave

## Comportamiento

### Condición de Display
```typescript
@if (uiStore.isLoading()) {
  <!-- overlay visible -->
}
```

### Estructura Visual
```
loading-overlay (fixed, full screen)
  └── loading-spinner (centered, white bg, rounded)
      ├── spinner (animated circle)
      └── loading-text ("Cargando...")
```

### Animaciones
- Overlay: `opacity: 0` → `opacity: 1` (fade-in 200ms)
- Spinner: `transform: rotate(360deg)` loop 800ms linear infinite

### Estilos
- Overlay: `background: rgba(0, 0, 0, 0.4)` (semi-transparente)
- Spinner container: `background: white`, `border-radius: 12px`, `box-shadow`
- Spinner: 48x48px, 4px border, border-radius 50%

### Posicionamiento
- `position: fixed; top: 0; left: 0; width: 100%; height: 100%`
- `display: flex; align-items: center; justify-content: center`
- `z-index: 9998` (justo debajo del toast que es 9999)

## Integración con App
Se integra en `app.component.ts`:
```typescript
template: `
  <router-outlet />
  <app-toast />
  <app-loading />
`
```

## Flujo de Uso
1. Service llama `uiStore.setLoading(true)` antes de operación async
2. Loading overlay aparece automáticamente
3. Operation completa (success o error)
4. Service llama `uiStore.setLoading(false)`
5. Loading overlay desaparece

## Ejemplo de Uso
```typescript
// En un service
async loadUsers() {
  this.uiStore.setLoading(true);
  try {
    const data = await this.api.get('/users').toPromise();
    // procesar data
  } finally {
    this.uiStore.setLoading(false);
  }
}
```

## Dependencias
- `UiStore` - para leer isLoading()
- `CommonModule` - para @if
- `@angular/core` - Component, inject

## Diferencia con Loading Local
| Aspecto | Loading Global | Loading Local |
|---------|----------------|---------------|
| Ubicación | App level | Componente individual |
| Uso | Operaciones API generales | Loading específico de componente |
| Visión | Full screen overlay | Spinner en área específica |

## Tests Esperados
- [ ] Se muestra cuando isLoading() es true
- [ ] Se oculta cuando isLoading() es false
- [ ] Spinner anima correctamente
- [ ] Fade-in animation funciona
- [ ] No bloquea interacción con toast (z-index correcto)
- [ ] Texto "Cargando..." visible