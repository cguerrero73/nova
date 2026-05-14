// Grid IDs - cada grid tiene un ID numérico único
export const GRID_IDS = {
  USERS: 1,
} as const;

export type GridId = typeof GRID_IDS[keyof typeof GRID_IDS];

// Operadores disponibles para filtros
export const OPERATORS = {
  EQ: 1,         // =
  NE: 2,         // !=
  CONTAINS: 3,   // contains
  GT: 4,         // >
  LT: 5,         // <
  GTE: 6,        // >=
  LTE: 7,        // <=
  IN: 8,         // in
  IS_NULL: 9,   // is null
  IS_NOT_NULL: 10, // is not null
} as const;

export type OperatorId = typeof OPERATORS[keyof typeof OPERATORS];

// Dirección de ordenamiento
export const SORT_DIRECTION = {
  ASC: 1,
  DESC: 2,
} as const;

export type SortDirection = typeof SORT_DIRECTION[keyof typeof SORT_DIRECTION];

// Metadata de campos para cada grid
export interface GridField {
  id: number;
  key: string;
  label: string;
  type: 'string' | 'number' | 'date' | 'boolean' | 'select';
  sortable?: boolean;
  filterable?: boolean;
  options?: { value: string; label: string }[];
}

// Definición de campos por grid
export const GRID_FIELDS: Record<GridId, GridField[]> = {
  [GRID_IDS.USERS]: [
    { id: 1, key: 'id', label: 'ID', type: 'string', sortable: true, filterable: true },
    { id: 2, key: 'name', label: 'Nombre', type: 'string', sortable: true, filterable: true },
    { id: 3, key: 'email', label: 'Email', type: 'string', sortable: true, filterable: true },
    { id: 4, key: 'status', label: 'Estado', type: 'select', sortable: true, filterable: true, 
      options: [
        { value: 'active', label: 'Activo' },
        { value: 'inactive', label: 'Inactivo' },
      ]
    },
    { id: 5, key: 'createdAt', label: 'Fecha Creación', type: 'date', sortable: true, filterable: true },
  ],
};

// Helper para obtener campos de un grid
export function getGridFields(gridId: GridId): GridField[] {
  return GRID_FIELDS[gridId] || [];
}

// Helper para obtener un campo por ID
export function getFieldById(gridId: GridId, fieldId: number): GridField | undefined {
  const fields = GRID_FIELDS[gridId];
  return fields?.find(f => f.id === fieldId);
}

// Helper para convertir operador ID a string Prisma
export function operatorIdToPrisma(operatorId: OperatorId): string {
  const map: Record<OperatorId, string> = {
    [OPERATORS.EQ]: 'equals',
    [OPERATORS.NE]: 'notEquals',
    [OPERATORS.CONTAINS]: 'contains',
    [OPERATORS.GT]: 'gt',
    [OPERATORS.LT]: 'lt',
    [OPERATORS.GTE]: 'gte',
    [OPERATORS.LTE]: 'lte',
    [OPERATORS.IN]: 'in',
    [OPERATORS.IS_NULL]: 'isNull',
    [OPERATORS.IS_NOT_NULL]: 'isNotNull',
  };
  return map[operatorId] || 'equals';
}

// Helper para convertir dirección de sort
export function sortDirectionToPrisma(direction: SortDirection): 'asc' | 'desc' {
  return direction === SORT_DIRECTION.ASC ? 'asc' : 'desc';
}
