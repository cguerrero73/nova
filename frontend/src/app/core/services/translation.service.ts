import { Injectable, inject, signal, computed } from '@angular/core';
import { Observable, of, tap, catchError, map, shareReplay } from 'rxjs';
import { ApiService } from './api.service';
import { LanguageDetector, SupportedLanguage, DEFAULT_LANGUAGE } from './language-detector.service';

export interface ScreenTranslations {
  screenId: string;
  translations: Record<string, string>;
}

@Injectable({ providedIn: 'root' })
export class TranslationService {
  private readonly api = inject(ApiService);
  private readonly languageDetector = inject(LanguageDetector);

  // Cache por pantalla: Map<screenId, translations>
  private translationsCache = new Map<string, Record<string, string>>();
  
  // Loading states
  private loadingStates = new Map<string, boolean>();
  
  // Signals reactivos
  private _loadedScreens = signal<Set<string>>(new Set());
  readonly loadedScreens = this._loadedScreens.asReadonly();

  // Idioma actual (computed del detector)
  readonly currentLanguage = computed(() => this.languageDetector.currentLanguage());

  /**
   * Cargar traducciones de una pantalla específica
   * Si ya están en cache, retorna inmediatamente
   */
  load(screenId: string): Observable<Record<string, string>> {
    // Si ya están cacheadas, retornar del cache
    if (this.translationsCache.has(screenId)) {
      return of(this.translationsCache.get(screenId)!);
    }

    // Si ya se están cargando, esperar
    if (this.loadingStates.get(screenId)) {
      // Polling hasta que terminen de cargar
      return this.waitForLoad(screenId);
    }

    // Cargar del backend
    this.loadingStates.set(screenId, true);
    
    const lang = this.languageDetector.getLanguage();
    console.log('[TranslationService] Loading screen:', screenId, 'lang:', lang);
    
    return this.api.get<ScreenTranslations>(`/screens/${screenId}`, { lang }).pipe(
      tap(response => {
        console.log('[TranslationService] Response:', response);
        if (response.success && response.data) {
          this.translationsCache.set(screenId, response.data.translations);
          this._loadedScreens.update(set => new Set([...set, screenId]));
        }
        this.loadingStates.set(screenId, false);
      }),
      map(response => {
        if (response.success && response.data) {
          return response.data.translations;
        }
        // Si falla, retornar traducciones vacías
        return {};
      }),
      catchError(() => {
        this.loadingStates.set(screenId, false);
        return of({});
      }),
      shareReplay(1)
    );
  }

  /**
   * Obtener traducción por key
   * Ejemplo: t('users.title') -> "Usuarios"
   */
  t(screenId: string, key: string): string {
    const translations = this.translationsCache.get(screenId);
    if (!translations) {
      return key; // Return key as fallback
    }
    return translations[key] || key;
  }

  /**
   * Obtener todas las traducciones de una pantalla
   */
  getTranslations(screenId: string): Record<string, string> {
    return this.translationsCache.get(screenId) || {};
  }

  /**
   * Verificar si las traducciones de una pantalla ya están cargadas
   */
  isLoaded(screenId: string): boolean {
    return this.translationsCache.has(screenId);
  }

  /**
   * Recargar traducciones de una pantalla (para cambio de idioma)
   */
  reload(screenId: string): Observable<Record<string, string>> {
    this.translationsCache.delete(screenId);
    this._loadedScreens.update(set => {
      const newSet = new Set(set);
      newSet.delete(screenId);
      return newSet;
    });
    return this.load(screenId);
  }

  /**
   * Cambiar idioma de la aplicación
   */
  setLanguage(lang: SupportedLanguage): void {
    this.languageDetector.setLanguage(lang);
    // Limpiar cache para recargar con nuevo idioma
    this.translationsCache.clear();
    this._loadedScreens.set(new Set());
  }

  /**
   * Obtener el idioma actual
   */
  getLanguage(): SupportedLanguage {
    return this.languageDetector.getLanguage();
  }

  private waitForLoad(screenId: string): Observable<Record<string, string>> {
    return new Observable(observer => {
      const check = () => {
        if (!this.loadingStates.get(screenId)) {
          const translations = this.translationsCache.get(screenId);
          observer.next(translations || {});
          observer.complete();
        } else {
          setTimeout(check, 50);
        }
      };
      check();
    });
  }
}
