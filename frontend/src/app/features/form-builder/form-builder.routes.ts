import { Routes } from '@angular/router';

export const FORM_BUILDER_ROUTES: Routes = [
  {
    path: 'forms/:formKey',
    loadComponent: () =>
      import('./runtime/form-runtime-container.component').then(
        (m) => m.FormRuntimeContainerComponent,
      ),
  },
  {
    path: 'designer/:formKey',
    loadComponent: () =>
      import('./designer/form-designer.component').then(
        (m) => m.FormDesignerComponent,
      ),
  },
  {
    path: 'designer/:formKey/:layoutName',
    loadComponent: () =>
      import('./designer/form-designer.component').then(
        (m) => m.FormDesignerComponent,
      ),
  },
];
