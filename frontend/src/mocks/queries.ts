import { GridId, OPERATORS, SORT_DIRECTION, getFieldById, operatorIdToPrisma, sortDirectionToPrisma, GRID_FIELDS } from '../constants/grids';
import { SavedQuery, GridQuery, QueryFilter } from '../models/query.model';
import { User } from './handlers';

// Base de datos de queries (en memoria para mock)
let savedQueries: SavedQuery[] = [
  {
    id: '1',
    gridId: 1,
    name: 'Todos los usuarios',
    userId: null,
    isPublic: true,
    isDefault: true,
    query: {
      fields: [1, 2, 3, 4, 5],
      sort: [{ field: 2, direction: SORT_DIRECTION.ASC }],
      filters: [],
      pagination: { pageSize: 20 },
    },
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    gridId: 1,
    name: 'Solo activos',
    userId: 'user-1',
    isPublic: false,
    isDefault: false,
    query: {
      fields: [1, 2, 3, 4],
      sort: [{ field: 2, direction: SORT_DIRECTION.ASC }],
      filters: [{ field: 4, operator: OPERATORS.EQ, value: 'active' }],
      pagination: { pageSize: 20 },
    },
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
  },
  {
    id: '3',
    gridId: 1,
    name: 'Buscar por nombre',
    userId: 'user-1',
    isPublic: false,
    isDefault: false,
    query: {
      fields: [1, 2, 3, 4, 5],
      sort: [{ field: 2, direction: SORT_DIRECTION.ASC }],
      filters: [{ field: 2, operator: OPERATORS.CONTAINS, value: '' }],
      pagination: { pageSize: 50 },
    },
    createdAt: '2024-01-03T00:00:00Z',
    updatedAt: '2024-01-03T00:00:00Z',
  },
];

// Función para construir where clause de Prisma
function buildPrismaWhere(gridId: GridId, filters: QueryFilter[]): Record<string, unknown> {
  const where: Record<string, unknown> = {};
  
  for (const filter of filters) {
    const fieldDef = getFieldById(gridId, filter.field);
    if (!fieldDef) continue;
    
    const fieldKey = fieldDef.key;
    const prismaOp = operatorIdToPrisma(filter.operator);
    
    // Casos especiales
    if (filter.operator === OPERATORS.IS_NULL) {
      where[fieldKey] = { isNull: true };
    } else if (filter.operator === OPERATORS.IS_NOT_NULL) {
      where[fieldKey] = { isNotNull: true };
    } else if (filter.operator === OPERATORS.IN) {
      where[fieldKey] = { in: filter.value as (string | number)[] };
    } else {
      where[fieldKey] = { [prismaOp]: filter.value };
    }
  }
  
  return where;
}

// Función para construir orderBy de Prisma
function buildPrismaOrderBy(gridId: GridId, sort: { field: number; direction: number }[]): Record<string, 'asc' | 'desc'>[] {
  return sort.map(s => {
    const fieldDef = getFieldById(gridId, s.field);
    if (!fieldDef) return {};
    return { [fieldDef.key]: sortDirectionToPrisma(s.direction as 1 | 2) };
  }).filter(o => Object.keys(o).length > 0);
}

// Función para ejecutar query en datos
function executeGridQuery(gridId: GridId, query: GridQuery, page: number = 1): { data: User[]; meta: { page: number; pageSize: number; total: number; totalPages: number } } {
  let results = [...mockUsers];
  
  // Aplicar filtros
  if (query.filters.length > 0) {
    const where = buildPrismaWhere(gridId, query.filters);
    
    // Filtrar manualmente (simulando Prisma)
    results = results.filter(row => {
      for (const [key, condition] of Object.entries(where)) {
        const value = row[key as keyof User];
        
        if (condition && typeof condition === 'object') {
          if ('equals' in condition && value !== condition.equals) return false;
          if ('notEquals' in condition && value === condition.notEquals) return false;
          if ('contains' in condition && !String(value).toLowerCase().includes(String(condition.contains).toLowerCase())) return false;
          if ('gt' in condition && Number(value) <= Number(condition.gt)) return false;
          if ('lt' in condition && Number(value) >= Number(condition.lt)) return false;
          if ('gte' in condition && Number(value) < Number(condition.gte)) return false;
          if ('lte' in condition && Number(value) > Number(condition.lte)) return false;
          if ('in' in condition && !(condition.in as (string | number)[]).includes(value)) return false;
          if ('isNull' in condition && value != null) return false;
          if ('isNotNull' in condition && value == null) return false;
        }
      }
      return true;
    });
  }
  
  // Aplicar sorting
  const orderBy = buildPrismaOrderBy(gridId, query.sort);
  if (orderBy.length > 0) {
    results.sort((a, b) => {
      for (const ob of orderBy) {
        const key = Object.keys(ob)[0];
        const dir = ob[key];
        const aVal = a[key as keyof User];
        const bVal = b[key as keyof User];
        
        let cmp = 0;
        if (aVal < bVal) cmp = -1;
        if (aVal > bVal) cmp = 1;
        
        if (cmp !== 0) return dir === 'asc' ? cmp : -cmp;
      }
      return 0;
    });
  }
  
  // Aplicar paginación
  const total = results.length;
  const pageSize = query.pagination.pageSize;
  const totalPages = Math.ceil(total / pageSize);
  const start = (page - 1) * pageSize;
  const data = results.slice(start, start + pageSize);
  
  return {
    data,
    meta: { page, pageSize, total, totalPages },
  };
}

// Exportar funciones para usar en handlers
export { executeGridQuery, savedQueries };
