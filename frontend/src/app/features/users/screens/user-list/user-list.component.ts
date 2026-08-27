import { Component, inject, OnInit, signal, computed, effect } from '@angular/core';
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
import { FormRuntimeComponent } from '../../../form-builder/runtime/form-runtime.component';
import { FormRuntimeService } from '../../../form-builder/services/runtime.service';
import { LayoutDefinition } from '../../../form-builder/models/layout-definition.model';
import {
  EntityDetailComponent,
  EntityTab,
} from '@shared/components/entity-detail/entity-detail.component';

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
    RelatedInfoComponent,
    FormRuntimeComponent,
  ],
  templateUrl: './user-list.component.html',
})
export class UserListComponent implements OnInit {
  readonly userService = inject(UserService);
  readonly queryService = inject(QueryService);
  readonly translate = inject(TranslationService);
  readonly uiStore = inject(UiStore);
  readonly router = inject(Router);
  readonly formRuntimeService = inject(FormRuntimeService);
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

  // Dynamic form layout for the detail drawer
  layout = signal<LayoutDefinition | null>(null);
  layoutLoading = signal<boolean>(false);
  layoutError = signal<string | null>(null);
  formValues = signal<Record<string, unknown>>({});
  formInitialValues = signal<Record<string, unknown>>({});

  private updateFormInitialValues(): void {
    const entity = this.detailEntity();
    this.formInitialValues.set(entity ? { ...entity } : {});
  }

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

  constructor() {
    // Load dynamic form layout when the detail drawer opens
    effect(() => {
      const entity = this.detailEntity();
      if (this.showDetail() && entity) {
        this.updateFormInitialValues();
        this.loadLayout();
      }
    });
  }

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
        const columns = config.columns ?? [];
        this.gridColumns.set(columns);

        // Construir columnas dinámicamente desde el backend
        if (columns.length > 0) {
          this.buildColumnsFromBackend(columns);
        } else {
          console.warn('[UserList] No columns from backend config, using fallback');
          this.buildColumns();
        }

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
          const rows = (response.data as Record<string, unknown>[]).map((row) => this.mapRowToUser(row));
          if (page === 1) {
            this.userService.users.set(rows);
          } else {
            this.userService.users.update((current) => [...current, ...rows]);
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
          const rows = (response.data as Record<string, unknown>[]).map((row) => this.mapRowToUser(row));
          if (page === 1) {
            this.userService.users.set(rows);
          } else {
            this.userService.users.update((current) => [...current, ...rows]);
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
    // Select the clicked row (deselect is handled by the dedicated button)
    this.selected.set(user);
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
      this.onFormSubmit(this.formValues());
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

    const user = this.mapRowToUser(selected as unknown as Record<string, unknown>);

    // Crear una copia del usuario (nuevo ID, limpio el id actual)
    const duplicated: User = {
      ...user,
      id: '', // ID vacío para crear nuevo
      name: `${user.name} (copia)`,
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
      // Mantener selección del grid sincronizada con el drawer
      this.selected.set(prevUser);
      if (this.showDetail()) {
        this.detailEntity.set(prevUser);
        this.hasUnsavedChanges.set(false);
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
      // Mantener selección del grid sincronizada con el drawer
      this.selected.set(nextUser);
      if (this.showDetail()) {
        this.detailEntity.set(nextUser);
        this.hasUnsavedChanges.set(false);
      }
    }
  }

  // === Dynamic Form Layout ===

  private loadLayout(): void {
    this.layoutLoading.set(true);
    this.layoutError.set(null);
    this.formRuntimeService.resolveForm('user-form').subscribe({
      next: (layout) => {
        this.layout.set(layout);
        this.layoutLoading.set(false);
      },
      error: (err) => {
        this.layout.set(null);
        this.layoutError.set(err?.message ?? 'Failed to load form layout');
        this.layoutLoading.set(false);
      },
    });
  }

  onFormSubmit(data: Record<string, unknown>): void {
    const entity = this.detailEntity();
    if (!entity) return;

    // Build user DTO from form data, only including known user keys
    const userKeys = ['name', 'email', 'status'] as const;
    const dto: Record<string, unknown> = {};
    for (const key of userKeys) {
      if (key in data) {
        dto[key] = data[key];
      }
    }

    // Preserve id and readonly fields from original entity
    const user: User = {
      ...entity,
      ...dto,
    } as User;

    this.onDetailSave(user);
  }

  onFormValueChange(values: Record<string, unknown>): void {
    this.formValues.set(values);
    this.hasUnsavedChanges.set(true);
  }

  // === Entity Detail Drawer ===

  /**
   * Normalize a grid row to the User model. The backend now returns domain keys
   * (id, name, email, status), but this helper still defends against raw DB column
   * names when a fallback is needed.
   */
  private mapRowToUser(row: Record<string, unknown>): User {
    return {
      ...row,
      id: String(row['id'] ?? row['usr_id'] ?? ''),
      name: String(row['name'] ?? row['usr_name'] ?? ''),
      email: String(row['email'] ?? row['usr_email'] ?? ''),
      status: (row['status'] ?? row['usr_status'] ?? 'active') as User['status'],
    } as User;
  }

  onRowDoubleClick(user: User) {
    // Abrir el drawer de detalle con doble click y marcar la fila seleccionada
    const mapped = this.mapRowToUser(user as unknown as Record<string, unknown>);
    this.selected.set(mapped);
    this.detailEntity.set(mapped);
    this.showDetail.set(true);
    this.hasUnsavedChanges.set(false);
  }

  onDetailClose() {
    this.showDetail.set(false);
    this.detailEntity.set(null);
    this.hasUnsavedChanges.set(false);
    this.layout.set(null);
    this.formValues.set({});
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
      this.layout.set(null);
      this.formValues.set({});
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
