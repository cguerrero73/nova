package wire

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nova/backend/internal/config"
	"github.com/nova/backend/internal/infrastructure/db"

	// API adapters (HTTP handlers)
	authapi "github.com/nova/backend/internal/adapters/api/auth"
	eventsapi "github.com/nova/backend/internal/adapters/api/events"
	objectsapi "github.com/nova/backend/internal/adapters/api/objects"
	orgsapi "github.com/nova/backend/internal/adapters/api/organizations"
	partsapi "github.com/nova/backend/internal/adapters/api/parts"
	stocksapi "github.com/nova/backend/internal/adapters/api/stocks"
	storesapi "github.com/nova/backend/internal/adapters/api/stores"
	structureapi "github.com/nova/backend/internal/adapters/api/structure"
	syscodesapi "github.com/nova/backend/internal/adapters/api/syscodes"
	usersapi "github.com/nova/backend/internal/adapters/api/users"

	// DB adapters (repositories)
	authdb "github.com/nova/backend/internal/adapters/db/auth"
	eventsdb "github.com/nova/backend/internal/adapters/db/events"
	objectsdb "github.com/nova/backend/internal/adapters/db/objects"
	orgsdb "github.com/nova/backend/internal/adapters/db/organizations"
	partsdb "github.com/nova/backend/internal/adapters/db/parts"
	stocksdb "github.com/nova/backend/internal/adapters/db/stocks"
	storesdb "github.com/nova/backend/internal/adapters/db/stores"
	structuredb "github.com/nova/backend/internal/adapters/db/structure"
	syscodesdb "github.com/nova/backend/internal/adapters/db/syscodes"
	usersdb "github.com/nova/backend/internal/adapters/db/users"

	// Domain services
	"github.com/nova/backend/internal/domain/auth"
	"github.com/nova/backend/internal/domain/events"
	"github.com/nova/backend/internal/domain/objects"
	"github.com/nova/backend/internal/domain/organizations"
	"github.com/nova/backend/internal/domain/parts"
	"github.com/nova/backend/internal/domain/stocks"
	"github.com/nova/backend/internal/domain/stores"
	"github.com/nova/backend/internal/domain/structure"
	"github.com/nova/backend/internal/domain/syscodes"
	"github.com/nova/backend/internal/domain/users"
)

// Container holds all dependencies
type Container struct {
	// Handlers (API adapters)
	AuthHandler      *authapi.AuthHandler
	UserHandler      *usersapi.UserHandler
	OrgHandler       *orgsapi.OrganizationHandler
	ObjectHandler    *objectsapi.ObjectHandler
	StructureHandler *structureapi.StructureHandler
	PartHandler      *partsapi.PartHandler
	StoreHandler     *storesapi.StoreHandler
	StockHandler     *stocksapi.StockHandler
	EventHandler     *eventsapi.EventHandler
	SyscodeHandler   *syscodesapi.SysCodeHandler
}

// NewContainer wires all dependencies
func NewContainer(pool *pgxpool.Pool, cfg *config.Config) *Container {
	// DB adapters (repositories)
	authUserRepo := authdb.NewPgUserRepository(pool)
	authSessionRepo := authdb.NewPgSessionRepository(pool)
	userRepo := usersdb.NewPgUserRepository(pool)
	orgRepo := orgsdb.NewPgOrganizationRepository(pool)
	objectRepo := objectsdb.NewPgObjectRepository(pool)
	structureRepo := structuredb.NewPgStructureRepository(pool)
	partRepo := partsdb.NewPgPartRepository(pool)
	storeRepo := storesdb.NewPgStoreRepository(pool)
	binRepo := storesdb.NewPgBinRepository(pool)
	stockRepo := stocksdb.NewPgStockRepository(pool)
	binStockRepo := stocksdb.NewPgBinStockRepository(pool)
	eventRepo := eventsdb.NewPgEventRepository(pool)
	syscodeRepo := syscodesdb.NewPgSysCodeRepository(pool)

	// Domain services
	authService := auth.NewAuthService(authUserRepo, authSessionRepo, cfg.JWT)
	userService := users.NewUserService(userRepo)
	orgService := organizations.NewOrganizationService(orgRepo)
	objectService := objects.NewObjectService(objectRepo)
	structureService := structure.NewStructureService(structureRepo)
	partService := parts.NewPartService(partRepo)
	storeService := stores.NewStoreService(storeRepo, binRepo)
	stockService := stocks.NewStockService(stockRepo, binStockRepo)
	eventService := events.NewEventService(eventRepo)
	syscodeService := syscodes.NewSysCodeService(syscodeRepo)

	// API adapters (handlers)
	return &Container{
		AuthHandler:      authapi.NewAuthHandler(authService),
		UserHandler:      usersapi.NewUserHandler(userService),
		OrgHandler:       orgsapi.NewOrganizationHandler(orgService),
		ObjectHandler:    objectsapi.NewObjectHandler(objectService),
		StructureHandler: structureapi.NewStructureHandler(structureService),
		PartHandler:      partsapi.NewPartHandler(partService),
		StoreHandler:     storesapi.NewStoreHandler(storeService),
		StockHandler:     stocksapi.NewStockHandler(stockService),
		EventHandler:     eventsapi.NewEventHandler(eventService),
		SyscodeHandler:   syscodesapi.NewSysCodeHandler(syscodeService),
	}
}

// InitializeDB creates the database connection
func InitializeDB(cfg *config.Config) (*db.PostgresDB, error) {
	return db.NewPostgresDB(db.DatabaseConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
		Schema:   cfg.Database.Schema,
	})
}
