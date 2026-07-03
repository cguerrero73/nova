import { Injectable, signal, computed, inject } from '@angular/core';
import { LayoutDefinition, Section } from '../models/layout-definition.model';
import { FormRuntimeService } from '../services/runtime.service';

/**
 * Signal-based local state store for the form runtime.
 * Manages layout, loading, and error state.
 */
@Injectable({ providedIn: 'root' })
export class FormRuntimeStore {
  private readonly runtimeService = inject(FormRuntimeService);

  // --- State ---
  private _layout = signal<LayoutDefinition | null>(null);
  private _loading = signal(false);
  private _error = signal<string | null>(null);

  // --- Public read-only signals ---
  readonly layout = this._layout.asReadonly();
  readonly loading = this._loading.asReadonly();
  readonly error = this._error.asReadonly();

  readonly sortedSections = computed<Section[]>(() => {
    const l = this._layout();
    if (!l) return [];
    return [...l.sections].sort((a, b) => a.order - b.order);
  });

  readonly formKey = computed(() => this._layout()?.formKey ?? null);
  readonly layoutName = computed(() => this._layout()?.layoutName ?? null);

  // --- Actions ---

  loadForm(formKey: string): void {
    this._loading.set(true);
    this._error.set(null);

    this.runtimeService.resolveForm(formKey).subscribe({
      next: (layout) => {
        this._layout.set(layout);
        this._loading.set(false);
      },
      error: (err) => {
        this._error.set(err?.message ?? 'Failed to load form');
        this._loading.set(false);
      },
    });
  }

  reset(): void {
    this._layout.set(null);
    this._loading.set(false);
    this._error.set(null);
  }
}
