import { Component, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink, RouterOutlet } from '@angular/router';
import { AuthService } from '@core/services/auth.service';
import { ThemeService } from '@core/services/theme.service';
import { ThemeSelectorComponent } from '../theme-selector/theme-selector.component';

@Component({
  selector: 'app-layout',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterOutlet, ThemeSelectorComponent],
  templateUrl: './layout.component.html',
  styleUrl: './layout.component.css',
})
export class LayoutComponent {
  private readonly authService = inject(AuthService);
  private readonly themeService = inject(ThemeService);
  
  user = this.authService.user;
  isAuthenticated = this.authService.isAuthenticated;
  isAdmin = this.authService.isAdmin;
  
  // Padding basado en densidad
  mainPadding = computed(() => {
    const density = this.themeService.currentDensity();
    const basePadding = '1rem';
    switch (density.name) {
      case 'minimal': return '0.5rem';
      case 'compact': return '1rem';
      case 'enterprise': return '1.5rem';
      default: return basePadding;
    }
  });
  
  showUserMenu = signal(false);
  showSidebar = signal(false);
  showKeyboardShortcuts = signal(false);
  sidebarCollapsed = signal(false);
  
  // Para el header: mostrar estado de la API
  apiStatus = signal<'connected' | 'disconnected'>('connected');

  toggleUserMenu(): void {
    this.showUserMenu.update(v => !v);
  }

  toggleSidebar(): void {
    this.showSidebar.update(v => !v);
  }

  toggleSidebarCollapse(): void {
    this.sidebarCollapsed.update(v => !v);
  }

  logout(): void {
    this.authService.logout();
  }
}
