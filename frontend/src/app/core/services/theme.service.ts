import { Injectable, signal, effect, inject } from '@angular/core';
import { DOCUMENT } from '@angular/common';
import { ThemeName, Density, THEMES, DENSITIES, ThemeConfig, DensityConfig } from '../models/theme.model';

const THEME_KEY = 'nova_theme';
const DENSITY_KEY = 'nova_density';

// Helper para convertir hex a HSL
function hexToHsl(hex: string): string {
  // Quitar # si existe
  hex = hex.replace('#', '');
  
  const r = parseInt(hex.substring(0, 2), 16) / 255;
  const g = parseInt(hex.substring(2, 4), 16) / 255;
  const b = parseInt(hex.substring(4, 6), 16) / 255;

  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  let h = 0, s = 0;
  const l = (max + min) / 2;

  if (max !== min) {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    switch (max) {
      case r: h = ((g - b) / d + (g < b ? 6 : 0)) / 6; break;
      case g: h = ((b - r) / d + 2) / 6; break;
      case b: h = ((r - g) / d + 4) / 6; break;
    }
  }

  return `${Math.round(h * 360)} ${Math.round(s * 100)}% ${Math.round(l * 100)}%`;
}

@Injectable({ providedIn: 'root' })
export class ThemeService {
  private readonly document = inject(DOCUMENT);

  // Signals
  private _theme = signal<ThemeName>(this.loadTheme());
  private _density = signal<Density>(this.loadDensity());

  // Readonly signals
  readonly theme = this._theme.asReadonly();
  readonly density = this._density.asReadonly();

  // Computed
  readonly currentTheme = () => THEMES[this._theme()];
  readonly currentDensity = () => DENSITIES[this._density()];

  constructor() {
    // Apply theme on init
    this.applyTheme(THEMES[this._theme()]);
    this.applyDensity(DENSITIES[this._density()]);

    // React to theme changes
    effect(() => {
      this.applyTheme(THEMES[this._theme()]);
      this.saveTheme(this._theme());
    });

    // React to density changes
    effect(() => {
      this.applyDensity(DENSITIES[this._density()]);
      this.saveDensity(this._density());
    });
  }

  setTheme(theme: ThemeName): void {
    this._theme.set(theme);
  }

  setDensity(density: Density): void {
    this._density.set(density);
  }

  toggleTheme(): void {
    const themes: ThemeName[] = ['light', 'dark', 'blue', 'green'];
    const currentIndex = themes.indexOf(this._theme());
    const nextIndex = (currentIndex + 1) % themes.length;
    this._theme.set(themes[nextIndex]);
  }

  private applyTheme(theme: ThemeConfig): void {
    const root = this.document.documentElement;
    
    // Convertir colores a HSL para Tailwind
    const backgroundHsl = hexToHsl(theme.colors.background);
    const foregroundHsl = hexToHsl(theme.colors.text);
    const borderHsl = hexToHsl(theme.colors.border);
    const primaryHsl = hexToHsl(theme.colors.primary);
    
    // Set CSS variables in HSL format for Tailwind
    root.style.setProperty('--background', backgroundHsl);
    root.style.setProperty('--foreground', foregroundHsl);
    root.style.setProperty('--border', borderHsl);
    root.style.setProperty('--primary', primaryHsl);
    root.style.setProperty('--primary-foreground', '0 0% 100%');
    
    // Also set color-* variables for direct use
    root.style.setProperty('--color-primary', theme.colors.primary);
    root.style.setProperty('--color-primary-hover', theme.colors.primaryHover);
    root.style.setProperty('--color-background', theme.colors.background);
    root.style.setProperty('--color-surface', theme.colors.surface);
    root.style.setProperty('--color-border', theme.colors.border);
    root.style.setProperty('--color-text', theme.colors.text);
    root.style.setProperty('--color-text-secondary', theme.colors.textSecondary);

    // Set data attribute for CSS selectors
    root.setAttribute('data-theme', theme.name);
    
    // Apply dark mode class based on theme
    if (theme.name === 'dark') {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
  }

  private applyDensity(density: DensityConfig): void {
    const root = this.document.documentElement;
    
    const spacingMap = { sm: '0.5rem', md: '1rem', lg: '1.5rem' };
    const fontSizeMap = { sm: '0.875rem', md: '1rem', lg: '1.125rem' };
    
    root.style.setProperty('--spacing-base', spacingMap[density.spacing]);
    root.style.setProperty('--font-size-base', fontSizeMap[density.fontSize]);
    
    root.setAttribute('data-density', density.name);
  }

  private loadTheme(): ThemeName {
    const stored = localStorage.getItem(THEME_KEY);
    if (stored && stored in THEMES) {
      return stored as ThemeName;
    }
    // Default to light
    return 'light';
  }

  private loadDensity(): Density {
    const stored = localStorage.getItem(DENSITY_KEY);
    if (stored && stored in DENSITIES) {
      return stored as Density;
    }
    return 'compact';
  }

  private saveTheme(theme: ThemeName): void {
    localStorage.setItem(THEME_KEY, theme);
  }

  private saveDensity(density: Density): void {
    localStorage.setItem(DENSITY_KEY, density);
  }
}