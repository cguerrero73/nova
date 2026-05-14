export type ThemeName = 'light' | 'dark' | 'blue' | 'green';
export type Density = 'minimal' | 'compact' | 'enterprise';

export interface ThemeConfig {
  name: ThemeName;
  label: string;
  colors: {
    primary: string;
    primaryHover: string;
    background: string;
    surface: string;
    border: string;
    text: string;
    textSecondary: string;
  };
}

export interface DensityConfig {
  name: Density;
  label: string;
  spacing: 'sm' | 'md' | 'lg';
  fontSize: 'sm' | 'md' | 'lg';
}

export interface UserPreferences {
  theme: ThemeName;
  density: Density;
}

export const THEMES: Record<ThemeName, ThemeConfig> = {
  light: {
    name: 'light',
    label: 'Claro',
    colors: {
      primary: '#2563eb',
      primaryHover: '#1d4ed8',
      background: '#f9fafb',
      surface: '#ffffff',
      border: '#e5e7eb',
      text: '#111827',
      textSecondary: '#6b7280',
    },
  },
  dark: {
    name: 'dark',
    label: 'Oscuro',
    colors: {
      primary: '#3b82f6',
      primaryHover: '#60a5fa',
      background: '#111827',
      surface: '#1f2937',
      border: '#374151',
      text: '#f9fafb',
      textSecondary: '#9ca3af',
    },
  },
  blue: {
    name: 'blue',
    label: 'Azul',
    colors: {
      primary: '#0284c7',
      primaryHover: '#0369a1',
      background: '#f0f9ff',
      surface: '#ffffff',
      border: '#bae6fd',
      text: '#0c4a6e',
      textSecondary: '#075985',
    },
  },
  green: {
    name: 'green',
    label: 'Verde',
    colors: {
      primary: '#16a34a',
      primaryHover: '#15803d',
      background: '#f0fdf4',
      surface: '#ffffff',
      border: '#bbf7d0',
      text: '#14532d',
      textSecondary: '#166534',
    },
  },
};

export const DENSITIES: Record<Density, DensityConfig> = {
  minimal: {
    name: 'minimal',
    label: 'Minimalista',
    spacing: 'sm',
    fontSize: 'sm',
  },
  compact: {
    name: 'compact',
    label: 'Compacto',
    spacing: 'md',
    fontSize: 'md',
  },
  enterprise: {
    name: 'enterprise',
    label: 'Enterprise',
    spacing: 'lg',
    fontSize: 'lg',
  },
};