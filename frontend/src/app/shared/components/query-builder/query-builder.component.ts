import { Component, input, output, signal, computed, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import {
  CdkDragDrop,
  DragDropModule,
  moveItemInArray,
  transferArrayItem,
} from '@angular/cdk/drag-drop';
import { OPERATORS, SORT_DIRECTION } from '@core/constants/grids';
import {
  SavedQuery,
  GridQuery,
  QueryFilter,
  QuerySort,
  GridColumn,
} from '@core/models/query.model';
import { QueryService } from '@core/services/query.service';

@Component({
  selector: 'app-query-builder',
  standalone: true,
  imports: [CommonModule, FormsModule, DragDropModule],
  templateUrl: './query-builder.component.html',
  styleUrls: ['./query-builder.component.css'],
})
export class QueryBuilderComponent {
  private readonly queryService = inject(QueryService);

  // Inputs
  gridId = input.required<number>();
  gridName = input.required<string>();
  queries = input<SavedQuery[]>([]);  // Queries del grid (desde el padre)
  initialQuery = input<SavedQuery | null>(null);

  // Outputs
  saved = output<SavedQuery>();
  closed = output<void>();
  queryIdSelected = output<string>();

  // State
  isOpen = signal(false);
  editMode = signal(false);
  activeTab = signal<'fields' | 'sort' | 'filter'>('fields');
  fields = signal<GridColumn[]>([]); // Campos cargados del backend

  // Query selector state
  selectedQueryId = signal<string | null>(null);
  isNewQuery = signal(false);
  isCopyMode = signal(false);

  // Referencias para drag-drop
  availableList: any;
  selectedList: any;

  // Form fields
  name = '';
  isPublic = false;
  isDefault = false;
  selectedFields = signal<number[]>([]);
  // availableFieldsList es un computed que deriva de fields y selectedFields
  // Se recalcula automaticamente cuando cualquiera de los dos cambia
  availableFieldsList = computed(() => {
    const allFields = this.fields();
    const selected = this.selectedFields();
    return allFields.filter(f => !selected.includes(f.id)).map(f => f.id);
  });
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
    return this.fields();
  });

  sortableFields = computed(() => {
    return this.fields().filter((f) => f.sortable);
  });

  filterableFields = computed(() => {
    return this.fields().filter((f) => f.filterable);
  });

  // Solo cargar fields cuando se abre el builder (lazy load)

  // El query builder recibe queries desde el padre, no las carga

  // Cargar fields del backend (solo cuando abre el builder)
  private loadFields(): void {
    if (this.fields().length > 0) return; // Ya cargados
    this.queryService.getFields(this.gridName()).subscribe({
      next: (columns) => this.fields.set(columns),
      error: (err) => console.error('Error loading fields:', err),
    });
  }

  open(query?: SavedQuery) {
    // Cargar fields solo cuando se abre
    console.log('QUERY::::::' , query);
    this.loadFields();

    // Reset selector state
    this.isNewQuery.set(false);
    this.isCopyMode.set(false);
    this.selectedQueryId.set(null);

    // Fallback to initialQuery input if no argument provided
    if (!query && this.initialQuery()) {
      query = this.initialQuery()!;
    }

    if (query) {
      this.editMode.set(true);
      this.name = query.name;
      this.isPublic = query.isPublic;
      this.isDefault = query.isDefault;
      this.selectedFields.set([...query.query.fields]);
      this.currentSort.set([...query.query.sort]);
      this.filters.set(query.query.filters.map((f) => ({ ...f })));
      this.selectedQueryId.set(query.id);
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

    const allFieldIds = this.fields().map((f) => f.id);
    this.selectedFields.set(allFieldIds);

    this.currentSort.set([]);
    this.filters.set([]);
    this.sortField = null;
    this.sortDirection = SORT_DIRECTION.ASC;
    this.newFilter = { field: 0, operator: 1, value: '' };
  }

  // Cuando el usuario selecciona una query del dropdown
  onQueryChange(queryId: string) {
    if (!queryId) {
      this.newQuery();
      return;
    }

    const query = this.queries().find(q => q.id === queryId);
    if (query) {
      this.isNewQuery.set(false);
      this.isCopyMode.set(false);
      this.selectedQueryId.set(query.id);
      this.editMode.set(true);
      this.name = query.name;
      this.isPublic = query.isPublic;
      this.isDefault = query.isDefault;
      this.selectedFields.set([...query.query.fields]);
      this.currentSort.set([...query.query.sort]);
      this.filters.set(query.query.filters.map((f) => ({ ...f })));
    }
  }

  // Crear una query nueva en blanco
  newQuery() {
    this.isNewQuery.set(true);
    this.isCopyMode.set(false);
    this.selectedQueryId.set(null);
    this.editMode.set(false);
    this.name = '';
    this.isPublic = false;
    this.isDefault = false;

    const allFieldIds = this.fields().map((f) => f.id);
    this.selectedFields.set(allFieldIds);

    this.currentSort.set([]);
    this.filters.set([]);
    this.sortField = null;
    this.sortDirection = SORT_DIRECTION.ASC;
    this.newFilter = { field: 0, operator: 1, value: '' };
  }

  // Copiar la query actual con un nuevo nombre
  copyQuery() {
    this.isNewQuery.set(false);
    this.isCopyMode.set(true);
    this.selectedQueryId.set(null);
    this.editMode.set(false);
    this.name = this.name ? `${this.name} (copia)` : '';
    // La config (fields, sort, filters) se mantiene igual
  }

  getFieldLabel(fieldId: number): string {
    const allFields = this.fields();
    const field = allFields.find((f) => f.id === fieldId);
    if (field) return field.label;

    const sortField = this.currentSort().find((s) => s.field === fieldId);
    if (sortField) {
      const sf = allFields.find((f) => f.id === sortField.field);
      if (sf) return sf.label;
    }

    const filter = this.filters().find((f) => f.field === fieldId);
    if (filter) {
      const ff = allFields.find((f) => f.id === filter.field);
      if (ff) return ff.label;
    }

    return String(fieldId);
  }

  getOperatorLabel(operatorId: number): string {
    const op = this.operatorList.find((o) => o.value === operatorId);
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
      // availableFieldsList es computed, se recalcula automaticamente
      this.selectedFields.set([...event.container.data]);
    }
  }

  // Drag & Drop - desde seleccionados a disponibles
  onDropAvailable(event: CdkDragDrop<number[]>) {
    if (event.previousContainer === event.container) {
      moveItemInArray(event.container.data, event.previousIndex, event.currentIndex);
      // availableFieldsList es computed, se recalcula automaticamente
    } else {
      transferArrayItem(
        event.previousContainer.data,
        event.container.data,
        event.previousIndex,
        event.currentIndex
      );
      this.selectedFields.set([...event.previousContainer.data]);
    }
  }

  addSort() {
    if (!this.sortField) return;

    const existing = this.currentSort().filter((s) => s.field !== this.sortField);

    this.currentSort.set([...existing, { field: this.sortField!, direction: this.sortDirection }]);

    this.sortField = null;
  }

  removeSort(fieldId: number) {
    this.currentSort.set(this.currentSort().filter((s) => s.field !== fieldId));
  }

  addFilter() {
    if (!this.newFilter.field || !this.newFilter.operator) return;

    this.filters.update((f) => [
      ...f,
      {
        field: this.newFilter.field,
        operator: this.newFilter.operator as 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10,
        value: this.newFilter.value,
      },
    ]);

    this.newFilter = { field: 0, operator: 1, value: '' };
  }

  removeFilter(index: number) {
    this.filters.update((f) => f.filter((_, i) => i !== index));
  }

  save() {
    if (!this.name || this.selectedFields().length === 0) return;

    const query: GridQuery = {
      fields: this.selectedFields(),
      sort: this.currentSort() as { field: number; direction: 1 | 2 }[],
      filters: this.filters().filter((f) => f.field > 0) as {
        field: number;
        operator: 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10;
        value: unknown;
      }[],
      pagination: { pageSize: 20 },
    };

    const savedQuery: SavedQuery = {
      id: this.initialQuery()?.id || '',
      gridId: this.initialQuery()?.gridId || 0,
      gridName: this.gridName(),
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
