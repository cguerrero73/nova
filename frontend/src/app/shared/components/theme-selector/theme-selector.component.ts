import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ThemeService } from '@core/services/theme.service';
import { ThemeName, Density, THEMES, DENSITIES, ThemeConfig } from '@core/models/theme.model';

@Component({
  selector: 'app-theme-selector',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './theme-selector.component.html',
  styleUrl: './theme-selector.component.css',
})
export class ThemeSelectorComponent {
  private readonly themeService = inject(ThemeService);

  // Current values from service - obtener el valor directamente
  get currentThemeValue(): ThemeConfig {
    return this.themeService.currentTheme();
  }
  
  get currentDensityValue(): string {
    return this.themeService.currentDensity().name;
  }

  // Available options
  themes = Object.values(THEMES);
  densities = Object.values(DENSITIES);

  // UI state
  isOpen = signal(false);

  onThemeChange(theme: ThemeName): void {
    this.themeService.setTheme(theme);
  }

  onDensityChange(density: Density): void {
    this.themeService.setDensity(density);
  }

  toggleDropdown(): void {
    this.isOpen.set(!this.isOpen());
  }

  closeDropdown(): void {
    this.isOpen.set(false);
  }
}