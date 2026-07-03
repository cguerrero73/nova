import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormDesignerStore } from '../state/designer.store';
import { FormRuntimeComponent } from '../runtime/form-runtime.component';
import { LayoutDefinition } from '../models/layout-definition.model';

/**
 * Opens FormRuntimeComponent in preview mode with the current draft in-memory.
 * Rendered as a modal overlay.
 */
@Component({
  selector: 'app-preview-dialog',
  standalone: true,
  imports: [CommonModule, FormRuntimeComponent],
  template: `
    @if (visible()) {
      <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" (click)="close()">
        <div
          class="bg-white rounded-lg shadow-xl w-full max-w-3xl max-h-[90vh] overflow-auto"
          (click)="$event.stopPropagation()"
        >
          <div class="flex items-center justify-between p-4 border-b">
            <h2 class="text-lg font-semibold">Preview</h2>
            <button (click)="close()" class="text-gray-400 hover:text-gray-600">✕</button>
          </div>
          <div class="p-4">
            @if (previewLayout(); as layout) {
              <app-form-runtime [layout]="layout" />
            } @else {
              <p class="text-sm text-gray-400 text-center py-8">No layout to preview.</p>
            }
          </div>
        </div>
      </div>
    }
  `,
})
export class PreviewDialogComponent {
  readonly store = inject(FormDesignerStore);

  visible = signal(false);
  previewLayout = signal<LayoutDefinition | null>(null);

  open(): void {
    const def = this.store.definition();
    if (def) {
      this.previewLayout.set({ ...def });
      this.visible.set(true);
    }
  }

  close(): void {
    this.visible.set(false);
    this.previewLayout.set(null);
  }
}
