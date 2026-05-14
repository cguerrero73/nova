import { Routes } from '@angular/router';
import { authGuard, loginGuard, adminGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  {
    path: '',
    redirectTo: 'dashboard',
    pathMatch: 'full',
  },
  {
    path: 'login',
    loadComponent: () =>
      import('./features/auth/screens/login/login.component').then(
        (m) => m.LoginComponent
      ),
    canActivate: [loginGuard],
  },
  {
    path: '',
    loadComponent: () =>
      import('./shared/components/layout/layout.component').then(
        (m) => m.LayoutComponent
      ),
    canActivate: [authGuard],
    children: [
      {
        path: 'dashboard',
        loadComponent: () =>
          import('./features/dashboard/screens/dashboard/dashboard.component').then(
            (m) => m.DashboardComponent
          ),
      },
      {
        path: 'users',
        loadComponent: () =>
          import('./features/users/screens/user-list/user-list.component').then(
            (m) => m.UserListComponent
          ),
      },
      {
        path: 'users/:id',
        loadComponent: () =>
          import('./features/users/screens/user-detail/user-detail.component').then(
            (m) => m.UserDetailComponent
          ),
      },
      {
        path: 'roles',
        loadComponent: () =>
          import('./features/roles/screens/role-list/role-list.component').then(
            (m) => m.RoleListComponent
          ),
      },
      {
        path: 'settings',
        loadComponent: () =>
          import('./features/settings/screens/settings/settings.component').then(
            (m) => m.SettingsComponent
          ),
      },
      {
        path: 'admin/auth-config',
        loadComponent: () =>
          import('./features/admin/screens/auth-config/auth-config.component').then(
            (m) => m.AuthConfigComponent
          ),
        canActivate: [adminGuard],
      },
      {
        path: 'admin/audit',
        loadComponent: () =>
          import('./features/admin/screens/audit-log/audit-log.component').then(
            (m) => m.AuditLogComponent
          ),
        canActivate: [adminGuard],
      },
    ],
  },
];