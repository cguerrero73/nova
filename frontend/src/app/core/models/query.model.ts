import { OperatorId, SortDirection } from '../constants/grids';

// Query filter enviado desde frontend
export interface QueryFilter {
  field: number; // ID del campo
  operator: OperatorId;
  value: unknown;
}

// Query sort enviado desde frontend
export interface QuerySort {
  field: number; // ID del campo
  direction: SortDirection;
}

// Query completo enviado desde frontend
export interface GridQuery {
  fields: number[]; // IDs de campos a mostrar
  sort: QuerySort[];
  filters: QueryFilter[];
  pagination: {
    pageSize: number;
  };
}

// Query guardada en backend
export interface SavedQuery {
  id: string;
  gridId: number;
  gridName: string;
  name: string;
  userId: string | null;
  isPublic: boolean;
  isDefault: boolean;
  query: GridQuery;
  createdAt: string;
  updatedAt: string;
}

// Request para crear/actualizar query
export interface SaveQueryRequest {
  gridId: number;
  name: string;
  isPublic: boolean;
  isDefault: boolean;
  query: GridQuery;
}

// Respuesta paginada
export interface PaginatedResponse<T> {
  success: boolean;
  data: T[];
  meta: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}

// API Response genérico
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

// Grid column metadata del backend
export interface GridColumn {
  id: number;
  key: string;
  label: string;
  type: 'string' | 'number' | 'date' | 'boolean' | 'select';
  sortable?: boolean;
  filterable?: boolean;
}

// Grid config del backend (metadata completa)
export interface GridConfig {
  gridId: number;
  gridName: string;
  baseQuery: string;
  availableFilters: number[];
  availableSort: number[];
  availableDisplay: number[];
  columns: GridColumn[];
}

// API Response para grid config
export interface GridConfigResponse {
  success: boolean;
  config: GridConfig;
  data?: never; // No usar data para este endpoint
}
