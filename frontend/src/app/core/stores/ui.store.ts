import { Injectable, signal, computed, inject } from '@angular/core';
import { Router } from '@angular/router';

export interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message: string;
  duration?: number;
  createdAt: Date;
}

@Injectable({ providedIn: 'root' })
export class UiStore {
  private readonly router = inject(Router);

  // Notifications
  private _notifications = signal<Notification[]>([]);
  readonly notifications = this._notifications.asReadonly();
  readonly hasNotifications = computed(() => this._notifications().length > 0);

  // Loading global
  private _isLoading = signal<boolean>(false);
  readonly isLoading = this._isLoading.asReadonly();

  // Sidebar state
  private _sidebarOpen = signal<boolean>(true);
  readonly sidebarOpen = this._sidebarOpen.asReadonly();

  // Theme actual (viene del ThemeService)
  private _theme = signal<'light' | 'dark'>('light');
  readonly theme = this._theme.asReadonly();

  constructor() {}

  // ============ NOTIFICATIONS ============

  showNotification(notification: Omit<Notification, 'id' | 'createdAt'>) {
    const id = crypto.randomUUID();
    const newNotification: Notification = {
      ...notification,
      id,
      createdAt: new Date(),
      duration: notification.duration ?? 5000,
    };

    this._notifications.update((current) => [...current, newNotification]);

    // Auto-remover después de duration
    if (newNotification.duration && newNotification.duration > 0) {
      setTimeout(() => {
        this.dismissNotification(id);
      }, newNotification.duration);
    }
  }

  dismissNotification(id: string) {
    this._notifications.update((current) =>
      current.filter((n) => n.id !== id)
    );
  }

  clearAllNotifications() {
    this._notifications.set([]);
  }

  // Shortcuts
  success(title: string, message: string) {
    this.showNotification({ type: 'success', title, message });
  }

  error(title: string, message: string) {
    this.showNotification({ type: 'error', title, message, duration: 8000 });
  }

  warning(title: string, message: string) {
    this.showNotification({ type: 'warning', title, message });
  }

  info(title: string, message: string) {
    this.showNotification({ type: 'info', title, message });
  }

  // ============ LOADING ============

  setLoading(loading: boolean) {
    this._isLoading.set(loading);
  }

  // ============ SIDEBAR ============

  toggleSidebar() {
    this._sidebarOpen.update((current) => !current);
  }

  setSidebarOpen(open: boolean) {
    this._sidebarOpen.set(open);
  }

  // ============ THEME ============

  setTheme(theme: 'light' | 'dark') {
    this._theme.set(theme);
  }

  // ============ AUTH ============

  logout() {
    // Limpiar tokens y redirigir a login
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
    this.router.navigate(['/login']);
  }
}