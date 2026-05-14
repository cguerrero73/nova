import { Injectable, signal } from '@angular/core';

export const SUPPORTED_LANGUAGES = ['en', 'es', 'pt', 'fr', 'it'] as const;
export type SupportedLanguage = typeof SUPPORTED_LANGUAGES[number];

export const DEFAULT_LANGUAGE: SupportedLanguage = 'en';

export const LANGUAGE_STORAGE_KEY = 'nova_language';

@Injectable({ providedIn: 'root' })
export class LanguageDetector {
  private _currentLanguage = signal<SupportedLanguage>(this.detectInitialLanguage());
  readonly currentLanguage = this._currentLanguage.asReadonly();

  private detectInitialLanguage(): SupportedLanguage {
    // 1. Check localStorage
    const stored = localStorage.getItem(LANGUAGE_STORAGE_KEY);
    if (stored && this.isValidLanguage(stored)) {
      return stored as SupportedLanguage;
    }

    // 2. Check browser language
    const browserLang = navigator.language.split('-')[0].toLowerCase();
    if (this.isValidLanguage(browserLang)) {
      return browserLang as SupportedLanguage;
    }

    // 3. Default to English
    return DEFAULT_LANGUAGE;
  }

  private isValidLanguage(lang: string): boolean {
    return SUPPORTED_LANGUAGES.includes(lang as SupportedLanguage);
  }

  setLanguage(lang: SupportedLanguage): void {
    if (this.isValidLanguage(lang)) {
      localStorage.setItem(LANGUAGE_STORAGE_KEY, lang);
      this._currentLanguage.set(lang);
    }
  }

  getLanguage(): SupportedLanguage {
    return this._currentLanguage();
  }

  getAvailableLanguages(): readonly SupportedLanguage[] {
    return SUPPORTED_LANGUAGES;
  }
}
