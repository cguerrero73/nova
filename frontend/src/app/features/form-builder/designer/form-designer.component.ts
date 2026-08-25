import { Component, OnInit, OnDestroy, inject, signal, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { FormDesignerStore } from '../state/designer.store';
import { LayoutPickerComponent } from './layout-picker.component';
import { AssignmentPanelComponent } from './assignment-panel.component';
import { FieldPaletteComponent } from './field-palette.component';
import { SectionCanvasComponent } from './section-canvas.component';
import { FieldSettingsComponent } from './field-settings.component';
import { PreviewDialogComponent } from './preview-dialog.component';
import { AuditPanelComponent } from './audit-panel.component';

/**
 * Main designer page — 3-column layout: palette | canvas | settings.
 * Top bar with save/publish/preview actions.
 */
@Component({
  selector: 'app-form-designer',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    LayoutPickerComponent,
    AssignmentPanelComponent,
    FieldPaletteComponent,
    SectionCanvasComponent,
    FieldSettingsComponent,
    PreviewDialogComponent,
    AuditPanelComponent,
  ],
  template: `
    <div class="form-designer h-full flex flex-col">
      <!-- Top bar -->
      <div class="flex items-center justify-between px-4 py-2 border-b bg-white">
        <div class="flex items-center gap-3">
          <h1 class="text-lg font-semibold text-gray-800">Form Designer</h1>
          @if (store.currentLayout(); as layout) {
            <span class="text-sm text-gray-500">{{ layout.fl_display_name || layout.fl_name }}</span>
            @if (store.dirty()) {
              <span class="text-xs bg-yellow-100 text-yellow-800 px-1.5 py-0.5 rounded">Unsaved</span>
            }
            @if (store.lastSavedAt(); as saved) {
              <span class="text-xs text-gray-400">Saved {{ formatTime(saved) }}</span>
            }
          }
        </div>
        <div class="flex items-center gap-2">
          <button
            class="px-3 py-1.5 text-sm border rounded hover:bg-gray-50"
            (click)="previewDialog.open()"
            [disabled]="!store.definition()"
          >
            Preview
          </button>
          <button
            class="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            (click)="store.saveDraft()"
            [disabled]="!store.dirty() || store.saving()"
          >
            {{ store.saving() ? 'Saving...' : 'Save Draft' }}
          </button>
          <button
            class="px-3 py-1.5 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
            (click)="showPublishDialog.set(true)"
            [disabled]="!store.currentLayout() || store.saving()"
          >
            Publish
          </button>
        </div>
      </div>

      <!-- Validation errors banner -->
      @if (store.validationErrors().length > 0) {
        <div class="px-4 py-2 bg-red-50 border-b border-red-200">
          <p class="text-xs text-red-800 font-semibold">Validation failed:</p>
          @for (err of store.validationErrors(); track err) {
            <p class="text-xs text-red-600">{{ err }}</p>
          }
        </div>
      }

      <!-- Main content: 3-column layout -->
      <div class="flex-1 flex overflow-hidden">
        <!-- Left sidebar: palette + layout picker -->
        <div class="w-56 border-r bg-gray-50 overflow-y-auto flex-shrink-0">
          <app-field-palette />
          <div class="border-t mt-2 pt-2">
            <app-layout-picker />
          </div>
          <div class="border-t mt-2 pt-2">
            <app-assignment-panel />
          </div>
          <app-audit-panel />
        </div>

        <!-- Center: canvas -->
        <div class="flex-1 overflow-y-auto">
          @if (store.loading()) {
            <div class="text-center text-gray-400 py-16">Loading...</div>
          } @else if (!store.currentLayout()) {
            <div class="text-center text-gray-400 py-16">
              <p class="text-sm">Select a layout to start designing.</p>
            </div>
          } @else {
            <app-section-canvas />
          }
        </div>

        <!-- Right sidebar: field settings -->
        <div class="w-72 overflow-y-auto flex-shrink-0 bg-white">
          <app-field-settings />
        </div>
      </div>

      <!-- Preview dialog -->
      <app-preview-dialog #previewDialog />

      <!-- Publish dialog -->
      @if (showPublishDialog()) {
        <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" (click)="showPublishDialog.set(false)">
          <div
            class="bg-white rounded-lg shadow-xl w-full max-w-md p-6"
            (click)="$event.stopPropagation()"
          >
            <h3 class="text-lg font-semibold mb-3">Publish Layout</h3>
            <p class="text-sm text-gray-600 mb-3">
              Enter a description for this version (commit message):
            </p>
            <textarea
              [(ngModel)]="publishDescription"
              class="w-full px-3 py-2 border rounded text-sm"
              rows="3"
              placeholder="Describe what changed..."
            ></textarea>
            <div class="flex justify-end gap-2 mt-4">
              <button
                class="px-4 py-2 text-sm border rounded hover:bg-gray-50"
                (click)="cancelPublish()"
              >
                Cancel
              </button>
              <button
                class="px-4 py-2 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
                (click)="confirmPublish()"
                [disabled]="!publishDescription.trim()"
              >
                Publish
              </button>
            </div>
          </div>
        </div>
      }
    </div>
  `,
  styles: [`
    .form-designer {
      min-height: 0;
    }
  `],
})
export class FormDesignerComponent implements OnInit, OnDestroy {
  readonly store = inject(FormDesignerStore);
  private readonly route = inject(ActivatedRoute);

  @ViewChild('previewDialog') previewDialog!: PreviewDialogComponent;

  showPublishDialog = signal(false);
  publishDescription = '';

  ngOnInit(): void {
    const formKey = this.route.snapshot.paramMap.get('formKey');
    const layoutName = this.route.snapshot.paramMap.get('layoutName');

    if (formKey) {
      this.store.init(formKey);

      // If layoutName is in the route, auto-select it after layouts load
      if (layoutName) {
        // Wait for layouts to load, then select
        const checkLayouts = () => {
          const layout = this.store.layouts().find((l) => l.fl_name === layoutName);
          if (layout) {
            this.store.selectLayout(layout);
          } else {
            setTimeout(checkLayouts, 100);
          }
        };
        setTimeout(checkLayouts, 200);
      }
    }
  }

  ngOnDestroy(): void {
    this.store.reset();
  }

  formatTime(iso: string): string {
    const d = new Date(iso);
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  confirmPublish(): void {
    this.store.publishLayout(this.publishDescription.trim());
    this.cancelPublish();
  }

  cancelPublish(): void {
    this.showPublishDialog.set(false);
    this.publishDescription = '';
  }
}
