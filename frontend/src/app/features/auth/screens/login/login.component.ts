import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '@core/services/auth.service';
import { TranslationService } from '@core/services/translation.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './login.component.html',
  styleUrl: './login.component.css',
})
export class LoginComponent implements OnInit {
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);
  private readonly translate = inject(TranslationService);

  // Traducciones
  t: Record<string, string> = {};

  // Form state
  email = signal('');
  password = signal('');
  rememberMe = signal(false);

  // UI state
  isLoading = signal(false);
  errorMessage = signal<string | null>(null);
  showPassword = signal(false);

  // Modo: 'login' | 'register'
  mode = signal<'login' | 'register'>('login');
  name = signal('');

  ngOnInit() {
    // Cargar traducciones de login
    this.translate.load('login').subscribe(translations => {
      this.t = translations;
    });
  }

  async onSubmit(): Promise<void> {
    this.errorMessage.set(null);
    this.isLoading.set(true);

    try {
      if (this.mode() === 'login') {
        await this.authService.login({
          email: this.email(),
          password: this.password()
        }).toPromise();
        
        this.router.navigate(['/users']);
      } else {
        await this.authService.register({
          email: this.email(),
          password: this.password(),
          name: this.name()
        }).toPromise();
        
        this.router.navigate(['/users']);
      }
    } catch (error: unknown) {
      // Manejar error de HTTP
      const httpError = error as { status?: number; error?: { success?: boolean; error?: string } };
      
      if (httpError.status === 401 || httpError.status === 400) {
        // Error de credenciales del mock
        this.errorMessage.set(this.t['error.credentials'] || httpError.error?.error || 'Credenciales inválidas');
      } else if (httpError.error?.success === false) {
        // Error del servidor con formato ApiResponse
        this.errorMessage.set(httpError.error.error || 'Error de autenticación');
      } else {
        this.errorMessage.set(this.t['error.connection'] || 'Error de conexión');
      }
    } finally {
      this.isLoading.set(false);
    }
  }

  toggleMode(): void {
    this.mode.set(this.mode() === 'login' ? 'register' : 'login');
    this.errorMessage.set(null);
  }

  togglePasswordVisibility(): void {
    this.showPassword.set(!this.showPassword());
  }
}