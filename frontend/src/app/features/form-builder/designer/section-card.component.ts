import { Component, input, output, ViewChild, TemplateRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import {
  CdkDragDrop,
  DragDropModule,
  moveItemInArray,
  transferArrayItem,
} from '@angular/cdk/drag-drop';
import { Section, FieldType } from '../models/layout-definition.model';
import { FieldCardComponent } from './field-card.component';
import { PaletteFieldType } from '../models/designer.model';

/**
 * One section in the canvas — has title, list of draggable fields,
 * and a drop zone for accepting fields from the palette or other sections.
 */
@Component({
  selector: 'app-section-card',
  standalone: true,
  imports: [CommonModule, FormsModule, DragDropModule, FieldCardComponent],
  template: `
    <div class="section-card border rounded-lg bg-white">
      <!-- Section header -->
      <div class="flex items-center justify-between px-3 py-2 border-b bg-gray-50 rounded-t-lg">
        <div class="flex items-center gap-2">
          <span class="text-gray-400 text-xs cursor-grab" cdkDragHandle>⋮⋮</span>
          <input
            type="text"
            [ngModel]="section().title || ''"
            (ngModelChange)="titleChange.emit($event)"
            class="text-sm font-semibold bg-transparent border-none outline-none focus:bg-white focus:border focus:rounded px-1 py-0.5"
            placeholder="Section title"
          />
        </div>
        <button
          class="text-gray-400 hover:text-red-500 text-xs px-2 py-1"
          (click)="removeSection.emit()"
          title="Remove section"
        >
          ✕
        </button>
      </div>

      <!-- Fields list with drop zone -->
      <div
        cdkDropList
        [cdkDropListData]="section().fields"
        [cdkDropListConnectedTo]="connectedLists()"
        [id]="listId()"
        (cdkDropListDropped)="onDrop($event)"
        class="field-list p-2 min-h-[60px] space-y-1"
        [class.drop-zone-active]="false"
      >
        @for (field of section().fields; track field.name) {
          <app-field-card
            [field]="field"
            [selected]="store_selectedFieldId() === field.name"
            (selectedField)="selectedField.emit($event)"
            (removeField)="removeField.emit($event)"
          />
        } @empty {
          <div class="drop-hint text-center text-xs text-gray-400 py-4 border-2 border-dashed border-gray-200 rounded">
            Drop fields here
          </div>
        }
      </div>
    </div>
  `,
  styles: [`
    .section-card {
      margin-bottom: 0.75rem;
    }
    .field-list.cdk-drop-list-receiving {
      background: #eff6ff;
    }
  `],
})
export class SectionCardComponent {
  section = input.required<Section>();
  listId = input.required<string>();
  connectedLists = input<string[]>([]);
  store_selectedFieldId = input<string | null>(null);

  titleChange = output<string>();
  removeSection = output<void>();
  selectedField = output<string>();
  removeField = output<string>();
  fieldDropped = output<{ sectionName: string; field: FieldType; index: number }>();
  fieldMoved = output<{ sectionName: string; previousIndex: number; currentIndex: number }>();

  onDrop(event: CdkDragDrop<FieldType[] | PaletteFieldType[]>): void {
    if (event.previousContainer === event.container) {
      // Reorder within same section
      this.fieldMoved.emit({
        sectionName: this.section().name,
        previousIndex: event.previousIndex,
        currentIndex: event.currentIndex,
      });
    } else {
      // Transfer from palette or another section
      const item = event.previousContainer.data[event.previousIndex];
      // Palette items have `type` and `label` but no `name` or `ui`
      if ('ui' in item) {
        // It's a FieldType — cross-section move
        this.fieldDropped.emit({
          sectionName: this.section().name,
          field: item as FieldType,
          index: event.currentIndex,
        });
      } else {
        // It's a PaletteFieldType — new field creation
        const paletteItem = item as PaletteFieldType;
        const newField = this.createFieldFromPalette(paletteItem, event.currentIndex);
        this.fieldDropped.emit({
          sectionName: this.section().name,
          field: newField,
          index: event.currentIndex,
        });
      }
    }
  }

  private createFieldFromPalette(palette: PaletteFieldType, index: number): FieldType {
    const baseName = `${palette.type}_${Date.now().toString(36)}`;
    const base = {
      name: baseName,
      ui: {
        label: palette.label,
        width: 'full' as const,
      },
      validators: [],
    };

    switch (palette.type) {
      case 'select':
      case 'radio':
      case 'multiselect':
        return { ...base, type: palette.type, options: [] } as FieldType;
      default:
        return { ...base, type: palette.type } as FieldType;
    }
  }
}
