import { Component, input, output, signal, computed, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CdkDragDrop, DragDropModule, moveItemInArray, transferArrayItem } from '@angular/cdk/drag-drop';
import { 
  GridId, 
  getGridFields, 
  OPERATORS, 
  SORT_DIRECTION 
} from '@core/constants/grids';
import { SavedQuery, GridQuery, QueryFilter, QuerySort } from '@core/models/query.model';

@Component({
  selector: 'app-query-builder',
  standalone: true,
  imports: [CommonModule, FormsModule, DragDropModule],
  templateUrl: './query-builder.component.html',
  styleUrls: ['./query-builder.component.css']
})
export class QueryBuilderComponent implements OnInit {
  // Inputs
  gridId = input.required<number>();
  initialQuery = input<SavedQuery | null>(null);
  
  // Outputs
  saved = output<SavedQuery>();
  closed = output<void>();

  // State
  isOpen = signal(false);
  editMode = signal(false);
  activeTab = signal<'fields' | 'sort' | 'filter'>('fields');
  
  // Referencias para drag-drop
  availableList: any;
  selectedList: any;
  
  // Form fields
  name = '';
  isPublic = false;
  isDefault = false;
  selectedFields = signal<number[]>([]);
  availableFieldsList = signal<number[]>([]);
  sortField: number | null = null;
  sortDirection = SORT_DIRECTION.ASC;
  filters = signal<QueryFilter[]>([]);
  currentSort = signal<QuerySort[]>([]);
  
  // Nuevo filtro
  newFilter = { field: 0, operator: 1, value: '' };

  // Helpers para el template
  isNullOperator = OPERATORS.IS_NULL;
  isNotNullOperator = OPERATORS.IS_NOT_NULL;

  // Constants
  SORT_DIRECTION = SORT_DIRECTION;
  OPERATORS = OPERATORS;

  operatorList = [
    { value: OPERATORS.EQ, label: '=' },
    { value: OPERATORS.NE, label: '≠' },
    { value: OPERATORS.CONTAINS, label: 'contiene' },
    { value: OPERATORS.GT, label: '>' },
    { value: OPERATORS.LT, label: '<' },
    { value: OPERATORS.GTE, label: '>=' },
    { value: OPERATORS.LTE, label: '<=' },
    { value: OPERATORS.IN, label: 'en' },
    { value: OPERATORS.IS_NULL, label: 'es nulo' },
    { value: OPERATORS.IS_NOT_NULL, label: 'no es nulo' },
  ];

  // Computed
  availableFields = computed(() => {
    return getGridFields(this.gridId() as GridId);
  });

  sortableFields = computed(() => {
    return this.availableFields().filter(f => f.sortable);
  });

  filterableFields = computed(() => {
    return this.availableFields().filter(f => f.filterable);
  });

  ngOnInit() {
    this.reset();
  }

  open(query?: SavedQuery) {
    if (query) {
      this.editMode.set(true);
      this.name = query.name;
      this.isPublic = query.isPublic;
      this.isDefault = query.isDefault;
      this.selectedFields.set([...query.query.fields]);
      this.currentSort.set([...query.query.sort]);
      this.filters.set(query.query.filters.map(f => ({ ...f })));
      this.updateAvailableFields();
    } else {
      this.reset();
    }
    this.isOpen.set(true);
  }

  close() {
    this.isOpen.set(false);
    this.closed.emit();
  }

  reset() {
    this.editMode.set(false);
    this.activeTab.set('fields');
    this.name = '';
    this.isPublic = false;
    this.isDefault = false;
    
    const fields = getGridFields(this.gridId() as GridId);
    const allFieldIds = fields.map(f => f.id);
    this.selectedFields.set(allFieldIds);
    this.availableFieldsList.set([]);
    
    this.currentSort.set([]);
    this.filters.set([]);
    this.sortField = null;
    this.sortDirection = SORT_DIRECTION.ASC;
    this.newFilter = { field: 0, operator: 1, value: '' };
  }

  updateAvailableFields() {
    const allFields = this.availableFields();
    const selected = this.selectedFields();
    const available = allFields
      .filter(f => !selected.includes(f.id))
      .map(f => f.id);
    this.availableFieldsList.set(available);
  }

  getFieldLabel(fieldId: number): string {
    const allFields = getGridFields(this.gridId() as GridId);
    const field = allFields.find(f => f.id === fieldId);
    if (field) return field.label;
    
    const sortField = this.currentSort().find(s => s.field === fieldId);
    if (sortField) {
      const sf = allFields.find(f => f.id === sortField.field);
      if (sf) return sf.label;
    }
    
    const filter = this.filters().find(f => f.field === fieldId);
    if (filter) {
      const ff = allFields.find(f => f.id === filter.field);
      if (ff) return ff.label;
    }
    
    return String(fieldId);
  }

  getOperatorLabel(operatorId: number): string {
    const op = this.operatorList.find(o => o.value === operatorId);
    return op?.label || '=';
  }

  // Drag & Drop - desde disponibles a seleccionados
  onDropSelected(event: CdkDragDrop<number[]>) {
    if (event.previousContainer === event.container) {
      moveItemInArray(event.container.data, event.previousIndex, event.currentIndex);
      this.selectedFields.set([...event.container.data]);
    } else {
      transferArrayItem(
        event.previousContainer.data,
        event.container.data,
        event.previousIndex,
        event.currentIndex
      );
      this.selectedFields.set([...event.container.data]);
      this.availableFieldsList.set([...event.previousContainer.data]);
    }
  }

  // Drag & Drop - desde seleccionados a disponibles
  onDropAvailable(event: CdkDragDrop<number[]>) {
    if (event.previousContainer === event.container) {
      moveItemInArray(event.container.data, event.previousIndex, event.currentIndex);
      this.availableFieldsList.set([...event.container.data]);
    } else {
      transferArrayItem(
        event.previousContainer.data,
        event.container.data,
        event.previousIndex,
        event.currentIndex
      );
      this.selectedFields.set([...event.previousContainer.data]);
      this.availableFieldsList.set([...event.container.data]);
    }
  }

  addSort() {
    if (!this.sortField) return;
    
    const existing = this.currentSort().filter(s => s.field !== this.sortField);
    
    this.currentSort.set([
      ...existing,
      { field: this.sortField!, direction: this.sortDirection }
    ]);
    
    this.sortField = null;
  }

  removeSort(fieldId: number) {
    this.currentSort.set(this.currentSort().filter(s => s.field !== fieldId));
  }

  addFilter() {
    if (!this.newFilter.field || !this.newFilter.operator) return;
    
    this.filters.update(f => [
      ...f, 
      { 
        field: this.newFilter.field, 
        operator: this.newFilter.operator as 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10, 
        value: this.newFilter.value 
      }
    ]);
    
    this.newFilter = { field: 0, operator: 1, value: '' };
  }

  removeFilter(index: number) {
    this.filters.update(f => f.filter((_, i) => i !== index));
  }

  save() {
    if (!this.name || this.selectedFields().length === 0) return;

    const query: GridQuery = {
      fields: this.selectedFields(),
      sort: this.currentSort() as { field: number; direction: 1 | 2 }[],
      filters: this.filters().filter(f => f.field > 0) as { field: number; operator: 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10; value: unknown }[],
      pagination: { pageSize: 20 },
    };

    const savedQuery: SavedQuery = {
      id: this.initialQuery()?.id || '',
      gridId: this.gridId(),
      name: this.name,
      userId: null,
      isPublic: this.isPublic,
      isDefault: this.isDefault,
      query,
      createdAt: this.initialQuery()?.createdAt || new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    this.saved.emit(savedQuery);
    this.close();
  }
}
