import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DragDropModule } from '@angular/cdk/drag-drop';
import { PALETTE_FIELD_TYPES, PaletteFieldType } from '../models/designer.model';

/**
 * Left sidebar with 8 field types to drag into sections.
 * Uses CDK DragDrop — items are cloned when dropped into a section.
 */
@Component({
  selector: 'app-field-palette',
  standalone: true,
  imports: [CommonModule, DragDropModule],
  template: `
    <div class="field-palette">
      <h3 class="text-sm font-semibold text-gray-700 mb-2">Field Types</h3>
      <div
        cdkDropList
        #paletteList="cdkDropList"
        [cdkDropListData]="paletteItems"
        [cdkDropListSortingDisabled]="true"
        class="space-y-1"
      >
        @for (item of paletteItems; track item.type) {
          <div
            cdkDrag
            class="palette-item flex items-center gap-2 px-3 py-2 bg-white border rounded cursor-grab hover:bg-gray-50 transition-colors"
          >
            <span class="text-gray-400 text-xs">⋮⋮</span>
            <span class="field-icon w-6 h-6 flex items-center justify-center bg-blue-100 text-blue-700 rounded text-xs font-bold">
              {{ item.icon }}
            </span>
            <span class="text-sm">{{ item.label }}</span>
          </div>
        }
      </div>
    </div>
  `,
  styles: [`
    .field-palette {
      padding: 0.5rem;
    }
    .palette-item {
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
      opacity: 0;
    }
    .cdk-drag-animating {
      transition: transform 250ms cubic-bezier(0, 0, 0.2, 1);
    }
  `],
})
export class FieldPaletteComponent {
  paletteItems: PaletteFieldType[] = PALETTE_FIELD_TYPES;
}
