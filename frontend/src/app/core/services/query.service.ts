import { Injectable, inject, signal } from '@angular/core';
import { Observable, from } from 'rxjs';
import { map } from 'rxjs/operators';
import { ApiService } from '@core/services/api.service';
import { UiStore } from '@core/stores/ui.store';
import { SavedQuery, GridQuery, QueryFilter, QuerySort, PaginatedResponse } from '@core/models/query.model';
import { GridId, GRID_IDS, OPERATORS, SORT_DIRECTION, GridField, getGridFields } from '@core/constants/grids';

@Injectable({ providedIn: 'root' })
export class QueryService {
  private readonly api = inject(ApiService);
  private readonly uiStore = inject(UiStore);

  // State
  readonly queries = signal<SavedQuery[]>([]);
  readonly selectedQuery = signal<SavedQuery | null>(null);
  readonly currentQuery = signal<GridQuery | null>(null);
  readonly currentMeta = signal({ page: 1, pageSize: 20, total: 0, totalPages: 0 });

  // Ejecutar query y obtener datos
  executeQuery(gridId: number, query: GridQuery, page: number = 1): Observable<PaginatedResponse<unknown>> {
    return this.api.postRaw<PaginatedResponse<unknown>>('/grid/data', {
      gridId,
      query,
      page,
    }).pipe(
      map(response => {
        this.currentMeta.set(response.meta);
        return response;
      })
    );
  }

  // Cargar queries de un grid
  loadByGridId(gridId: number): Observable<SavedQuery[]> {
    this.uiStore.setLoading(true);
    
    return this.api.get<SavedQuery[]>('/queries', { gridId }).pipe(
      map(response => {
        const queries = response?.data || [];
        this.queries.set(queries);
        
        // Buscar default
        const defaultQuery = queries.find(q => q.isDefault);
        if (defaultQuery) {
          this.selectedQuery.set(defaultQuery);
          this.currentQuery.set(defaultQuery.query);
        }
        
        this.uiStore.setLoading(false);
        return queries;
      })
    );
  }

  // Obtener un query por ID
  getById(id: string): Observable<SavedQuery | null> {
    return this.api.getById<SavedQuery>('/queries', id).pipe(
      map(response => response?.data || null)
    );
  }

  // Guardar nuevo query
  save(query: Omit<SavedQuery, 'id' | 'createdAt' | 'updatedAt'>): Observable<SavedQuery> {
    return this.api.post<SavedQuery>('/queries', query).pipe(
      map(response => {
        const saved = response.data!;
        this.queries.update(list => [...list, saved]);
        return saved;
      })
    );
  }

  // Actualizar query
  update(id: string, query: Partial<SavedQuery>): Observable<SavedQuery> {
    return this.api.put<SavedQuery>('/queries', id, query).pipe(
      map(response => {
        const updated = response.data!;
        this.queries.update(list => 
          list.map(q => q.id === id ? updated : q)
        );
        return updated;
      })
    );
  }

  // Eliminar query
  delete(id: string): Observable<boolean> {
    return this.api.delete('/queries', id).pipe(
      map(response => {
        this.queries.update(list => list.filter(q => q.id !== id));
        
        // Si era el seleccionado, limpiar
        if (this.selectedQuery()?.id === id) {
          this.selectedQuery.set(null);
          this.currentQuery.set(null);
        }
        
        return response.success;
      })
    );
  }

  // Seleccionar un query y devolver la query para ejecutar
  selectQuery(query: SavedQuery): void {
    this.selectedQuery.set(query);
    this.currentQuery.set(query.query);
  }

  // Limpiar selección
  clearSelection(): void {
    this.selectedQuery.set(null);
    this.currentQuery.set(null);
  }

  // Obtener fields de un grid
  getFields(gridId: GridId): GridField[] {
    return getGridFields(gridId);
  }

  // Crear query vacía por defecto
  createDefaultQuery(gridId: number): GridQuery {
    const fields = getGridFields(gridId as GridId);
    return {
      fields: fields.map(f => f.id),
      sort: [],
      filters: [],
      pagination: { pageSize: 20 },
    };
  }
}
