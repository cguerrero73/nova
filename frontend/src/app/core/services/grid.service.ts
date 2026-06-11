import { Injectable, inject, signal, computed } from '@angular/core';
import { Observable, of } from 'rxjs';
import { map, catchError } from 'rxjs/operators';
import { QueryService } from './query.service';
import { UiStore } from '../stores/ui.store';
import {
  SavedQuery,
  GridQuery,
  GridColumn,
  PaginatedResponse,
} from '../models/query.model';

// Cache para config de grids por nombre
const configCache = new Map<string, GridColumn[]>();

@Injectable({ providedIn: 'root' })
export class GridService {
  private readonly queryService = inject(QueryService);
  private readonly uiStore = inject(UiStore);

  // === STATE SIGNALS ===

  // Grid ID actual (obtenido de la config)
  readonly currentGridId = signal<number | null>(null);

  // Config del grid (columnas, labels, etc.)
  readonly gridConfig = signal<GridColumn[]>([]);

  // Queries guardadas del grid actual
  readonly queries = signal<SavedQuery[]>([]);

  // Query ID actualmente seleccionado (default o elegido por usuario)
  readonly selectedQueryId = signal<string | null>(null);

  // Datos resultantes de la última ejecución
  readonly currentData = signal<unknown[]>([]);

  // Metadata de paginación
  readonly meta = signal<{ page: number; pageSize: number; total: number; totalPages: number }>({
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0,
  });

  // Loading state
  readonly loading = signal(false);

  // === COMPUTED ===

  // El query actualmente seleccionado (objeto completo)
  readonly selectedQuery = computed(() => {
    const id = this.selectedQueryId();
    if (!id) return null;
    return this.queries().find((q) => q.id === id) || null;
  });

  // El query por default del grid
  readonly defaultQuery = computed(() => {
    return this.queries().find((q) => q.isDefault) || null;
  });

  // === METHODS ===

  /**
   * Cargar configuración del grid por nombre
   * Guarda columns, labels, tipos, etc. y el gridId para usar en queries
   */
  loadConfig(gridName: string): Observable<GridColumn[]> {
    // Si está en cache, devolver directo
    if (configCache.has(gridName)) {
      const columns = configCache.get(gridName)!;
      this.gridConfig.set(columns);
      return of(columns);
    }

    this.uiStore.setLoading(true);

    return this.queryService.getConfig(gridName).pipe(
      map((config) => {
        if (!config) {
          this.uiStore.setLoading(false);
          return [];
        }

        // Guardar gridId para usar en queries
        this.currentGridId.set(config.gridId);

        // Guardar columns en cache y signal
        configCache.set(gridName, config.columns);
        this.gridConfig.set(config.columns);

        this.uiStore.setLoading(false);
        return config.columns;
      }),
      catchError((err) => {
        console.error('[GridService] Error loading config:', err);
        this.uiStore.setLoading(false);
        return of([]);
      })
    );
  }

  /**
   * Cargar queries del grid actual usando el gridId guardado
   * (debe llamarse después de loadConfig que guarda el currentGridId)
   */
  loadQueries(): Observable<SavedQuery[]> {
    const gridId = this.currentGridId();
    if (!gridId) {
      console.warn('[GridService] No gridId set, call loadConfig first');
      return of([]);
    }

    this.uiStore.setLoading(true);

    return this.queryService.loadByGridId(gridId).pipe(
      map((queries) => {
        this.queries.set(queries);

        // Seleccionar el default automáticamente
        const defaultQuery = queries.find((q) => q.isDefault);
        if (defaultQuery) {
          this.selectedQueryId.set(defaultQuery.id);
        } else {
          this.selectedQueryId.set(null);
        }

        this.uiStore.setLoading(false);
        return queries;
      }),
      catchError((err) => {
        console.error('[GridService] Error loading queries:', err);
        this.uiStore.setLoading(false);
        return of([]);
      })
    );
  }

  /**
   * Seleccionar un query por ID
   */
  selectQuery(queryId: string): void {
    this.selectedQueryId.set(queryId);
  }

  /**
   * Ejecutar el query seleccionado actualmente
   */
  executeCurrentQuery(page: number = 1, pageSize: number = 20): void {
    const queryId = this.selectedQueryId();
    if (!queryId) {
      console.warn('[GridService] No query selected');
      return;
    }

    this.executeQuery(queryId, page, pageSize);
  }

  /**
   * Ejecutar un query específico por ID
   */
  executeQuery(queryId: string, page: number = 1, pageSize: number = 20): void {
    this.loading.set(true);

    this.queryService.executeQuery(queryId, page, pageSize).subscribe({
      next: (response) => {
        if (page === 1) {
          this.currentData.set(response.data);
        } else {
          this.currentData.update((current) => [...current, ...response.data]);
        }
        this.meta.set(response.meta);
        this.loading.set(false);
      },
      error: (err) => {
        console.error('[GridService] Error executing query:', err);
        this.loading.set(false);
      },
    });
  }

  /**
   * Invalidar cache de config (útil después de modificar columns)
   */
  invalidateConfigCache(): void {
    configCache.clear();
  }
}