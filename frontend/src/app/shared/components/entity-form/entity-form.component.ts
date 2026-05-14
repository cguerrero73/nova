import { Component, input, output, signal, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

export interface FormField {
  key: string;
  label: string;
  type: 'text' | 'email' | 'password' | 'number' | 'select' | 'textarea' | 'date' | 'checkbox';
  required?: boolean;
  placeholder?: string;
  options?: { value: string; label: string }[];
  readonly?: boolean;
  rows?: number;
}

@Component({
  selector: 'app-entity-form',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <form class="space-y-6">
      @for (field of fields(); track field.key) {
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Label -->
          <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ field.label }}
            @if (field.required) {
              <span class="text-red-500">*</span>
            }
          </label>
          
          <!-- Input -->
          <div>
            @switch (field.type) {
              @case ('textarea') {
                <textarea
                  [ngModel]="getValue(field.key)"
                  (ngModelChange)="onFieldChange(field.key, $event)"
                  [name]="field.key"
                  [placeholder]="field.placeholder || ''"
                  [attr.required]="field.required ? true : null"
                  [attr.readonly]="field.readonly ? true : null"
                  [rows]="field.rows || 3"
                  class="w-full px-3 py-2 border rounded-md bg-background text-foreground"
                  [style.border-color]="'var(--color-border)'"
                  [style.background-color]="'var(--color-background)'"
                  [style.color]="'var(--color-text)'"
                ></textarea>
              }
              @case ('select') {
                <select
                  [ngModel]="getValue(field.key)"
                  (ngModelChange)="onFieldChange(field.key, $event)"
                  [name]="field.key"
                  [attr.required]="field.required ? true : null"
                  [attr.readonly]="field.readonly ? true : null"
                  class="w-full px-3 py-2 border rounded-md bg-background text-foreground"
                  [style.border-color]="'var(--color-border)'"
                  [style.background-color]="'var(--color-background)'"
                  [style.color]="'var(--color-text)'"
                >
                  @for (opt of field.options; track opt.value) {
                    <option [value]="opt.value">{{ opt.label }}</option>
                  }
                </select>
              }
              @case ('checkbox') {
                <input
                  type="checkbox"
                  [ngModel]="getValue(field.key)"
                  (ngModelChange)="onFieldChange(field.key, $event)"
                  [name]="field.key"
                  [attr.readonly]="field.readonly ? true : null"
                  class="w-4 h-4 rounded border-gray-300 text-primary-600"
                />
              }
              @case ('date') {
                <input
                  type="date"
                  [ngModel]="getValue(field.key)"
                  (ngModelChange)="onFieldChange(field.key, $event)"
                  [name]="field.key"
                  [attr.required]="field.required ? true : null"
                  [attr.readonly]="field.readonly ? true : null"
                  class="w-full px-3 py-2 border rounded-md bg-background text-foreground"
                  [style.border-color]="'var(--color-border)'"
                  [style.background-color]="'var(--color-background)'"
                  [style.color]="'var(--color-text)'"
                />
              }
              @case ('password') {
                <input
                  type="password"
                  [ngModel]="getValue(field.key)"
                  (ngModelChange)="onFieldChange(field.key, $event)"
                  [name]="field.key"
                  [placeholder]="field.placeholder || ''"
                  [attr.required]="field.required ? true : null"
                  [attr.readonly]="field.readonly ? true : null"
                  class="w-full px-3 py-2 border rounded-md bg-background text-foreground"
                  [style.border-color]="'var(--color-border)'"
                  [style.background-color]="'var(--color-background)'"
                  [style.color]="'var(--color-text)'"
                />
              }
              @case ('email') {
                <input
                  type="email"
                  [ngModel]="getValue(field.key)"
                  (ngModelChange)="onFieldChange(field.key, $event)"
                  [name]="field.key"
                  [placeholder]="field.placeholder || ''"
                  [attr.required]="field.required ? true : null"
                  [attr.readonly]="field.readonly ? true : null"
                  class="w-full px-3 py-2 border rounded-md bg-background text-foreground"
                  [style.border-color]="'var(--color-border)'"
                  [style.background-color]="'var(--color-background)'"
                  [style.color]="'var(--color-text)'"
                />
              }
              @case ('number') {
                <input
                  type="number"
                  [ngModel]="getValue(field.key)"
                  (ngModelChange)="onFieldChange(field.key, $event)"
                  [name]="field.key"
                  [placeholder]="field.placeholder || ''"
                  [attr.required]="field.required ? true : null"
                  [attr.readonly]="field.readonly ? true : null"
                  class="w-full px-3 py-2 border rounded-md bg-background text-foreground"
                  [style.border-color]="'var(--color-border)'"
                  [style.background-color]="'var(--color-background)'"
                  [style.color]="'var(--color-text)'"
                />
              }
              @default {
                <input
                  type="text"
                  [ngModel]="getValue(field.key)"
                  (ngModelChange)="onFieldChange(field.key, $event)"
                  [name]="field.key"
                  [placeholder]="field.placeholder || ''"
                  [attr.required]="field.required ? true : null"
                  [attr.readonly]="field.readonly ? true : null"
                  class="w-full px-3 py-2 border rounded-md bg-background text-foreground"
                  [style.border-color]="'var(--color-border)'"
                  [style.background-color]="'var(--color-background)'"
                  [style.color]="'var(--color-text)'"
                />
              }
            }
          </div>
        </div>
      }
    </form>
  `,
  styles: [`
    :host { display: block; }
  `]
})
export class EntityFormComponent {
  // Campos del formulario
  fields = input.required<FormField[]>();
  
  // Datos iniciales
  initialData = input<Record<string, unknown>>({});
  
  // Evento cuando cambia un campo
  fieldChange = output<{ key: string; value: unknown }>();
  
  // Datos actuales (copia editable)
  private data = signal<Record<string, unknown>>({});
  
  constructor() {
    // Sincronizar con initialData
    effect(() => {
      this.data.set({ ...this.initialData() });
    });
  }
  
  getValue(key: string): unknown {
    return this.data()[key];
  }
  
  onFieldChange(key: string, value: unknown) {
    this.data.update(current => ({
      ...current,
      [key]: value
    }));
    this.fieldChange.emit({ key, value });
  }
}