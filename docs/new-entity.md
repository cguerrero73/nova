# Guía: Crear una Nueva Entidad

Metodología paso a paso para agregar una entidad completa en Nova, desde la base de datos hasta la pantalla.

---

## Índice

1. [Orden de trabajo](#1-orden-de-trabajo)
2. [Backend: Domain Layer](#2-backend-domain-layer)
3. [Backend: Adapters](#3-backend-adapters)
4. [Backend: Wire y Routes](#4-backend-wire-y-routes)
5. [Frontend: Models y Service](#5-frontend-models-y-service)
6. [Frontend: Screens](#6-frontend-screens)
7. [Frontend: Routes](#7-frontend-routes)
8. [Checklist rápida](#8-checklist-rápida)

---

## 1. Orden de trabajo

```
Backend (7 archivos)              Frontend (4+ archivos)
─────────────────────             ──────────────────────
  1. domain/entity.go               8. models/*.model.ts
  2. domain/ports.go                9. services/*.service.ts
  3. domain/dto.go                 10. screens/*-list/
  4. domain/service.go             11. screens/*-detail/
  5. adapters/api/handler.go       12. app.routes.ts
  6. adapters/db/repository.go
  7. wire/wire.go + main.go
```

Cada paso tiene tests presenciales y compilación limpia antes de pasar al siguiente.

---

## 2. Backend: Domain Layer

### 2.1 `domain/<entidad>/entity.go`

La entidad de negocio. Sin tags de base de datos, solo JSON y lógica de dominio.

```go
package users

import "time"

type User struct {
	ID         string    `json:"usr_id"`
	Code       string    `json:"usr_code"`
	Name       string    `json:"usr_name"`
	Email      string    `json:"usr_email"`
	Password   string    `json:"-"`           // nunca se serializa
	Phone      string    `json:"usr_phone"`
	Status     string    `json:"usr_status"`
	DefaultOrg string    `json:"usr_default_org"`
	NotUsed    *string   `json:"usr_notused,omitempty"`
	TenantID   string    `json:"usr_tenant_id"`
	CreatedAt  time.Time `json:"usr_created_at"`
	UpdatedAt  time.Time `json:"usr_updated_at"`
	CreatedBy  string    `json:"usr_created_by,omitempty"`
	UpdatedBy  string    `json:"usr_updated_by,omitempty"`
}

// Métodos de negocio sobre el dominio
func (u *User) IsActive() bool {
	return u.Status == "ACT" && (u.NotUsed == nil || *u.NotUsed != "+")
}
```

Reglas:

- Un archivo por entidad
- Métodos de negocio _sobre_ la entidad (receiver)
- Sin dependencias externas

### 2.2 `domain/<entidad>/ports.go`

Define el **contrato** que los adapters deben cumplir. El domain no sabe quién implementa esto.

```go
package users

import "context"

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByCode(ctx context.Context, code string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindAll(ctx context.Context, tenantID string, limit, offset int) ([]*User, int, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}
```

Reglas:

- Cada método recibe `context.Context` como primer parámetro
- Usar tipos del domain (`*User`), nunca tipos de adapter
- `FindAll` devuelve `([]*T, int, error)` — el `int` es el total para paginación

### 2.3 `domain/<entidad>/dto.go`

DTOs de entrada/salida. Separados de la entidad para no exponer campos internos.

```go
package users

type CreateUserRequest struct {
	Code       string `json:"usr_code"`
	Name       string `json:"usr_name" validate:"required"`
	Email      string `json:"usr_email" validate:"required,email"`
	Password   string `json:"usr_password" validate:"required,min=8"`
	Phone      string `json:"usr_phone"`
	DefaultOrg string `json:"usr_default_org"`
}

type UpdateUserRequest struct {
	Name       string `json:"usr_name"`
	Email      string `json:"usr_email" validate:"email"`
	Phone      string `json:"usr_phone"`
	Status     string `json:"usr_status"`
	DefaultOrg string `json:"usr_default_org"`
}
```

Reglas:

- No reutilizar la entidad como request — siempre un DTO separado
- Tags `validate` para validación
- Campos opcionales son zero-value por defecto

### 2.4 `domain/<entidad>/service.go`

La lógica de negocio. Solo depende de la interfaz del repositorio.

```go
package users

import (
	"context"
	"time"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/nova/backend/pkg/errors"
)

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) FindByID(ctx context.Context, id string) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) FindAll(ctx context.Context, tenantID string, limit, offset int) ([]*User, int, error) {
	return s.repo.FindAll(ctx, tenantID, limit, offset)
}

func (s *UserService) Create(ctx context.Context, tenantID string, req *CreateUserRequest) (*User, error) {
	// 1. Validar negocio
	existing, _ := s.repo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.ErrUserExists()
	}

	// 2. Procesar
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := &User{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashed),
		Status:    "ACT",
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 3. Persistir vía puerto
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, id string, req *UpdateUserRequest) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.ErrUserNotFound()
	}

	// Merge parcial
	if req.Name != "" { user.Name = req.Name }
	if req.Email != "" { user.Email = req.Email }
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil { return err }
	if user == nil { return errors.ErrUserNotFound() }
	return s.repo.Delete(ctx, id)
}
```

Patrón de cada método:

1. **Validar** — reglas de negocio, existencia, permisos
2. **Operar** — transformar, calcular, mergear
3. **Persistir** — llamar al repo
4. **Devolver** — resultado o error

---

## 3. Backend: Adapters

### 3.1 `adapters/api/<entidad>/handler.go`

Adapter de entrada: HTTP → Domain.

```go
package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nova/backend/internal/domain/users"
	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/pkg/errors"
)

type UserHandler struct {
	userService *users.UserService
}

func NewUserHandler(service *users.UserService) *UserHandler {
	return &UserHandler{userService: service}
}

func (h *UserHandler) List(c *fiber.Ctx) error {
	tenant := c.Locals("tenant").(string)
	pagination := dto.PaginationQuery{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 20),
	}
	users, total, err := h.userService.FindAll(c.Context(), tenant,
		pagination.GetLimit(), pagination.GetOffset())
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	return c.JSON(dto.NewListResponse(users,
		dto.NewPaginationMeta(pagination.Page, pagination.PageSize, total)))
}

func (h *UserHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.userService.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	if user == nil {
		return c.Status(404).JSON(dto.NewErrorResponse("NOT_FOUND", "User not found"))
	}
	return c.JSON(dto.NewSuccessResponse(user))
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	tenant := c.Locals("tenant").(string)
	var req users.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}
	user, err := h.userService.Create(c.Context(), tenant, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	return c.Status(201).JSON(dto.NewSuccessResponse(user))
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req users.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}
	user, err := h.userService.Update(c.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	return c.JSON(dto.NewSuccessResponse(user))
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.userService.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	return c.JSON(dto.NewMessageResponse("User deleted successfully"))
}
```

Endpoint estándar por entidad:

| Método | Ruta            | Handler | Descripción           |
| ------ | --------------- | ------- | --------------------- |
| GET    | `/{entity}`     | List    | Listar con paginación |
| GET    | `/{entity}/:id` | Get     | Obtener por ID        |
| POST   | `/{entity}`     | Create  | Crear nuevo           |
| PUT    | `/{entity}/:id` | Update  | Actualizar parcial    |
| DELETE | `/{entity}/:id` | Delete  | Borrado lógico        |

### 3.2 `adapters/db/<entidad>/<entidad>_repository.go`

Adapter de salida: Domain → PostgreSQL con pgx.

```go
package db

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nova/backend/internal/domain/users"
)

type PgUserRepository struct {
	pool *pgxpool.Pool
}

func NewPgUserRepository(pool *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{pool: pool}
}

func (r *PgUserRepository) FindByID(ctx context.Context, id string) (*users.User, error) {
	query := `SELECT usr_id, usr_name, ... FROM eamusers WHERE usr_id = $1`
	var user users.User
	err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, ...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}
// ... resto de implementaciones
```

Reglas:

- `package db`
- `pgx.ErrNoRows` → devolver `nil, nil` (no es error)
- Nombres de tablas con prefijo `eam` (`eamusers`, `eamstores`, etc.)
- Columnas con prefijo `usr_`, `sto_`, etc.

---

## 4. Backend: Wire y Routes

### 4.1 `infrastructure/wire/wire.go`

Agregar al `NewContainer`:

```go
// Repositorio
userRepo := userssecondary.NewPgUserRepository(pool)

// Service
userService := users.NewUserService(userRepo)

// Handler
UserHandler: usersapi.NewUserHandler(userService),
```

### 4.2 `cmd/server/main.go`

Agregar las rutas:

```go
usersGroup := protected.Group("/users")
usersGroup.Get("/", c.UserHandler.List)
usersGroup.Get("/:id", c.UserHandler.Get)
usersGroup.Post("/", c.UserHandler.Create)
usersGroup.Put("/:id", c.UserHandler.Update)
usersGroup.Delete("/:id", c.UserHandler.Delete)
```

---

## 5. Frontend: Models y Service

### 5.1 `features/<entidad>/models/<entidad>.model.ts`

```ts
export interface User {
  id: string;
  name: string;
  email: string;
  status: 'active' | 'inactive' | 'suspended';
  createdAt: string;
  updatedAt: string;
}

export interface CreateUserDto {
  name: string;
  email: string;
  password?: string;
  status?: 'active' | 'inactive' | 'suspended';
}

export interface UpdateUserDto {
  name?: string;
  email?: string;
  status?: 'active' | 'inactive' | 'suspended';
}
```

### 5.2 `features/<entidad>/services/<entidad>.service.ts`

```ts
import { Injectable, signal, computed, inject } from '@angular/core';
import { ApiService } from '@core/services/api.service';
import { UiStore } from '@core/stores/ui.store';
import { User, CreateUserDto, UpdateUserDto } from '../models/user.model';

@Injectable({ providedIn: 'root' })
export class UserService {
  private readonly api = inject(ApiService);
  private readonly uiStore = inject(UiStore);

  readonly users = signal<User[]>([]);
  readonly selectedUser = signal<User | null>(null);
  readonly meta = signal({ page: 1, pageSize: 20, total: 0, totalPages: 0 });

  async loadUsers(params = {}) {
    this.uiStore.setLoading(true);
    try {
      const res = await this.api.get<User[]>('/users', params).toPromise();
      if (res?.success && res.data) {
        this.users.set(res.data);
        this.meta.set({ page: res.meta.page, ... });
      }
    } finally {
      this.uiStore.setLoading(false);
    }
  }

  async createUser(dto: CreateUserDto) {
    const res = await this.api.post<User>('/users', dto).toPromise();
    if (res?.success) this.uiStore.success('Creado', 'Usuario creado correctamente');
    return res?.data ?? null;
  }

  async updateUser(id: string, dto: UpdateUserDto) { /* ... */ }
  async deleteUser(id: string) { /* ... */ }
}
```

---

## 6. Frontend: Screens

### 6.1 `features/<entidad>/screens/<entidad>-list/`

Componente standalone con signals y TanStack Table.

Estructura del template:

```
ToolbarComponent (compartido)
├── create, delete, refresh, print, prev/next
QueryBuilderComponent (compartido)
├── selector de queries guardadas
DataGridComponent (TanStack Table)
├── columnas definidas en el TS
EntityDetailComponent (drawer lateral)
├── tabs: view / comments / documents / audit
├── EntityFormComponent para edición
├── RelatedInfoComponent
```

El componente sigue este ciclo:

```ts
@Component({ selector: 'app-user-list', standalone: true, ... })
export class UserListComponent implements OnInit {
  // Inyectar: UserService, QueryService, TranslationService, UiStore

  ngOnInit() {
    // 1. Cargar traducciones del archivo 'users'
    this.translate.load('users').subscribe(translations => {
      this.t = translations;
      this.buildTabs();       // con traducciones
      this.buildFormFields(); // con traducciones
      this.buildColumns();    // con traducciones
    });

    // 2. Cargar queries guardadas y ejecutar la default
    this.queryService.loadByGridId(GRID_IDS.USERS).subscribe(() => {
      const defaultQuery = this.queryService.queries().find(q => q.isDefault);
      if (defaultQuery) this.executeQuery(defaultQuery.query);
    });
  }
}
```

Columnas del grid:

```ts
this.columns = [
  { accessorKey: 'id', header: 'ID', size: 80 },
  { accessorKey: 'name', header: 'Nombre', size: 200 },
  { accessorKey: 'email', header: 'Email' },
  {
    accessorKey: 'status',
    header: 'Estado',
    cell: info => (status === 'active' ? 'Activo' : 'Inactivo'),
  },
  {
    accessorKey: 'createdAt',
    header: 'Fecha Creación',
    cell: info => new Date(info.getValue()).toLocaleDateString(),
  },
];
```

### 6.2 `features/<entidad>/screens/<entidad>-detail/`

Pantalla de detalle/edición para ruta `/:id`. Sigue el mismo patrón de componentes compartidos.

---

## 7. Frontend: Routes

En `app.routes.ts`, lazy loading:

```ts
{
  path: 'users',
  loadComponent: () =>
    import('./features/users/screens/user-list/user-list.component')
      .then(m => m.UserListComponent),
},
{
  path: 'users/:id',
  loadComponent: () =>
    import('./features/users/screens/user-detail/user-detail.component')
      .then(m => m.UserDetailComponent),
},
```

---

## 8. Checklist rápida

### Backend

- [ ] `domain/<ent>/entity.go` — struct + métodos de negocio
- [ ] `domain/<ent>/ports.go` — interface del repositorio
- [ ] `domain/<ent>/dto.go` — request/response
- [ ] `domain/<ent>/service.go` — lógica de negocio
- [ ] `adapters/api/<ent>/handler.go` — HTTP handler (Fiber)
- [ ] `adapters/db/<ent>/<ent>_repository.go` — PostgreSQL (pgx)
- [ ] `infrastructure/wire/wire.go` — DI wiring
- [ ] `cmd/server/main.go` — rutas REST
- [ ] `go build ./...` — compila

### Frontend

- [ ] `features/<ent>/models/<ent>.model.ts` — interfaces TS
- [ ] `features/<ent>/services/<ent>.service.ts` — signals + API calls
- [ ] `features/<ent>/screens/<ent>-list/` — grid + toolbar + drawer
- [ ] `features/<ent>/screens/<ent>-detail/` — detalle/edición
- [ ] `app.routes.ts` — lazy loading
- [ ] `pnpm start` — compila sin errores

### Base de datos

- [ ] Migración UP en `migrations/tenant/`
- [ ] Migración DOWN en `migrations/tenant/`
- [ ] Seed data si aplica en `migrations/tenant/`
