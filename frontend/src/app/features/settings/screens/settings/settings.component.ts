import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="p-6">
      <h1 class="text-2xl font-bold mb-6" [style.color]="'var(--color-text)'">Configuración</h1>
      <div class="p-4 rounded-lg" [style.background-color]="'var(--color-surface)'" [style.border]="'1px solid var(--color-border)'">
        <p [style.color]="'var(--color-text-secondary)'">Pantalla de Configuración - En desarrollo</p>
      </div>
    </div>
  `,
})
export class SettingsComponent {}