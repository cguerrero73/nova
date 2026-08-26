import { Injectable, inject } from '@angular/core';
import { Observable, of } from 'rxjs';
import { map, shareReplay } from 'rxjs/operators';
import { ApiService } from './api.service';
import { FieldOption } from '@features/form-builder/models/layout-definition.model';

export interface SysCode {
  sys_id: string;
  sys_type: string;
  sys_code: string;
  sys_ucode: string;
  sys_desc: string;
  sys_system: string;
}

@Injectable({ providedIn: 'root' })
export class SyscodeService {
  private readonly api = inject(ApiService);
  private readonly cache = new Map<string, Observable<FieldOption[]>>();

  getByType(type: string): Observable<FieldOption[]> {
    if (!type) return of([]);

    const cached = this.cache.get(type);
    if (cached) return cached;

    const request$ = this.api.get<SysCode[]>(`/syscodes/type/${type}`).pipe(
      map((response) => {
        const codes = response?.data ?? [];
        return codes.map((code) => ({
          label: code.sys_desc || code.sys_ucode || code.sys_code,
          value: code.sys_code,
        }));
      }),
      shareReplay({ bufferSize: 1, refCount: true })
    );

    this.cache.set(type, request$);
    return request$;
  }

  invalidateCache(): void {
    this.cache.clear();
  }
}
