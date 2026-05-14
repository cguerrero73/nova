import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthConfigService } from '@core/services/auth-config.service';
import { AuthConfig, AuthMethod, UpdateAuthConfigRequest } from '@core/models/auth-config.model';

interface LocalConfig {
  requireEmailVerification: boolean;
  minPasswordLength: number;
  maxLoginAttempts: number;
  lockoutDuration: number;
}

interface LdapConfig {
  url: string;
  port: number;
  baseDn: string;
  bindDn?: string;
  bindPassword?: string;
  userSearchFilter: string;
  groupSearchFilter?: string;
}

interface OAuth2Config {
  provider: 'google' | 'github' | 'microsoft';
  clientId: string;
  clientSecret: string;
  redirectUri: string;
  scopes: string[];
}

interface AzureAdConfig {
  tenantId: string;
  clientId: string;
  clientSecret: string;
  redirectUri: string;
  authority: string;
}

interface SamlConfig {
  entryPoint: string;
  issuer: string;
  redirectUrl: string;
  cert: string;
}

@Component({
  selector: 'app-auth-config',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './auth-config.component.html',
  styleUrl: './auth-config.component.css',
})
export class AuthConfigComponent implements OnInit {
  private readonly authConfigService = inject(AuthConfigService);

  // State
  config = signal<AuthConfig | null>(null);
  isLoading = signal(true);
  isSaving = signal(false);
  showSecret = signal(false);
  errorMessage = signal<string | null>(null);
  successMessage = signal<string | null>(null);

  // Form values (for editing)
  selectedMethod = signal<AuthMethod>('local');
  isActive = signal(true);
  
  // Local config
  localConfig = signal<LocalConfig>({
    requireEmailVerification: false,
    minPasswordLength: 8,
    maxLoginAttempts: 5,
    lockoutDuration: 15,
  });

  // LDAP config
  ldapConfig = signal<LdapConfig>({
    url: '',
    port: 389,
    baseDn: '',
    bindDn: '',
    bindPassword: '',
    userSearchFilter: '(uid={{username}})',
    groupSearchFilter: '',
  });

  // OAuth2 config
  oauth2Config = signal<OAuth2Config>({
    provider: 'google',
    clientId: '',
    clientSecret: '',
    redirectUri: '',
    scopes: ['openid', 'profile', 'email'],
  });

  // Azure AD config
  azureAdConfig = signal<AzureAdConfig>({
    tenantId: '',
    clientId: '',
    clientSecret: '',
    redirectUri: '',
    authority: '',
  });

  // SAML config
  samlConfig = signal<SamlConfig>({
    entryPoint: '',
    issuer: '',
    redirectUrl: '',
    cert: '',
  });

  authMethods: { value: AuthMethod; label: string; description: string }[] = [
    { value: 'local', label: 'Local', description: 'Usuario y contraseña almacenados en la base de datos' },
    { value: 'ldap', label: 'LDAP/Active Directory', description: 'Autenticación contra servidor LDAP' },
    { value: 'oauth2', label: 'OAuth 2.0', description: 'Google, GitHub, etc.' },
    { value: 'azure-ad', label: 'Azure AD', description: 'Microsoft Azure Active Directory' },
    { value: 'saml', label: 'SAML', description: 'Security Assertion Markup Language' },
  ];

  ngOnInit(): void {
    this.loadConfig();
  }

  loadConfig(): void {
    this.isLoading.set(true);
    this.authConfigService.getConfig().subscribe({
      next: (response) => {
        const cfg = response.data;
        if (!cfg) {
          this.errorMessage.set('Configuración no encontrada');
          this.isLoading.set(false);
          return;
        }
        
        this.config.set(cfg);
        this.selectedMethod.set(cfg.method);
        this.isActive.set(cfg.isActive);
        
        if (cfg.local) this.localConfig.set({ ...cfg.local });
        if (cfg.ldap) this.ldapConfig.set({ ...cfg.ldap });
        if (cfg.oauth2) this.oauth2Config.set({ ...cfg.oauth2 });
        if (cfg.azureAd) this.azureAdConfig.set({ ...cfg.azureAd });
        if (cfg.saml) this.samlConfig.set({ ...cfg.saml });
        
        this.isLoading.set(false);
      },
      error: (err) => {
        this.errorMessage.set('Error al cargar configuración');
        this.isLoading.set(false);
      }
    });
  }

  saveConfig(): void {
    this.isSaving.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    const configData: UpdateAuthConfigRequest = {
      method: this.selectedMethod(),
      isActive: this.isActive(),
    };

    // Add method-specific config
    switch (this.selectedMethod()) {
      case 'local':
        configData.local = this.localConfig();
        break;
      case 'ldap':
        configData.ldap = this.ldapConfig();
        break;
      case 'oauth2':
        configData.oauth2 = this.oauth2Config();
        break;
      case 'azure-ad':
        configData.azureAd = this.azureAdConfig();
        break;
      case 'saml':
        configData.saml = this.samlConfig();
        break;
    }

    this.authConfigService.updateConfig(configData).subscribe({
      next: (response) => {
        if (response.data) {
          this.config.set(response.data);
        }
        this.successMessage.set('Configuración guardada correctamente');
        this.isSaving.set(false);
        
        // Clear success message after 3 seconds
        setTimeout(() => this.successMessage.set(null), 3000);
      },
      error: (err) => {
        this.errorMessage.set('Error al guardar configuración');
        this.isSaving.set(false);
      }
    });
  }

  onMethodChange(): void {
    // Reset dependent configs if needed
  }

  // Helper methods for local config
  updateLocalMinPasswordLength(value: number): void {
    this.localConfig.update(c => ({ ...c, minPasswordLength: value }));
  }

  updateLocalMaxLoginAttempts(value: number): void {
    this.localConfig.update(c => ({ ...c, maxLoginAttempts: value }));
  }

  updateLocalLockoutDuration(value: number): void {
    this.localConfig.update(c => ({ ...c, lockoutDuration: value }));
  }

  toggleLocalRequireEmailVerification(): void {
    this.localConfig.update(c => ({ ...c, requireEmailVerification: !c.requireEmailVerification }));
  }

  // Helper methods for LDAP config
  updateLdapUrl(value: string): void {
    this.ldapConfig.update(c => ({ ...c, url: value }));
  }

  updateLdapPort(value: number): void {
    this.ldapConfig.update(c => ({ ...c, port: value }));
  }

  updateLdapBaseDn(value: string): void {
    this.ldapConfig.update(c => ({ ...c, baseDn: value }));
  }

  updateLdapBindDn(value: string): void {
    this.ldapConfig.update(c => ({ ...c, bindDn: value }));
  }

  updateLdapBindPassword(value: string): void {
    this.ldapConfig.update(c => ({ ...c, bindPassword: value }));
  }

  updateLdapUserSearchFilter(value: string): void {
    this.ldapConfig.update(c => ({ ...c, userSearchFilter: value }));
  }

  updateLdapGroupSearchFilter(value: string): void {
    this.ldapConfig.update(c => ({ ...c, groupSearchFilter: value }));
  }

  // Helper methods for OAuth2 config
  updateOAuth2Provider(value: 'google' | 'github' | 'microsoft'): void {
    this.oauth2Config.update(c => ({ ...c, provider: value }));
  }

  updateOAuth2ClientId(value: string): void {
    this.oauth2Config.update(c => ({ ...c, clientId: value }));
  }

  updateOAuth2ClientSecret(value: string): void {
    this.oauth2Config.update(c => ({ ...c, clientSecret: value }));
  }

  updateOAuth2RedirectUri(value: string): void {
    this.oauth2Config.update(c => ({ ...c, redirectUri: value }));
  }

  updateOAuth2Scopes(value: string): void {
    this.oauth2Config.update(c => ({ ...c, scopes: value.split(',').map(s => s.trim()) }));
  }

  // Helper methods for Azure AD config
  updateAzureTenantId(value: string): void {
    this.azureAdConfig.update(c => ({ ...c, tenantId: value }));
  }

  updateAzureClientId(value: string): void {
    this.azureAdConfig.update(c => ({ ...c, clientId: value }));
  }

  updateAzureClientSecret(value: string): void {
    this.azureAdConfig.update(c => ({ ...c, clientSecret: value }));
  }

  updateAzureAuthority(value: string): void {
    this.azureAdConfig.update(c => ({ ...c, authority: value }));
  }

  updateAzureRedirectUri(value: string): void {
    this.azureAdConfig.update(c => ({ ...c, redirectUri: value }));
  }

  // Helper methods for SAML config
  updateSamlEntryPoint(value: string): void {
    this.samlConfig.update(c => ({ ...c, entryPoint: value }));
  }

  updateSamlIssuer(value: string): void {
    this.samlConfig.update(c => ({ ...c, issuer: value }));
  }

  updateSamlRedirectUrl(value: string): void {
    this.samlConfig.update(c => ({ ...c, redirectUrl: value }));
  }

  updateSamlCert(value: string): void {
    this.samlConfig.update(c => ({ ...c, cert: value }));
  }
}