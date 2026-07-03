import { Routes } from '@angular/router';

export const FORM_BUILDER_ROUTES: Routes = [
  {
    path: 'forms/:formKey',
    loadComponent: () =>
      import('./runtime/form-runtime-container.component').then(
        (m) => m.FormRuntimeContainerComponent,
      ),
  },
];
