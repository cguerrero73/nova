import { Component, inject, computed, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import {
  CdkDragDrop,
  DragDropModule,
  moveItemInArray,
  transferArrayItem,
} from '@angular/cdk/drag-drop';
import { FormDesignerStore } from '../state/designer.store';
import { Section, FieldType } from '../models/layout-definition.model';
import { SectionCardComponent } from './section-card.component';
import { PaletteFieldType } from '../models/designer.model';

/**
 * Middle canvas — hosts sections with drop zones.
 * Manages cross-section field moves and palette-to-section drops.
 */
@Component({
  selector: 'app-section-canvas',
  standalone: true,
  imports: [CommonModule, FormsModule, DragDropModule, SectionCardComponent],
  template: `
    <div class="section-canvas p-4 min-h-[400px]">
      <!-- Add section button -->
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-800">Canvas</h2>
        <button
          class="text-sm bg-blue-600 text-white px-3 py-1.5 rounded hover:bg-blue-700"
          (click)="addSection()"
        >
          + Add Section
        </button>
      </div>

      @for (section of store.sortedSections(); track section.name; let i = $index) {
        <app-section-card
          [section]="section"
          [listId]="section.name"
          [connectedLists]="allListIds()"
          [store_selectedFieldId]="store.selectedFieldId()"
          (titleChange)="onTitleChange(section.name, $event)"
          (removeSection)="onRemoveSection(section.name)"
          (selectedField)="store.selectField($event)"
          (removeField)="onRemoveField(section.name, $event)"
          (fieldDropped)="onFieldDropped($event)"
          (fieldMoved)="onFieldMoved($event)"
        />
      } @empty {
        <div class="text-center text-gray-400 py-16 border-2 border-dashed border-gray-200 rounded-lg">
          <p class="text-sm">No sections yet.</p>
          <p class="text-xs mt-1">Click "Add Section" or drag fields from the palette.</p>
        </div>
      }

      <!-- Validation errors -->
      @if (store.validationErrors().length > 0) {
        <div class="mt-4 p-3 bg-red-50 border border-red-200 rounded">
          <h4 class="text-sm font-semibold text-red-800 mb-1">Validation Errors</h4>
          @for (err of store.validationErrors(); track err) {
            <p class="text-xs text-red-600">{{ err }}</p>
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .section-canvas {
      background: #f9fafb;
    }
  `],
})
export class SectionCanvasComponent {
  readonly store = inject(FormDesignerStore);

  private sectionCounter = signal(0);

  allListIds = computed<string[]>(() =>
    this.store.sortedSections().map((s) => s.name),
  );

  addSection(): void {
    const count = this.sectionCounter() + 1;
    this.sectionCounter.set(count);
    const section: Section = {
      name: `section_${count}`,
      title: `Section ${count}`,
      order: this.store.sortedSections().length,
      fields: [],
    };
    this.store.addSection(section);
  }

  onTitleChange(sectionName: string, title: string): void {
    const section = this.store.sortedSections().find((s) => s.name === sectionName);
    if (section) {
      this.store.updateSection(sectionName, { ...section, title });
    }
  }

  onRemoveSection(sectionName: string): void {
    this.store.removeSection(sectionName);
  }

  onRemoveField(sectionName: string, fieldName: string): void {
    this.store.removeField(sectionName, fieldName);
  }

  onFieldDropped(event: { sectionName: string; field: FieldType; index: number }): void {
    const def = this.store.definition();
    if (!def) return;

    // Remove from source section if it was a move
    let sourceSectionName: string | null = null;
    for (const s of def.sections) {
      if (s.fields.some((f) => f.name === event.field.name)) {
        sourceSectionName = s.name;
        break;
      }
    }

    const sections = def.sections.map((s) => {
      // Remove from source
      if (sourceSectionName && s.name === sourceSectionName) {
        const filtered = s.fields.filter((f) => f.name !== event.field.name);
        // If same section (reorder), insert at new index
        if (s.name === event.sectionName) {
          const fields = [...filtered];
          fields.splice(event.index, 0, event.field);
          return { ...s, fields };
        }
        return { ...s, fields: filtered };
      }
      // Insert into target
      if (s.name === event.sectionName) {
        const fields = [...s.fields];
        fields.splice(event.index, 0, event.field);
        return { ...s, fields };
      }
      return s;
    });

    this.store.updateDefinition({ ...def, sections });
  }

  onFieldMoved(event: { sectionName: string; previousIndex: number; currentIndex: number }): void {
    const def = this.store.definition();
    if (!def) return;

    const sections = def.sections.map((s) => {
      if (s.name !== event.sectionName) return s;
      const fields = [...s.fields];
      moveItemInArray(fields, event.previousIndex, event.currentIndex);
      return { ...s, fields };
    });

    this.store.updateDefinition({ ...def, sections });
  }
}
