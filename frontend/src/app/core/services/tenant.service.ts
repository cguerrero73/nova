import { Injectable, signal } from '@angular/core';

/**
 * TenantService extracts and stores the tenant code from the application URL.
 *
 * The tenant is obtained from one of these sources (in order):
 * 1. Query parameter: ?tenant=acme
 * 2. Subdomain (e.g. acme.localhost)
 *
 * Once set, it's provided to the HTTP interceptor which adds it as
 * X-Tenant-Code header on every API request.
 */
@Injectable({ providedIn: 'root' })
export class TenantService {
  private readonly _tenant = signal<string>('');

  /** Reactive signal — components can read it if needed */
  readonly tenant = this._tenant.asReadonly();

  constructor() {
    this.initialize();
  }

  private initialize(): void {
    const tenant = this.extractFromQueryParam() ?? this.extractFromSubdomain() ?? '';

    if (tenant) {
      this._tenant.set(tenant);
    }
  }

  private extractFromQueryParam(): string | null {
    const params = new URLSearchParams(window.location.search);
    return params.get('tenant');
  }

  private extractFromSubdomain(): string | null {
    const hostname = window.location.hostname;
    // localhost or IP → no subdomain
    if (
      hostname === 'localhost' ||
      hostname === '127.0.0.1' ||
      /^\d+\.\d+\.\d+\.\d+$/.test(hostname)
    ) {
      return null;
    }
    const parts = hostname.split('.');
    // subdomain.example.com → subdomain
    if (parts.length >= 3) {
      return parts[0];
    }
    return null;
  }

  /** Returns the current tenant code. */
  getTenant(): string {
    return this._tenant();
  }

  /** Returns true if a tenant is configured. */
  hasTenant(): boolean {
    return this._tenant() !== '';
  }
}
