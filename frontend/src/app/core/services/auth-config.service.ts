import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '@core/services/api.service';
import { AuthConfig, UpdateAuthConfigRequest } from '@core/models/auth-config.model';
import { ApiResponse } from '@core/models/api.model';

@Injectable({ providedIn: 'root' })
export class AuthConfigService {
  private readonly api = inject(ApiService);

  getConfig(): Observable<ApiResponse<AuthConfig>> {
    return this.api.get<AuthConfig>('/auth/config');
  }

  updateConfig(config: UpdateAuthConfigRequest): Observable<ApiResponse<AuthConfig>> {
    return this.api.put<AuthConfig>('/auth/config', 'current', config);
  }
}