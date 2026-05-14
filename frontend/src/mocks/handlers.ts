import { http, HttpResponse, delay } from 'msw';
import { GRID_IDS, OPERATORS, SORT_DIRECTION } from '../app/core/constants/grids';
import { SavedQuery, GridQuery, QueryFilter } from '../app/core/models/query.model';

interface User {
  id: string;
  name: string;
  email: string;
  status: 'active' | 'inactive' | 'suspended';
  createdAt: string;
  updatedAt: string;
}

// Mock data de usuarios
let mockUsers: User[] = [
  { id: '1', name: 'John Doe', email: 'john.doe@example.com', status: 'active', createdAt: '2024-01-15T10:00:00Z', updatedAt: '2024-01-15T10:00:00Z' },
  { id: '2', name: 'Jane Smith', email: 'jane.smith@example.com', status: 'active', createdAt: '2024-01-16T11:30:00Z', updatedAt: '2024-01-16T11:30:00Z' },
  { id: '3', name: 'Bob Johnson', email: 'bob.johnson@example.com', status: 'inactive', createdAt: '2024-01-17T09:15:00Z', updatedAt: '2024-01-17T09:15:00Z' },
  { id: '4', name: 'Alice Williams', email: 'alice.williams@example.com', status: 'active', createdAt: '2024-01-18T14:20:00Z', updatedAt: '2024-01-18T14:20:00Z' },
  { id: '5', name: 'Charlie Brown', email: 'charlie.brown@example.com', status: 'active', createdAt: '2024-01-19T08:45:00Z', updatedAt: '2024-01-19T08:45:00Z' },
  { id: '6', name: 'Diana Prince', email: 'diana.prince@example.com', status: 'inactive', createdAt: '2024-01-20T16:00:00Z', updatedAt: '2024-01-20T16:00:00Z' },
  { id: '7', name: 'Ethan Hunt', email: 'ethan.hunt@example.com', status: 'active', createdAt: '2024-01-21T12:30:00Z', updatedAt: '2024-01-21T12:30:00Z' },
  { id: '8', name: 'Fiona Gallagher', email: 'fiona.gallagher@example.com', status: 'active', createdAt: '2024-01-22T10:15:00Z', updatedAt: '2024-01-22T10:15:00Z' },
  { id: '9', name: 'George Miller', email: 'george.miller@example.com', status: 'inactive', createdAt: '2024-01-23T15:45:00Z', updatedAt: '2024-01-23T15:45:00Z' },
  { id: '10', name: 'Hannah Montana', email: 'hannah.montana@example.com', status: 'active', createdAt: '2024-01-24T11:00:00Z', updatedAt: '2024-01-24T11:00:00Z' },
  { id: '11', name: 'Ivan Drago', email: 'ivan.drago@example.com', status: 'active', createdAt: '2024-01-25T13:20:00Z', updatedAt: '2024-01-25T13:20:00Z' },
  { id: '12', name: 'Julia Roberts', email: 'julia.roberts@example.com', status: 'active', createdAt: '2024-01-26T09:30:00Z', updatedAt: '2024-01-26T09:30:00Z' },
  { id: '13', name: 'Kevin Hart', email: 'kevin.hart@example.com', status: 'inactive', createdAt: '2024-01-27T14:45:00Z', updatedAt: '2024-01-27T14:45:00Z' },
  { id: '14', name: 'Laura Palmer', email: 'laura.palmer@example.com', status: 'active', createdAt: '2024-01-28T10:00:00Z', updatedAt: '2024-01-28T10:00:00Z' },
  { id: '15', name: 'Mike Ross', email: 'mike.ross@example.com', status: 'active', createdAt: '2024-01-29T16:30:00Z', updatedAt: '2024-01-29T16:30:00Z' },
  { id: '16', name: 'Nancy Drew', email: 'nancy.drew@example.com', status: 'active', createdAt: '2024-01-30T11:15:00Z', updatedAt: '2024-01-30T11:15:00Z' },
  { id: '17', name: 'Oscar Isaac', email: 'oscar.isaac@example.com', status: 'inactive', createdAt: '2024-01-31T08:00:00Z', updatedAt: '2024-01-31T08:00:00Z' },
  { id: '18', name: 'Pamela Anderson', email: 'pamela.anderson@example.com', status: 'active', createdAt: '2024-02-01T12:00:00Z', updatedAt: '2024-02-01T12:00:00Z' },
  { id: '19', name: 'Quincy Jones', email: 'quincy.jones@example.com', status: 'active', createdAt: '2024-02-02T15:30:00Z', updatedAt: '2024-02-02T15:30:00Z' },
  { id: '20', name: 'Rachel Green', email: 'rachel.green@example.com', status: 'active', createdAt: '2024-02-03T09:45:00Z', updatedAt: '2024-02-03T09:45:00Z' },
  { id: '21', name: 'Steve Rogers', email: 'steve.rogers@example.com', status: 'active', createdAt: '2024-02-04T14:00:00Z', updatedAt: '2024-02-04T14:00:00Z' },
  { id: '22', name: 'Tina Turner', email: 'tina.turner@example.com', status: 'inactive', createdAt: '2024-02-05T10:30:00Z', updatedAt: '2024-02-05T10:30:00Z' },
  { id: '23', name: 'Uma Thurman', email: 'uma.thurman@example.com', status: 'active', createdAt: '2024-02-06T16:15:00Z', updatedAt: '2024-02-06T16:15:00Z' },
  { id: '24', name: 'Victor Frankenstein', email: 'victor.frankenstein@example.com', status: 'active', createdAt: '2024-02-07T11:45:00Z', updatedAt: '2024-02-07T11:45:00Z' },
  { id: '25', name: 'Wendy Darling', email: 'wendy.darling@example.com', status: 'active', createdAt: '2024-02-08T13:00:00Z', updatedAt: '2024-02-08T13:00:00Z' },
];

// Queries guardados
let savedQueries: SavedQuery[] = [
  {
    id: '1',
    gridId: GRID_IDS.USERS,
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
    gridId: GRID_IDS.USERS,
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
];

// ========== HELPERS ==========

function executeGridQuery(gridId: number, query: GridQuery, page: number = 1): { data: User[]; meta: { page: number; pageSize: number; total: number; totalPages: number } } {
  let results = [...mockUsers];
  
  // Map de operadores
  const operatorMap: Record<number, string> = {
    [OPERATORS.EQ]: 'equals',
    [OPERATORS.NE]: 'notEquals',
    [OPERATORS.CONTAINS]: 'contains',
    [OPERATORS.GT]: 'gt',
    [OPERATORS.LT]: 'lt',
    [OPERATORS.GTE]: 'gte',
    [OPERATORS.LTE]: 'lte',
    [OPERATORS.IN]: 'in',
  };

  // Map de campos por gridId
  const fieldMap: Record<number, Record<number, string>> = {
    [GRID_IDS.USERS]: {
      1: 'id',
      2: 'name',
      3: 'email',
      4: 'status',
      5: 'createdAt',
    },
  };
  
  const fields = fieldMap[gridId] || {};
  
  // Aplicar filtros
  if (query.filters.length > 0) {
    results = results.filter(row => {
      for (const filter of query.filters) {
        const fieldKey = fields[filter.field];
        if (!fieldKey) continue;
        
        const value = row[fieldKey as keyof User];
        const op = operatorMap[filter.operator];
        
        switch (filter.operator) {
          case OPERATORS.EQ:
            if (value !== filter.value) return false;
            break;
          case OPERATORS.NE:
            if (value === filter.value) return false;
            break;
          case OPERATORS.CONTAINS:
            if (!String(value).toLowerCase().includes(String(filter.value).toLowerCase())) return false;
            break;
          case OPERATORS.GT:
            if (Number(value) <= Number(filter.value)) return false;
            break;
          case OPERATORS.LT:
            if (Number(value) >= Number(filter.value)) return false;
            break;
          case OPERATORS.GTE:
            if (Number(value) < Number(filter.value)) return false;
            break;
          case OPERATORS.LTE:
            if (Number(value) > Number(filter.value)) return false;
            break;
          case OPERATORS.IN:
            if (!(filter.value as string[]).includes(String(value))) return false;
            break;
        }
      }
      return true;
    });
  }
  
  // Aplicar sorting
  if (query.sort.length > 0) {
    results.sort((a, b) => {
      for (const s of query.sort) {
        const fieldKey = fields[s.field];
        if (!fieldKey) continue;
        
        const aVal = a[fieldKey as keyof User];
        const bVal = b[fieldKey as keyof User];
        const dir = s.direction === SORT_DIRECTION.ASC ? 1 : -1;
        
        if (aVal < bVal) return -1 * dir;
        if (aVal > bVal) return 1 * dir;
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

// ========== HANDLERS ==========

export const handlers = [
  // GET /api/v1/users - método original (para compatibilidad)
  http.get('/api/v1/users', async ({ request }) => {
    await delay(500);
    const url = new URL(request.url);
    const page = Number(url.searchParams.get('page')) || 1;
    const pageSize = Number(url.searchParams.get('pageSize')) || 20;
    const search = url.searchParams.get('search') || '';

    let users = [...mockUsers];
    if (search) {
      const searchLower = search.toLowerCase();
      users = users.filter(u => 
        u.name.toLowerCase().includes(searchLower) ||
        u.email.toLowerCase().includes(searchLower)
      );
    }

    const total = users.length;
    const totalPages = Math.ceil(total / pageSize);
    const start = (page - 1) * pageSize;
    const data = users.slice(start, start + pageSize);

    return HttpResponse.json({
      success: true,
      data,
      meta: { page, pageSize, total, totalPages }
    });
  }),

  // POST /api/v1/users - crear nuevo usuario
  http.post('/api/v1/users', async ({ request }) => {
    await delay(300);
    const body = await request.json() as { name: string; email: string; status?: 'active' | 'inactive' | 'suspended' };
    
    const newUser: User = {
      id: String(Date.now()),
      name: body.name,
      email: body.email,
      status: body.status || 'active',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
    
    mockUsers.push(newUser);
    
    return HttpResponse.json({
      success: true,
      data: newUser
    });
  }),

  // PUT /api/v1/users/:id - actualizar usuario
  http.put('/api/v1/users/:id', async ({ params, request }) => {
    await delay(300);
    const body = await request.json() as { name?: string; email?: string; status?: 'active' | 'inactive' | 'suspended' };
    const userId = params['id'];
    
    const index = mockUsers.findIndex(u => u.id === userId);
    if (index === -1) {
      return HttpResponse.json(
        { success: false, error: 'User not found' },
        { status: 404 }
      );
    }
    
    const updatedUser: User = {
      ...mockUsers[index],
      ...body,
      updatedAt: new Date().toISOString()
    };
    mockUsers[index] = updatedUser;
    
    return HttpResponse.json({
      success: true,
      data: updatedUser
    });
  }),

  // DELETE /api/v1/users/:id - eliminar usuario
  http.delete('/api/v1/users/:id', async ({ params }) => {
    await delay(300);
    const userId = params['id'];
    
    const index = mockUsers.findIndex(u => u.id === userId);
    if (index === -1) {
      return HttpResponse.json(
        { success: false, error: 'User not found' },
        { status: 404 }
      );
    }
    
    mockUsers.splice(index, 1);
    
    return HttpResponse.json({
      success: true,
      data: null
    });
  }),

  // POST /api/v1/grid/data - ejecutar query en el grid
  http.post('/api/v1/grid/data', async ({ request }) => {
    await delay(300);
    
    const body = await request.json() as {
      gridId: number;
      query: GridQuery;
      page?: number;
    };
    
    const { gridId, query, page = 1 } = body;
    
    const result = executeGridQuery(gridId, query, page);
    
    return HttpResponse.json({
      success: true,
      data: result.data,
      meta: result.meta,
    });
  }),

  // GET /api/v1/queries - listar queries de un grid
  http.get('/api/v1/queries', async ({ request }) => {
    await delay(200);
    
    const url = new URL(request.url);
    const gridId = Number(url.searchParams.get('gridId'));
    const userId = url.searchParams.get('userId') || 'anonymous';
    
    if (!gridId) {
      return HttpResponse.json(
        { success: false, error: 'gridId es requerido' },
        { status: 400 }
      );
    }
    
    // Filtrar queries: propias + públicas del grid
    const queries = savedQueries.filter(q => 
      q.gridId === gridId && 
      (q.isPublic || q.userId === userId)
    );
    
    return HttpResponse.json({
      success: true,
      data: queries,
    });
  }),

  // GET /api/v1/queries/:id - obtener query específico
  http.get('/api/v1/queries/:id', async ({ params }) => {
    await delay(200);
    
    const query = savedQueries.find(q => q['id'] === params['id']);
    
    if (!query) {
      return HttpResponse.json(
        { success: false, error: 'Query no encontrado' },
        { status: 404 }
      );
    }
    
    return HttpResponse.json({
      success: true,
      data: query,
    });
  }),

  // POST /api/v1/queries - crear query
  http.post('/api/v1/queries', async ({ request }) => {
    await delay(300);
    
    const body = await request.json() as SavedQuery;
    const userId = 'user-1'; // Simulado
    
    // Si es default, quitar default de otros
    if (body.isDefault) {
      savedQueries = savedQueries.map(q => 
        q.gridId === body.gridId ? { ...q, isDefault: false } : q
      );
    }
    
    const newQuery: SavedQuery = {
      id: String(Date.now()),
      gridId: body.gridId,
      name: body.name,
      userId,
      isPublic: body.isPublic,
      isDefault: body.isDefault,
      query: body.query,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    
    savedQueries.push(newQuery);
    
    return HttpResponse.json({
      success: true,
      data: newQuery,
    });
  }),

  // PUT /api/v1/queries/:id - actualizar query
  http.put('/api/v1/queries/:id', async ({ params, request }) => {
    await delay(300);
    
    const body = await request.json() as Partial<SavedQuery>;
    const index = savedQueries.findIndex(q => q['id'] === params['id']);
    
    if (index === -1) {
      return HttpResponse.json(
        { success: false, error: 'Query no encontrado' },
        { status: 404 }
      );
    }
    
    // Si es default, quitar default de otros
    if (body.isDefault) {
      savedQueries = savedQueries.map(q => 
        q.gridId === savedQueries[index].gridId ? { ...q, isDefault: false } : q
      );
    }
    
    savedQueries[index] = {
      ...savedQueries[index],
      ...body,
      updatedAt: new Date().toISOString(),
    };
    
    return HttpResponse.json({
      success: true,
      data: savedQueries[index],
    });
  }),

  // DELETE /api/v1/queries/:id - eliminar query
  http.delete('/api/v1/queries/:id', async ({ params }) => {
    await delay(200);
    
    const index = savedQueries.findIndex(q => q['id'] === params['id']);
    
    if (index === -1) {
      return HttpResponse.json(
        { success: false, error: 'Query no encontrado' },
        { status: 404 }
      );
    }
    
    savedQueries.splice(index, 1);
    
    return HttpResponse.json({
      success: true,
      data: null,
    });
  }),

  // ========== AUTH HANDLERS ==========

  // POST /api/v1/auth/login
  http.post('/api/v1/auth/login', async ({ request }) => {
    await delay(500);
    
    const body = await request.json() as { email: string; password: string };
    const { email, password } = body;

    // Mock authentication - accept any user with password "password123"
    if (password === 'password123') {
      const mockUser = {
        id: 'user-1',
        email,
        name: email.split('@')[0],
        roles: ['user'],
      };

      return HttpResponse.json({
        success: true,
        data: {
          user: mockUser,
          accessToken: 'mock_access_token_' + Date.now(),
          refreshToken: 'mock_refresh_token_' + Date.now(),
          expiresIn: 3600,
        },
      });
    }

    return HttpResponse.json(
      { success: false, error: 'Credenciales inválidas' },
      { status: 401 }
    );
  }),

  // POST /api/v1/auth/register
  http.post('/api/v1/auth/register', async ({ request }) => {
    await delay(500);
    
    const body = await request.json() as { email: string; password: string; name: string };
    const { email, name } = body;

    // Mock registration
    const mockUser = {
      id: 'user-' + Date.now(),
      email,
      name,
      roles: ['user'],
    };

    return HttpResponse.json({
      success: true,
      data: {
        user: mockUser,
        accessToken: 'mock_access_token_' + Date.now(),
        refreshToken: 'mock_refresh_token_' + Date.now(),
        expiresIn: 3600,
      },
    });
  }),

  // POST /api/v1/auth/refresh
  http.post('/api/v1/auth/refresh', async ({ request }) => {
    await delay(200);
    
    const body = await request.json() as { refreshToken: string };
    
    if (body.refreshToken) {
      return HttpResponse.json({
        success: true,
        data: {
          accessToken: 'mock_access_token_' + Date.now(),
          expiresIn: 3600,
        },
      });
    }

    return HttpResponse.json(
      { success: false, error: 'Refresh token inválido' },
      { status: 401 }
    );
  }),

  // POST /api/v1/auth/logout
  http.post('/api/v1/auth/logout', async () => {
    await delay(100);
    return HttpResponse.json({ success: true, data: null });
  }),

  // GET /api/v1/auth/config
  http.get('/api/v1/auth/config', async () => {
    await delay(200);
    return HttpResponse.json({
      success: true,
      data: {
        id: 'auth-config-1',
        method: 'local',
        isActive: true,
        local: {
          requireEmailVerification: false,
          minPasswordLength: 8,
          maxLoginAttempts: 5,
          lockoutDuration: 15,
        },
      },
    });
  }),

  // ========== SCREENS / TRANSLATIONS HANDLERS ==========

  // GET /api/v1/screens/:screenId - obtener metadata + traducciones de una pantalla
  http.get('/api/v1/screens/:screenId', async ({ params, request }) => {
    await delay(300);
    
    const url = new URL(request.url);
    const lang = url.searchParams.get('lang') || 'en';
    const screenId = params['screenId'];

    // Traducciones por idioma y pantalla
    const translations: Record<string, Record<string, Record<string, string>>> = {
      login: {
        en: {
          'title': 'Nova',
          'subtitle': 'Sign in to your account',
          'form.name.label': 'Full name',
          'form.email.label': 'Email address',
          'form.password.label': 'Password',
          'form.remember.label': 'Remember me',
          'button.submit.login': 'Sign in',
          'button.submit.register': 'Create account',
          'button.toggle.login': "Don't have an account?",
          'button.toggle.register': 'Already have an account?',
          'link.register': 'Sign up',
          'link.login': 'Sign in',
          'error.credentials': 'Invalid credentials',
          'error.connection': 'Connection error',
        },
        es: {
          'title': 'Nova',
          'subtitle': 'Iniciar sesión',
          'form.name.label': 'Nombre completo',
          'form.email.label': 'Correo electrónico',
          'form.password.label': 'Contraseña',
          'form.remember.label': 'Recordarme',
          'button.submit.login': 'Iniciar sesión',
          'button.submit.register': 'Crear cuenta',
          'button.toggle.login': '¿No tienes cuenta?',
          'button.toggle.register': '¿Ya tienes cuenta?',
          'link.register': 'Regístrate',
          'link.login': 'Iniciar sesión',
          'error.credentials': 'Credenciales inválidas',
          'error.connection': 'Error de conexión',
        },
        pt: {
          'title': 'Nova',
          'subtitle': 'Entrar na conta',
          'form.name.label': 'Nome completo',
          'form.email.label': 'Endereço de email',
          'form.password.label': 'Senha',
          'form.remember.label': 'Lembrar-me',
          'button.submit.login': 'Entrar',
          'button.submit.register': 'Criar conta',
          'button.toggle.login': 'Não tem uma conta?',
          'button.toggle.register': 'Já tem uma conta?',
          'link.register': 'Cadastre-se',
          'link.login': 'Entrar',
          'error.credentials': 'Credenciais inválidas',
          'error.connection': 'Erro de conexão',
        },
        fr: {
          'title': 'Nova',
          'subtitle': 'Se connecter',
          'form.name.label': 'Nom complet',
          'form.email.label': 'Adresse email',
          'form.password.label': 'Mot de passe',
          'form.remember.label': 'Se souvenir de moi',
          'button.submit.login': 'Se connecter',
          'button.submit.register': 'Créer un compte',
          'button.toggle.login': 'Pas de compte?',
          'button.toggle.register': 'Déjà un compte?',
          'link.register': "S'inscrire",
          'link.login': 'Se connecter',
          'error.credentials': 'Identifiants invalides',
          'error.connection': 'Erreur de connexion',
        },
        it: {
          'title': 'Nova',
          'subtitle': 'Accedi al tuo account',
          'form.name.label': 'Nome completo',
          'form.email.label': 'Indirizzo email',
          'form.password.label': 'Password',
          'form.remember.label': 'Ricordami',
          'button.submit.login': 'Accedi',
          'button.submit.register': 'Crea account',
          'button.toggle.login': 'Non hai un account?',
          'button.toggle.register': 'Hai già un account?',
          'link.register': 'Registrati',
          'link.login': 'Accedi',
          'error.credentials': 'Credenziali invalide',
          'error.connection': 'Errore di connessione',
        },
      },
      users: {
        en: {
          'title': 'Users',
          'toolbar.save': 'Save',
          'toolbar.create': 'Create',
          'toolbar.delete': 'Delete',
          'toolbar.prev': 'Prev',
          'toolbar.next': 'Next',
          'toolbar.duplicate': 'Duplicate',
          'toolbar.refresh': 'Refresh',
          'toolbar.print': 'Print',
          'form.id.label': 'ID',
          'form.name.label': 'Name',
          'form.name.placeholder': 'Full name',
          'form.email.label': 'Email',
          'form.email.placeholder': 'email@example.com',
          'form.status.label': 'Status',
          'form.createdAt.label': 'Created Date',
          'status.active': 'Active',
          'status.inactive': 'Inactive',
          'status.suspended': 'Suspended',
          'messages.save_success': 'Saved successfully',
          'messages.delete_confirm': 'Are you sure you want to delete?',
          'messages.delete_success': 'Deleted successfully',
          'grid.empty': 'No users found',
          'grid.selected': 'Selected user',
          'grid.deselect': 'Deselect',
          'query.select': 'Select query...',
          'tabs.view': 'Record View',
          'tabs.comments': 'Comments',
          'tabs.documents': 'Documents',
          'tabs.audit': 'History',
          'detail.title': 'User',
        },
        es: {
          'title': 'Usuarios',
          'toolbar.save': 'Salvar',
          'toolbar.create': 'Crear',
          'toolbar.delete': 'Borrar',
          'toolbar.prev': 'Prev',
          'toolbar.next': 'Next',
          'toolbar.duplicate': 'Duplicar',
          'toolbar.refresh': 'Refrescar',
          'toolbar.print': 'Imprimir',
          'form.id.label': 'ID',
          'form.name.label': 'Nombre',
          'form.name.placeholder': 'Nombre completo',
          'form.email.label': 'Email',
          'form.email.placeholder': 'correo@ejemplo.com',
          'form.status.label': 'Estado',
          'form.createdAt.label': 'Fecha de Creación',
          'status.active': 'Activo',
          'status.inactive': 'Inactivo',
          'status.suspended': 'Suspendido',
          'messages.save_success': 'Grabado correctamente',
          'messages.delete_confirm': '¿Está seguro que desea eliminar?',
          'messages.delete_success': 'Eliminado correctamente',
          'grid.empty': 'No se encontraron usuarios',
          'grid.selected': 'Usuario seleccionado',
          'grid.deselect': 'Deseleccionar',
          'query.select': 'Seleccionar query...',
          'tabs.view': 'Visualización del Registro',
          'tabs.comments': 'Comentarios',
          'tabs.documents': 'Documentos',
          'tabs.audit': 'Historial',
          'detail.title': 'Usuario',
        },
        pt: {
          'title': 'Usuários',
          'toolbar.save': 'Salvar',
          'toolbar.create': 'Criar',
          'toolbar.delete': 'Excluir',
          'toolbar.prev': 'Anterior',
          'toolbar.next': 'Próximo',
          'toolbar.duplicate': 'Duplicar',
          'toolbar.refresh': 'Atualizar',
          'toolbar.print': 'Imprimir',
          'form.id.label': 'ID',
          'form.name.label': 'Nome',
          'form.email.label': 'Email',
          'form.status.label': 'Status',
          'form.createdAt.label': 'Data de Criação',
          'status.active': 'Ativo',
          'status.inactive': 'Inativo',
          'status.suspended': 'Suspenso',
          'messages.save_success': 'Salvo com sucesso',
          'messages.delete_confirm': 'Tem certeza que deseja excluir?',
          'messages.delete_success': 'Excluído com sucesso',
          'grid.empty': 'Nenhum usuário encontrado',
          'grid.selected': 'Usuário selecionado',
          'grid.deselect': 'Desselecionar',
          'query.select': 'Selecionar query...',
        },
        fr: {
          'title': 'Utilisateurs',
          'toolbar.save': 'Enregistrer',
          'toolbar.create': 'Créer',
          'toolbar.delete': 'Supprimer',
          'toolbar.prev': 'Précédent',
          'toolbar.next': 'Suivant',
          'toolbar.duplicate': 'Dupliquer',
          'toolbar.refresh': 'Actualiser',
          'toolbar.print': 'Imprimer',
          'form.id.label': 'ID',
          'form.name.label': 'Nom',
          'form.email.label': 'Email',
          'form.status.label': 'Statut',
          'form.createdAt.label': 'Date de création',
          'status.active': 'Actif',
          'status.inactive': 'Inactif',
          'status.suspended': 'Suspendu',
          'messages.save_success': 'Enregistré avec succès',
          'messages.delete_confirm': 'Êtes-vous sûr de vouloir supprimer?',
          'messages.delete_success': 'Supprimé avec succès',
          'grid.empty': 'Aucun utilisateur trouvé',
          'grid.selected': 'Utilisateur sélectionné',
          'grid.deselect': 'Désélectionner',
          'query.select': 'Sélectionner une requête...',
        },
        it: {
          'title': 'Utenti',
          'toolbar.save': 'Salva',
          'toolbar.create': 'Crea',
          'toolbar.delete': 'Elimina',
          'toolbar.prev': 'Precedente',
          'toolbar.next': 'Successivo',
          'toolbar.duplicate': 'Duplica',
          'toolbar.refresh': 'Aggiorna',
          'toolbar.print': 'Stampa',
          'form.id.label': 'ID',
          'form.name.label': 'Nome',
          'form.email.label': 'Email',
          'form.status.label': 'Stato',
          'form.createdAt.label': 'Data di creazione',
          'status.active': 'Attivo',
          'status.inactive': 'Inattivo',
          'status.suspended': 'Sospeso',
          'messages.save_success': 'Salvato con successo',
          'messages.delete_confirm': 'Sei sicuro di voler eliminare?',
          'messages.delete_success': 'Eliminato con successo',
          'grid.empty': 'Nessun utente trovato',
          'grid.selected': 'Utente selezionato',
          'grid.deselect': 'Deseleziona',
          'query.select': 'Seleziona query...',
        },
      },
    };

    const screenTranslations = translations[screenId as keyof typeof translations];
    if (!screenTranslations) {
      return HttpResponse.json(
        { success: false, error: 'Screen not found' },
        { status: 404 }
      );
    }

    const langTranslations = screenTranslations[lang as keyof typeof screenTranslations] || screenTranslations['en'];

    return HttpResponse.json({
      success: true,
      data: {
        screenId,
        translations: langTranslations,
      },
    });
  }),
];
