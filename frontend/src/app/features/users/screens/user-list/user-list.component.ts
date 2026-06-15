import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { Router } from '@angular/router';
import { createAngularTable, getCoreRowModel, ColumnDef } from '@tanstack/angular-table';
import { UserService } from '../../services/user.service';
import { QueryService } from '@core/services/query.service';
import { TranslationService } from '@core/services/translation.service';
import { UiStore } from '@core/stores/ui.store';
import { GRID_IDS } from '@core/constants/grids';
import { User } from '../../models/user.model';
import { SavedQuery, GridQuery, GridColumn } from '@core/models/query.model';
import { DataGridComponent, GridMeta } from '@shared/components/data-grid/data-grid.component';
import { QueryBuilderComponent } from '@shared/components/query-builder/query-builder.component';
import { ToolbarComponent } from '@shared/components/toolbar/toolbar.component';
import {
  EntityDetailComponent,
  EntityTab,
} from '@shared/components/entity-detail/entity-detail.component';
import {
  EntityFormComponent,
  FormField,
} from '@shared/components/entity-form/entity-form.component';
import {
  RelatedInfoComponent,
  RelatedSection,
} from '@shared/components/related-info/related-info.component';

type UserRow = User;

@Component({
  selector: 'app-user-list',
  standalone: true,
  imports: [
    DataGridComponent,
    QueryBuilderComponent,
    ToolbarComponent,
    EntityDetailComponent,
    EntityFormComponent,
    RelatedInfoComponent,
  ],
  templateUrl: './user-list.component.html',
})
export class UserListComponent implements OnInit {
  readonly userService = inject(UserService);
  readonly queryService = inject(QueryService);
  readonly translate = inject(TranslationService);
  readonly uiStore = inject(UiStore);
  readonly router = inject(Router);
  readonly GRID_IDS = GRID_IDS;

  // Traducciones
  t: Record<string, string> = {};

  private searchTimeout?: ReturnType<typeof setTimeout>;
  selected = signal<User | null>(null);
  currentGridId = signal<number | null>(null);  // Grid ID obtenido de la config

  selectedUser = this.selected.asReadonly();
  selectedQueryId = signal<string>('');

  // Estado del drawer de detalle
  showDetail = signal<boolean>(false);
  detailEntity = signal<User | null>(null);
  detailLoading = signal<boolean>(false);
  detailSaving = signal<boolean>(false);

  // Track de cambios sin guardar en el drawer
  hasUnsavedChanges = signal<boolean>(false);

  // Computed: navegación previa/siguiente (siempre visible, funcionando basado en selección actual)
  canNavigatePrev = computed(() => {
    // Si el drawer está abierto, navegar dentro del detailEntity
    if (this.showDetail()) {
      const currentId = this.detailEntity()?.id;
      if (!currentId) return false;
      const users = this.userService.users();
      const currentIndex = users.findIndex((u) => u.id === currentId);
      return currentIndex > 0;
    }
    // Si no hay drawer abierto, usar selected
    const currentId = this.selected()?.id;
    if (!currentId) return false;
    const users = this.userService.users();
    const currentIndex = users.findIndex((u) => u.id === currentId);
    return currentIndex > 0;
  });

  canNavigateNext = computed(() => {
    // Si el drawer está abierto, navegar dentro del detailEntity
    if (this.showDetail()) {
      const currentId = this.detailEntity()?.id;
      if (!currentId) return false;
      const users = this.userService.users();
      const currentIndex = users.findIndex((u) => u.id === currentId);
      return currentIndex < users.length - 1;
    }
    // Si no hay drawer abierto, usar selected
    const currentId = this.selected()?.id;
    if (!currentId) return false;
    const users = this.userService.users();
    const currentIndex = users.findIndex((u) => u.id === currentId);
    return currentIndex < users.length - 1;
  });

  // Tabs de información relacionada (se construyen con traducciones en ngOnInit)
  tabs: EntityTab[] = [];

  // Campos del formulario de edición (se construyen con traducciones en ngOnInit)
  formFields: FormField[] = [];

  // Información relacionada (mock)
  relatedSections: RelatedSection[] = [
    {
      id: 'comments',
      title: 'Comentarios',
      items: [
        {
          id: '1',
          title: 'Usuario creado desde admin',
          description: 'Creado por admin@nova.com',
          date: '2024-01-15 10:30',
        },
        {
          id: '2',
          title: 'Email verificado',
          description: 'Verificación automática',
          date: '2024-01-15 10:35',
        },
      ],
      emptyMessage: 'Sin comentarios',
    },
    {
      id: 'documents',
      title: 'Documentos',
      items: [],
      emptyMessage: 'Sin documentos adjuntos',
    },
    {
      id: 'audit',
      title: 'Historial de Cambios',
      items: [
        { id: '1', title: 'Creación del registro', date: '2024-01-15 10:30:00' },
        { id: '2', title: 'Actualización de email', date: '2024-01-16 14:22:00' },
      ],
      emptyMessage: 'Sin historial',
    },
  ];

  gridMeta = computed<GridMeta>(() => ({
    page: this.queryService.currentMeta().page,
    pageSize: this.queryService.currentMeta().pageSize,
    total: this.queryService.currentMeta().total,
    totalPages: this.queryService.currentMeta().totalPages,
  }));

  selectedQueryForBuilder = computed(() => {
    return this.queryService.selectedQuery();
  });

  // Columnas del grid (se construyen con traducciones en ngOnInit)
  columns: ColumnDef<UserRow>[] = [];

  // Campos del grid cargados del backend (para construir columnas dinámicamente)
  gridColumns = signal<GridColumn[]>([]);

  // Nombre del grid para esta pantalla
  readonly GRID_NAME = 'BMUSER';

  ngOnInit() {
    console.log('[UserList] Starting unified flow for grid:', this.GRID_NAME);

    // FLUJO UNIFICADO:
    // 1. Cargar traducciones
    // 2. Cargar config del grid por nombre → obtener metadata de campos
    // 3. Construir columnas desde campos del backend
    // 4. Cargar queries del grid
    // 5. Ejecutar query default

    this.translate.load('users').subscribe({
      next: (translations) => {
        console.log('[UserList] Translations loaded:', translations);
        this.t = translations;
        this.buildTabs();
        this.buildFormFields();

        // Continuar con el flujo: cargar config del grid
        this.loadGridConfig();
      },
      error: (err) => {
        console.error('[UserList] Error loading translations:', err);
      },
    });
  }

  // PASO 2: Cargar config del grid por nombre
  private loadGridConfig(): void {
    console.log('[UserList] Step 2: Loading grid config for', this.GRID_NAME);

    this.queryService.getConfig(this.GRID_NAME).subscribe({
      next: (config) => {
        if (!config) {
          console.error('[UserList] No config returned');
          this.buildColumns();
          this.loadGridQueries(GRID_IDS.USERS);
          return;
        }

        console.log('[UserList] Grid config loaded, gridId:', config.gridId);
        this.currentGridId.set(config.gridId);
        this.gridColumns.set(config.columns);

        // Construir columnas dinámicamente desde el backend
        this.buildColumnsFromBackend(config.columns);

        // Continuar con el flujo: cargar queries usando el gridId
        this.loadGridQueries(config.gridId);
      },
      error: (err) => {
        console.error('[UserList] Error loading grid config:', err);
        // Fallback: usar columnas por defecto
        this.buildColumns();
        this.loadGridQueries(GRID_IDS.USERS);
      },
    });
  }

  // PASO 3: Cargar queries del grid
  private loadGridQueries(gridId: number): void {
    console.log('[UserList] Step 3: Loading queries for gridId', gridId);

    this.queryService.loadByGridId(gridId).subscribe({
      next: (queries) => {
        console.log('[UserList] Queries loaded:', queries.length);

        // Buscar query default
        const defaultQuery = queries.find((q) => q.isDefault);
        if (defaultQuery) {
          console.log('[UserList] Default query found:', defaultQuery.id, defaultQuery.name);
          this.selectedQueryId.set(defaultQuery.id);
          this.executeQuery(defaultQuery.query);
        } else {
          console.log('[UserList] No default query found, queries:', queries.map(q => ({id: q.id, name: q.name, isDefault: q.isDefault})));
          this.loadDefaultData();
        }
      },
      error: (err) => {
        console.error('[UserList] Error loading queries:', err);
        this.loadDefaultData();
      },
    });
  }

  // Construir tabs con traducciones
  private buildTabs(): void {
    this.tabs = [
      { id: 'view', label: this.t['tabs.view'] || 'Visualización del Registro' },
      { id: 'comments', label: this.t['tabs.comments'] || 'Comentarios' },
      { id: 'documents', label: this.t['tabs.documents'] || 'Documentos' },
      { id: 'audit', label: this.t['tabs.audit'] || 'Historial' },
    ];
  }

  // Construir form fields con traducciones
  private buildFormFields(): void {
    this.formFields = [
      { key: 'id', label: this.t['form.id.label'] || 'ID', type: 'text', readonly: true },
      {
        key: 'name',
        label: this.t['form.name.label'] || 'Nombre',
        type: 'text',
        required: true,
        placeholder: this.t['form.name.placeholder'] || 'Nombre completo',
      },
      {
        key: 'email',
        label: this.t['form.email.label'] || 'Email',
        type: 'email',
        required: true,
        placeholder: this.t['form.email.placeholder'] || 'correo@ejemplo.com',
      },
      {
        key: 'status',
        label: this.t['form.status.label'] || 'Estado',
        type: 'select',
        required: true,
        options: [
          { value: 'active', label: this.t['status.active'] || 'Activo' },
          { value: 'inactive', label: this.t['status.inactive'] || 'Inactivo' },
          { value: 'suspended', label: this.t['status.suspended'] || 'Suspendido' },
        ],
      },
      {
        key: 'createdAt',
        label: this.t['form.createdAt.label'] || 'Fecha Creación',
        type: 'date',
        readonly: true,
      },
    ];
  }

  // Construir columnas del grid con traducciones (fallback si no hay campos del backend)
  private buildColumns(): void {
    this.columns = [
      {
        accessorKey: 'id',
        header: this.t['form.id.label'] || 'ID',
        size: 80,
      },
      {
        accessorKey: 'name',
        header: this.t['form.name.label'] || 'Nombre',
        size: 200,
      },
      {
        accessorKey: 'email',
        header: this.t['form.email.label'] || 'Email',
      },
      {
        accessorKey: 'status',
        header: this.t['form.status.label'] || 'Estado',
        cell: (info: { getValue: () => unknown }) => {
          const status = info.getValue() as string;
          const activeLabel = this.t['status.active'] || 'Activo';
          const inactiveLabel = this.t['status.inactive'] || 'Inactivo';
          return status === 'active' ? activeLabel : inactiveLabel;
        },
      },
      {
        accessorKey: 'createdAt',
        header: this.t['form.createdAt.label'] || 'Fecha Creación',
        cell: (info: { getValue: () => unknown }) => {
          const date = new Date(info.getValue() as string);
          return date.toLocaleDateString();
        },
      },
    ];
  }

  // Construir columnas del grid desde campos del backend
  private buildColumnsFromBackend(columns: GridColumn[]): void {
    console.log('[UserList] Building columns from backend, count:', columns.length);

    this.columns = columns.map((col) => {
      const columnDef: ColumnDef<UserRow> = {
        id: String(col.id),
        accessorKey: col.key,
        header: col.label,
        size: this.getColumnSize(col.type),
      };

      // Agregar cell renderer según tipo de dato
      if (col.type === 'date') {
        columnDef.cell = (info: { getValue: () => unknown }) => {
          const value = info.getValue();
          if (!value) return '';
          const date = new Date(value as string);
          return date.toLocaleDateString();
        };
      }

      if (col.type === 'boolean') {
        columnDef.cell = (info: { getValue: () => unknown }) => {
          const value = info.getValue() as boolean;
          return value ? '✓' : '✗';
        };
      }

      if (col.type === 'select') {
        columnDef.cell = (info: { getValue: () => unknown }) => {
          return String(info.getValue() ?? '');
        };
      }

      return columnDef;
    });

    console.log('[UserList] Columns built:', this.columns.length);
  }

  // Obtener tamaño de columna según tipo de dato
  private getColumnSize(type: GridColumn['type']): number {
    switch (type) {
      case 'number':
        return 100;
      case 'date':
        return 120;
      case 'boolean':
        return 80;
      case 'string':
      default:
        return 200;
    }
  }

  loadDefaultData() {
    const defaultQuery: GridQuery = {
      fields: [1, 2, 3, 4, 5],
      sort: [{ field: 2, direction: 1 as const }],
      filters: [],
      pagination: { pageSize: 20 },
    };
    this.executeQuery(defaultQuery);
  }

  // Track de última query ejecutada para detectar cambios
  private lastExecutedQuery: string = '';
  private loadedPages = new Set<number>();

  executeQuery(query: GridQuery, page: number = 1) {
    // Generar un ID único de la query para detectar cambios
    const querySignature = JSON.stringify(query) + '_' + page;

    // Resetear páginas cargadas si cambió la query o es página 1
    if (page === 1 || this.lastExecutedQuery !== querySignature) {
      this.loadedPages.clear();
      this.lastExecutedQuery = querySignature;
    }

    // Evitar cargar la misma página dos veces
    if (this.loadedPages.has(page)) {
      console.log('[UserList] Page already loaded:', page);
      return;
    }

    this.loadedPages.add(page);
    this.uiStore.setLoading(true);

    // Use queryId if a saved query is selected, otherwise use direct query
    const selectedQueryId = this.selectedQueryId();
    if (selectedQueryId) {
      this.queryService.executeQuery(selectedQueryId, page, 20).subscribe({
        next: (response) => {
          if (page === 1) {
            this.userService.users.set(response.data as User[]);
          } else {
            this.userService.users.update((current) => [...current, ...(response.data as User[])]);
          }
          this.uiStore.setLoading(false);
        },
        error: () => {
          this.uiStore.setLoading(false);
          this.loadedPages.delete(page);
        },
      });
    } else {
      // Use executeQueryDirect for default/custom queries
      this.queryService.executeQueryDirect(
        GRID_IDS.USERS,
        query.fields,
        query.filters,
        query.sort,
        page,
        20
      ).subscribe({
        next: (response) => {
          if (page === 1) {
            this.userService.users.set(response.data as User[]);
          } else {
            this.userService.users.update((current) => [...current, ...(response.data as User[])]);
          }
          this.uiStore.setLoading(false);
        },
        error: () => {
          this.uiStore.setLoading(false);
          this.loadedPages.delete(page);
        },
      });
    }
  }

  onQuerySelect(event: Event) {
    const queryId = (event.target as HTMLSelectElement).value;

    if (!queryId) {
      this.loadDefaultData();
      this.queryService.clearSelection();
      this.selectedQueryId.set('');
      this.loadedPages.clear();
      return;
    }

    const query = this.queryService.queries().find((q) => q.id === queryId);
    if (query) {
      this.queryService.selectQuery(query);
      this.selectedQueryId.set(queryId);
      this.loadedPages.clear();
      this.executeQuery(query.query);
    }
  }

  onQuerySaved(savedQuery: SavedQuery) {
    const gridId = this.currentGridId() ?? GRID_IDS.USERS;
    this.uiStore.setLoading(true);

    const operation = savedQuery.id
      ? this.queryService.update(savedQuery.id, savedQuery)
      : this.queryService.save(savedQuery);

    operation.subscribe({
      next: () => {
        // Recargar queries y desactivar loading
        this.queryService.loadByGridId(gridId).subscribe({
          complete: () => {
            this.uiStore.setLoading(false);
          }
        });
      },
      error: () => {
        this.uiStore.setLoading(false);
      }
    });
  }

  onQueryIdSelected(queryId: string): void {
    console.log('[UserList] Query selected:', queryId);
    this.selectedQueryId.set(queryId);
    this.loadedPages.clear();
  }

  onSearch(event: Event) {
    const value = (event.target as HTMLInputElement).value;
    clearTimeout(this.searchTimeout);
    this.searchTimeout = setTimeout(() => {
      const currentQuery = this.queryService.currentQuery();
      if (currentQuery) {
        const newQuery = {
          ...currentQuery,
          filters: value ? [{ field: 2 as const, operator: 3 as const, value }] : [],
        };
        this.loadedPages.clear();
        this.executeQuery(newQuery);
      }
    }, 300);
  }

  onRowClick(user: User) {
    console.log('Row clicked:', user);
  }

  onUserSelect(user: User) {
    // Toggle selection: if already selected, deselect; otherwise select
    if (this.selected()?.id === user.id) {
      this.selected.set(null);
    } else {
      this.selected.set(user);
    }
  }

  onPageChange(page: number) {
    const currentQuery = this.queryService.currentQuery();
    if (currentQuery) {
      this.executeQuery(currentQuery, page);
    }
  }

  // Toolbar events
  onCreate() {
    // Crear nuevo usuario en el drawer (el mismo formulario de edición)
    const newUser: User = {
      id: '', // ID vacío para indicar que es nuevo
      name: '',
      email: '',
      status: 'active',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    this.detailEntity.set(newUser);
    this.showDetail.set(true);
    this.hasUnsavedChanges.set(true);
  }

  onDelete() {
    const selected = this.selected();
    const message = this.t['messages.delete_confirm'] || '¿Está seguro que desea eliminar?';
    if (selected && confirm(`${message} ${selected.name}?`)) {
      this.userService.deleteUser(selected.id).then((success) => {
        if (success) {
          this.selected.set(null);
        }
      }).catch((err) => console.error('Error deleting user:', err));
    }
  }

  // Guardar desde toolbar (cuando hay cambios en el drawer)
  onSaveFromToolbar() {
    if (this.showDetail() && this.hasUnsavedChanges()) {
      const entity = this.detailEntity();
      if (entity) {
        this.onDetailSave(entity);
      }
    }
  }

  onPrint() {
    window.print();
  }

  onRefresh() {
    // Recargar los datos actuales
    const currentQuery = this.queryService.currentQuery();
    this.loadedPages.clear();
    if (currentQuery) {
      this.executeQuery(currentQuery, 1);
    } else {
      this.loadDefaultData();
    }
  }

  onDuplicate() {
    // Duplicar el registro actual: crear una copia y abrir en modo edición
    const selected = this.selected();
    if (!selected) return;

    // Crear una copia del usuario (nuevo ID, limpio el id actual)
    const duplicated: User = {
      ...selected,
      id: '', // ID vacío para crear nuevo
      name: `${selected.name} (copia)`,
    };

    // Abrir el drawer con el nuevo entity
    this.detailEntity.set(duplicated);
    this.showDetail.set(true);
    this.hasUnsavedChanges.set(true);
  }

  // Navegación previous/next - actualiza el drawer si está abierto, o usa selected si está cerrado
  onPrev() {
    // Determinar el ID actual basado en si el drawer está abierto
    const currentId = this.showDetail() ? this.detailEntity()?.id : this.selected()?.id;

    if (!currentId) return;

    const users = this.userService.users();
    const currentIndex = users.findIndex((u) => u.id === currentId);

    if (currentIndex > 0) {
      const prevUser = users[currentIndex - 1];
      if (this.showDetail()) {
        // Si el drawer está abierto, actualizar el entity del drawer
        this.detailEntity.set(prevUser);
        this.hasUnsavedChanges.set(false);
      } else {
        // Si el drawer está cerrado, solo mover la selección
        this.selected.set(prevUser);
      }
    }
  }

  onNext() {
    // Determinar el ID actual basado en si el drawer está abierto
    const currentId = this.showDetail() ? this.detailEntity()?.id : this.selected()?.id;

    if (!currentId) return;

    const users = this.userService.users();
    const currentIndex = users.findIndex((u) => u.id === currentId);

    if (currentIndex < users.length - 1) {
      const nextUser = users[currentIndex + 1];
      if (this.showDetail()) {
        // Si el drawer está abierto, actualizar el entity del drawer
        this.detailEntity.set(nextUser);
        this.hasUnsavedChanges.set(false);
      } else {
        // Si el drawer está cerrado, solo mover la selección
        this.selected.set(nextUser);
      }
    }
  }

  // === Entity Detail Drawer ===

  onRowDoubleClick(user: User) {
    // Abrir el drawer de detalle con doble click
    this.detailEntity.set(user);
    this.showDetail.set(true);
    this.hasUnsavedChanges.set(false);
  }

  onDetailClose() {
    this.showDetail.set(false);
    this.detailEntity.set(null);
    this.hasUnsavedChanges.set(false);
  }

  onDetailSave(entity: unknown) {
    this.detailSaving.set(true);
    const user = entity as User;
    
    const operation = user.id
      ? this.userService.updateUser(user.id, user)
      : this.userService.createUser(user);
    
    operation.then((savedUser) => {
      this.detailSaving.set(false);
      this.showDetail.set(false);
      this.hasUnsavedChanges.set(false);
      if (user.id) {
        // Update existing
        this.userService.users.update((users) =>
          users.map((u) => (u.id === savedUser?.id ? savedUser : u))
        );
      } else if (savedUser) {
        // Add new
        this.userService.users.update((users) => [...users, savedUser]);
      }
    }).catch((err) => {
      this.detailSaving.set(false);
      console.error('Error saving user:', err);
    });
  }

  // Obtener el valor de un campo del formulario
  getFieldValue(key: string): unknown {
    const entity = this.detailEntity();
    if (!entity) return '';
    return (entity as unknown as Record<string, unknown>)[key] ?? '';
  }

  // Actualizar un campo del entity editable
  onFieldChange(field: string, value: unknown) {
    this.detailEntity.update((entity) => {
      if (!entity) return null;
      return { ...entity, [field]: value } as User;
    });
    // Track de cambios sin guardar
    this.hasUnsavedChanges.set(true);
  }

  // Helper para obtener los datos del entity como Record
  getDetailEntityData(): Record<string, unknown> {
    const entity = this.detailEntity();
    if (!entity) return {};
    return entity as unknown as Record<string, unknown>;
  }

  onRelatedItemClick(event: unknown) {
    console.log('Related item clicked:', event);
    // TODO: implementar acciones según el tipo de item
  }

  private onSearchInternal(term: string) {
    clearTimeout(this.searchTimeout);
    this.searchTimeout = setTimeout(() => {
      const currentQuery = this.queryService.currentQuery();
      if (currentQuery) {
        const newQuery = {
          ...currentQuery,
          filters: term ? [{ field: 2 as const, operator: 3 as const, value: term }] : [],
        };
        this.executeQuery(newQuery);
      }
    }, 300);
  }
}
