import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DragDropModule } from '@angular/cdk/drag-drop';
import { FieldType } from '../models/layout-definition.model';

/**
 * One field row in the section canvas — shows label, type badge, drag handle.
 */
@Component({
  selector: 'app-field-card',
  standalone: true,
  imports: [CommonModule, DragDropModule],
  template: `
    <div
      cdkDrag
      class="field-card flex items-center gap-2 px-3 py-2 bg-white border rounded cursor-grab hover:bg-gray-50 transition-colors"
      [class.border-blue-400]="selected()"
      [class.bg-blue-50]="selected()"
      (click)="selectedField.emit(field().name)"
    >
      <span class="text-gray-400 text-xs">⋮⋮</span>
      <span class="type-badge text-xs px-1.5 py-0.5 rounded bg-gray-100 text-gray-600 font-mono">
        {{ field().type }}
      </span>
      <span class="text-sm flex-1 truncate">{{ field().ui.label }}</span>
      @if (hasRequired()) {
        <span class="text-xs text-red-500">*</span>
      }
      <button
        class="text-gray-400 hover:text-red-500 text-xs px-1"
        (click)="$event.stopPropagation(); removeField.emit(field().name)"
        title="Remove field"
      >
        ✕
      </button>
    </div>
  `,
  styles: [`
    .field-card {
      user-select: none;
    }
    .cdk-drag-preview {
      box-sizing: border-box;
      border-radius: 4px;
      box-shadow: 0 5px 5px -3px rgba(0, 0, 0, 0.2),
                  0 8px 10px 1px rgba(0, 0, 0, 0.14),
                  0 3px 14px 2px rgba(0, 0, 0, 0.12);
    }
    .cdk-drag-placeholder {
      opacity: 0.3;
      border: 2px dashed #93c5fd;
      background: #eff6ff;
      border-radius: 4px;
      min-height: 36px;
    }
    .cdk-drag-animating {
      transition: transform 250ms cubic-bezier(0, 0, 0.2, 1);
    }
  `],
})
export class FieldCardComponent {
  field = input.required<FieldType>();
  selected = input(false);

  selectedField = output<string>();
  removeField = output<string>();

  hasRequired(): boolean {
    return this.field().validators?.some((v) => v.kind === 'required') ?? false;
  }
}
