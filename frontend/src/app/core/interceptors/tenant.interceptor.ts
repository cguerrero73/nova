import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { TenantService } from '../services/tenant.service';

/**
 * HTTP interceptor that adds the X-Tenant-Code header to every API request.
 *
 * The tenant code is extracted from the URL by TenantService on app startup
 * (from ?tenant= query param or subdomain).
 */
export const tenantInterceptor: HttpInterceptorFn = (req, next) => {
  const tenantService = inject(TenantService);
  const tenant = tenantService.getTenant();

  if (!tenant) {
    return next(req);
  }

  // Skip if the request already has the header (e.g. login with body-tenant)
  if (req.headers.has('X-Tenant-Code')) {
    return next(req);
  }

  const modifiedReq = req.clone({
    setHeaders: {
      'X-Tenant-Code': tenant,
    },
  });

  return next(modifiedReq);
};
