import { Injectable, inject, signal, computed } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, tap, catchError, of } from 'rxjs';
import { ApiService } from './api.service';
import { LanguageDetector, SupportedLanguage, LANGUAGE_STORAGE_KEY } from './language-detector.service';

export interface AuthUser {
  id: string;
  email: string;
  name: string;
  roles: string[];
  language?: string;
  picture?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface AuthResponse {
  user: AuthUser;
  accessToken: string;
  refreshToken?: string;
  expiresIn: number;
}

const TOKEN_KEY = 'nova_access_token';
const REFRESH_TOKEN_KEY = 'nova_refresh_token';
const USER_KEY = 'nova_user';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly api = inject(ApiService);
  private readonly router = inject(Router);
  private readonly languageDetector = inject(LanguageDetector);

  // Signals para estado reactivo
  private _user = signal<AuthUser | null>(this.loadUserFromStorage());
  private _isAuthenticated = signal<boolean>(this.hasValidToken());
  private _isLoading = signal<boolean>(false);

  // Computed values
  readonly user = this._user.asReadonly();
  readonly isAuthenticated = this._isAuthenticated.asReadonly();
  readonly isLoading = this._isLoading.asReadonly();

  readonly isAdmin = computed(() => 
    this._user()?.roles?.includes('admin') ?? false
  );

  constructor() {
    // Verificar token al inicializar
    this.checkTokenValidity();
  }

  login(credentials: LoginRequest): Observable<{ success: boolean; data: AuthResponse }> {
    this._isLoading.set(true);
    return this.api.postRaw<{ success: boolean; data: AuthResponse }>('/auth/login', credentials).pipe(
      tap((response) => {
        if (response.success && response.data) {
          this.setSession(response.data);
        }
        this._isLoading.set(false);
      }),
      catchError((error) => {
        this._isLoading.set(false);
        throw error;
      })
    );
  }

  register(data: RegisterRequest): Observable<{ success: boolean; data: AuthResponse }> {
    this._isLoading.set(true);
    return this.api.postRaw<{ success: boolean; data: AuthResponse }>('/auth/register', data).pipe(
      tap((response) => {
        if (response.success && response.data) {
          this.setSession(response.data);
        }
        this._isLoading.set(false);
      }),
      catchError((error) => {
        this._isLoading.set(false);
        throw error;
      })
    );
  }

  logout(): void {
    // Llamar al endpoint de logout (opcional - invalidate token en server)
    this.api.postRaw('/auth/logout', {}).subscribe({
      next: () => {},
      error: () => {} // Ignorar errores en logout
    });

    this.clearSession();
    this.router.navigate(['/login']);
  }

  refreshToken(): Observable<{ accessToken: string; expiresIn: number } | null> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) {
      return of(null);
    }

    return this.api.postRaw<{ accessToken: string; expiresIn: number }>('/auth/refresh', {
      refreshToken
    }).pipe(
      tap((response) => {
        this.setAccessToken(response.accessToken, response.expiresIn);
      }),
      catchError(() => {
        this.clearSession();
        return of(null);
      })
    );
  }

  getAccessToken(): string | null {
    return localStorage.getItem(TOKEN_KEY);
  }

  getRefreshToken(): string | null {
    return localStorage.getItem(REFRESH_TOKEN_KEY);
  }

  /**
   * Actualizar el idioma del usuario en la sesión
   * También actualiza el LanguageDetector y guarda en localStorage
   */
  setUserLanguage(language: SupportedLanguage): void {
    // Actualizar en el LanguageDetector
    this.languageDetector.setLanguage(language);
    
    // Actualizar en el usuario en memoria
    this._user.update(user => {
      if (!user) return user;
      return { ...user, language };
    });
  }

  private setSession(authResult: AuthResponse): void {
    const { accessToken, refreshToken, user, expiresIn } = authResult;
    
    const expireDate = new Date(Date.now() + expiresIn * 1000);
    
    localStorage.setItem(TOKEN_KEY, accessToken);
    if (refreshToken) {
      localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
    }
    localStorage.setItem(USER_KEY, JSON.stringify(user));
    // Store expiry for checking
    localStorage.setItem('nova_token_expiry', expireDate.toISOString());

    // Set language from user preference (post-login)
    if (user.language) {
      this.languageDetector.setLanguage(user.language as SupportedLanguage);
    } else {
      // If no language preference, use the one already detected
      localStorage.setItem(LANGUAGE_STORAGE_KEY, this.languageDetector.getLanguage());
    }

    this._user.set(user);
    this._isAuthenticated.set(true);
  }

  private clearSession(): void {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    localStorage.removeItem('nova_token_expiry');
    
    this._user.set(null);
    this._isAuthenticated.set(false);
  }

  private loadUserFromStorage(): AuthUser | null {
    const userJson = localStorage.getItem(USER_KEY);
    if (!userJson) return null;
    
    try {
      return JSON.parse(userJson) as AuthUser;
    } catch {
      return null;
    }
  }

  private hasValidToken(): boolean {
    const token = localStorage.getItem(TOKEN_KEY);
    const expiry = localStorage.getItem('nova_token_expiry');
    
    if (!token || !expiry) return false;
    
    // Check if token is expired
    return new Date(expiry) > new Date();
  }

  private checkTokenValidity(): void {
    const expiry = localStorage.getItem('nova_token_expiry');
    
    if (expiry && new Date(expiry) <= new Date()) {
      // Token expired - try to refresh
      this.refreshToken().subscribe({
        next: (result) => {
          if (!result) {
            this.clearSession();
            this.router.navigate(['/login']);
          }
        },
        error: () => {
          this.clearSession();
          this.router.navigate(['/login']);
        }
      });
    }
  }

  private setAccessToken(token: string, expiresIn: number): void {
    localStorage.setItem(TOKEN_KEY, token);
    const expireDate = new Date(Date.now() + expiresIn * 1000);
    localStorage.setItem('nova_token_expiry', expireDate.toISOString());
  }
}