import { Component, inject, signal, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { FormDesignerStore } from '../state/designer.store';
import { FormDesignerService } from '../services/designer.service';
import { FormLayout } from '../models/designer.model';

/**
 * Lists layouts for the current form, lets user pick one to edit or create new.
 * The `default` layout shows a badge "Default — applies to all roles without specific assignment".
 * Creating a layout asks for a unique slug-friendly name.
 */
@Component({
  selector: 'app-layout-picker',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="layout-picker">
      <h3 class="text-sm font-semibold text-gray-700 mb-2">Layouts</h3>

      <div class="space-y-1">
        @for (layout of store.layouts(); track layout.fl_id) {
          <button
            class="layout-item w-full text-left px-3 py-2 rounded text-sm transition-colors"
            [class.bg-blue-50]="store.currentLayout()?.fl_id === layout.fl_id"
            [class.border-blue-300]="store.currentLayout()?.fl_id === layout.fl_id"
            [class.border]="store.currentLayout()?.fl_id === layout.fl_id"
            [class.hover:bg-gray-50]="store.currentLayout()?.fl_id !== layout.fl_id"
            (click)="selectLayout(layout)"
          >
            <div class="flex items-center gap-2">
              <span class="font-medium">{{ layout.fl_display_name || layout.fl_name }}</span>
              @if (layout.fl_name === 'default') {
                <span class="badge-default text-xs bg-green-100 text-green-800 px-1.5 py-0.5 rounded">
                  Default
                </span>
              }
            </div>
            @if (layout.fl_name === 'default') {
              <p class="text-xs text-gray-500 mt-0.5">
                Applies to all roles without specific assignment
              </p>
            }
          </button>
        }
      </div>

      <!-- Create new layout -->
      <div class="mt-3 pt-3 border-t">
        @if (!showCreateForm()) {
          <button
            class="w-full text-sm text-blue-600 hover:text-blue-800 py-1"
            (click)="showCreateForm.set(true)"
          >
            + New layout
          </button>
        } @else {
          <div class="space-y-2">
            <input
              type="text"
              [(ngModel)]="newLayoutName"
              placeholder="slug-friendly-name"
              class="w-full px-2 py-1 text-sm border rounded"
              (keydown.enter)="createLayout()"
            />
            <input
              type="text"
              [(ngModel)]="newLayoutDisplayName"
              placeholder="Display name"
              class="w-full px-2 py-1 text-sm border rounded"
            />
            <div class="flex gap-1">
              <button
                class="flex-1 text-xs bg-blue-600 text-white px-2 py-1 rounded hover:bg-blue-700"
                (click)="createLayout()"
                [disabled]="!newLayoutName.trim()"
              >
                Create
              </button>
              <button
                class="text-xs px-2 py-1 text-gray-600 hover:text-gray-800"
                (click)="cancelCreate()"
              >
                Cancel
              </button>
            </div>
            @if (createError()) {
              <p class="text-xs text-red-600">{{ createError() }}</p>
            }
          </div>
        }
      </div>
    </div>
  `,
  styles: [`
    .layout-picker {
      padding: 0.5rem;
    }
  `],
})
export class LayoutPickerComponent {
  readonly store = inject(FormDesignerStore);
  private readonly designerService = inject(FormDesignerService);

  showCreateForm = signal(false);
  newLayoutName = '';
  newLayoutDisplayName = '';
  createError = signal<string | null>(null);

  selectLayout(layout: FormLayout): void {
    this.store.selectLayout(layout);
  }

  createLayout(): void {
    const name = this.newLayoutName.trim();
    if (!name) return;

    if (!/^[a-z0-9-]+$/.test(name)) {
      this.createError.set('Name must be slug-friendly (lowercase, numbers, hyphens)');
      return;
    }

    this.createError.set(null);
    this.designerService
      .createLayout(this.store.formKey(), {
        name,
        displayName: this.newLayoutDisplayName.trim() || name,
      })
      .subscribe({
        next: (layout) => {
          this.store.init(this.store.formKey());
          this.cancelCreate();
          this.store.selectLayout(layout);
        },
        error: (err) => {
          this.createError.set(err?.error?.message ?? 'Failed to create layout');
        },
      });
  }

  cancelCreate(): void {
    this.showCreateForm.set(false);
    this.newLayoutName = '';
    this.newLayoutDisplayName = '';
    this.createError.set(null);
  }
}
