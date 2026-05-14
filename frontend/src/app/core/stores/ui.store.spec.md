# SPEC: UiStore (Centralized UI State)

## Overview
Store centralizado para estado de UI de la aplicación. Usa Angular Signals para reactivity. Es un singleton injectable a nivel root.

## Ubicación
`frontend/src/app/core/stores/ui.store.ts`

## State Manageado

### 1. Notifications (cola de notificaciones)
```typescript
// Signal interno
private _notifications = signal<Notification[]>([])

// API pública (readonly)
notifications: Signal<Notification[]>
hasNotifications: Signal<boolean>
```

**Tipos de Notification:**
```typescript
interface Notification {
  id: string;           // UUID único
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message: string;
  duration?: number;    // ms, default 5000
  createdAt: Date;
}
```

### 2. Loading Global
```typescript
// Signal
private _isLoading = signal<boolean>(false)

// API pública
isLoading: Signal<boolean>
```

**Uso:** Actividades asincrónicas largas (llamadas API, carga de datos)

### 3. Sidebar
```typescript
// Signal
private _sidebarOpen = signal<boolean>(true)

// API pública
sidebarOpen: Signal<boolean>
```

### 4. Theme
```typescript
// Signal
private _theme = signal<'light' | 'dark'>('light)

// API pública
theme: Signal<'light' | 'dark'>
```

## Métodos

### Notifications
```typescript
// Mostrar notificación personalizada
showNotification(notification: Omit<Notification, 'id' | 'createdAt'>): void

// Métodos shortcuts
success(title: string, message: string): void
error(title: string, message: string): void    // duration default: 8000ms
warning(title: string, message: string): void
info(title: string, message: string): void

// Gestionar notificaciones
dismissNotification(id: string): void
clearAllNotifications(): void
```

### Loading
```typescript
setLoading(loading: boolean): void
```

### Sidebar
```typescript
toggleSidebar(): void
setSidebarOpen(open: boolean): void
```

### Theme
```typescript
setTheme(theme: 'light' | 'dark'): void
```

### Auth
```typescript
logout(): void  // Limpia tokens y redirige a /login
```

## Integración

### Con Error Interceptor
- El interceptor llama `showNotification()` automáticamente en errores HTTP

### Con Services
- Services injectan UiStore y usan:
  - `setLoading(true/false)` antes/después de operaciones async
  - `success/error/warning/info()` para feedback de operaciones CRUD

### Con Components
- Components inyectan UiStore para:
  - Leer `notifications()` para mostrar en ToastComponent
  - Leer `isLoading()` para loading overlay
  - Llamar métodos de acción cuando sea necesario

## Dependencias
- `@angular/core` - inject, signal, computed
- `@angular/router` - Router para logout

## Notas de Implementación
- El UiStore es un servicio root (`providedIn: 'root'`)
- Todas las propiedades públicas son readonly (uso de `asReadonly()`)
- Los métodos shortcuts usan durations defaults diferentes:
  - success: 5000ms
  - error: 8000ms (más visible)
  - warning: 5000ms
  - info: 5000ms

## Tests Esperados
- [ ] showNotification agrega item a la cola
- [ ] dismissNotification remueve item
- [ ] setLoading cambia el estado
- [ ] success/error/warning/info llaman showNotification
- [ ] logout redirige a /login y limpia localStorage
- [ ] auto-dismiss funciona después del duration