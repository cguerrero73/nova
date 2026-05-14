package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/nova/backend/internal/config"
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
	"github.com/nova/backend/internal/infrastructure/db"
	"github.com/nova/backend/internal/infrastructure/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.NewPostgresDB(db.DatabaseConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
		Schema:   cfg.Database.Schema,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("Connected to database")

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Health check (no auth required)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now(),
		})
	})

	// Tenant middleware for all routes
	tenantMiddleware := middleware.NewTenantMiddleware()
	app.Use(tenantMiddleware.ExtractTenant())

	// Initialize repositories with pgx
	userRepo := users.NewPgUserRepository(database.Pool)
	authUserRepo := auth.NewPgUserRepository(database.Pool)
	sessionRepo := auth.NewPgSessionRepository(database.Pool)
	orgRepo := organizations.NewPgOrganizationRepository(database.Pool)
	objectRepo := objects.NewPgObjectRepository(database.Pool)
	structureRepo := structure.NewPgStructureRepository(database.Pool)
	partRepo := parts.NewPgPartRepository(database.Pool)
	storeRepo := stores.NewPgStoreRepository(database.Pool)
	binRepo := stores.NewPgBinRepository(database.Pool)
	stockRepo := stocks.NewPgStockRepository(database.Pool)
	binStockRepo := stocks.NewPgBinStockRepository(database.Pool)
	eventRepo := events.NewPgEventRepository(database.Pool)
	syscodeRepo := syscodes.NewPgSysCodeRepository(database.Pool)

	// Initialize services
	authService := auth.NewAuthService(authUserRepo, sessionRepo, cfg.JWT)
	userService := users.NewUserService(userRepo)
	orgService := organizations.NewOrganizationService(orgRepo)
	objectService := objects.NewObjectService(objectRepo)
	structureService := structure.NewStructureService(structureRepo)
	partService := parts.NewPartService(partRepo)
	storeService := stores.NewStoreService(storeRepo, binRepo)
	stockService := stocks.NewStockService(stockRepo, binStockRepo)
	eventService := events.NewEventService(eventRepo)
	syscodeService := syscodes.NewSysCodeService(syscodeRepo)

	// Initialize handlers
	authHandler := auth.NewAuthHandler(authService)
	userHandler := users.NewUserHandler(userService)
	orgHandler := organizations.NewOrganizationHandler(orgService)
	objectHandler := objects.NewObjectHandler(objectService)
	structureHandler := structure.NewStructureHandler(structureService)
	partHandler := parts.NewPartHandler(partService)
	storeHandler := stores.NewStoreHandler(storeService)
	stockHandler := stocks.NewStockHandler(stockService)
	eventHandler := events.NewEventHandler(eventService)
	syscodeHandler := syscodes.NewSysCodeHandler(syscodeService)

	// Auth middleware
	authMw := middleware.NewAuthMiddleware(cfg.JWT)

	// Routes
	api := app.Group("/api/v1")

	// Public routes (no auth required)
	authGroup := api.Group("/auth")
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/refresh", authHandler.Refresh)

	// Protected routes (auth required)
	protected := api.Group("", authMw.Authenticate())

	// Auth routes requiring auth
	protected.Post("/logout", authHandler.Logout)
	protected.Get("/auth/me", authHandler.Me)

	// Users CRUD
	usersGroup := protected.Group("/users")
	usersGroup.Get("/", userHandler.List)
	usersGroup.Get("/:id", userHandler.Get)
	usersGroup.Post("/", userHandler.Create)
	usersGroup.Put("/:id", userHandler.Update)
	usersGroup.Delete("/:id", userHandler.Delete)

	// Organizations CRUD
	orgsGroup := protected.Group("/organizations")
	orgsGroup.Get("/", orgHandler.List)
	orgsGroup.Get("/:id", orgHandler.Get)
	orgsGroup.Post("/", orgHandler.Create)
	orgsGroup.Put("/:id", orgHandler.Update)
	orgsGroup.Delete("/:id", orgHandler.Delete)

	// Objects CRUD
	objectsGroup := protected.Group("/objects")
	objectsGroup.Get("/", objectHandler.List)
	objectsGroup.Get("/:id", objectHandler.Get)
	objectsGroup.Post("/", objectHandler.Create)
	objectsGroup.Put("/:id", objectHandler.Update)
	objectsGroup.Delete("/:id", objectHandler.Delete)
	objectsGroup.Get("/:id/children", objectHandler.GetChildren)

	// Structure CRUD
	structureGroup := protected.Group("/structure")
	structureGroup.Get("/", structureHandler.List)
	structureGroup.Post("/", structureHandler.Create)
	structureGroup.Put("/:id", structureHandler.Update)
	structureGroup.Delete("/:id", structureHandler.Delete)

	// Parts CRUD
	partsGroup := protected.Group("/parts")
	partsGroup.Get("/", partHandler.List)
	partsGroup.Get("/:id", partHandler.Get)
	partsGroup.Post("/", partHandler.Create)
	partsGroup.Put("/:id", partHandler.Update)
	partsGroup.Delete("/:id", partHandler.Delete)

	// Stores CRUD
	storesGroup := protected.Group("/stores")
	storesGroup.Get("/", storeHandler.List)
	storesGroup.Get("/:id", storeHandler.Get)
	storesGroup.Post("/", storeHandler.Create)
	storesGroup.Put("/:id", storeHandler.Update)
	storesGroup.Delete("/:id", storeHandler.Delete)

	// Stocks CRUD
	stocksGroup := protected.Group("/stocks")
	stocksGroup.Get("/", stockHandler.List)
	stocksGroup.Get("/low-stock", stockHandler.GetLowStock)
	stocksGroup.Get("/:id", stockHandler.Get)
	stocksGroup.Post("/", stockHandler.Create)
	stocksGroup.Put("/:id", stockHandler.Update)
	stocksGroup.Delete("/:id", stockHandler.Delete)
	stocksGroup.Post("/:id/adjust", stockHandler.AdjustQuantity)

	// Bin Stocks
	binStocksGroup := protected.Group("/bin-stocks")
	binStocksGroup.Get("/", stockHandler.ListBinStocks)
	binStocksGroup.Post("/", stockHandler.CreateBinStock)
	binStocksGroup.Put("/:id", stockHandler.UpdateBinStock)
	binStocksGroup.Delete("/:id", stockHandler.DeleteBinStock)

	// Events CRUD
	eventsGroup := protected.Group("/events")
	eventsGroup.Get("/", eventHandler.List)
	eventsGroup.Get("/:id", eventHandler.Get)
	eventsGroup.Post("/", eventHandler.Create)
	eventsGroup.Put("/:id", eventHandler.Update)
	eventsGroup.Delete("/:id", eventHandler.Delete)
	eventsGroup.Get("/object/:code/:org", eventHandler.GetByObject)
	eventsGroup.Put("/:id/status", eventHandler.UpdateStatus)

	// SysCodes CRUD
	syscodesGroup := protected.Group("/syscodes")
	syscodesGroup.Get("/", syscodeHandler.List)
	syscodesGroup.Get("/:id", syscodeHandler.Get)
	syscodesGroup.Post("/", syscodeHandler.Create)
	syscodesGroup.Put("/:id", syscodeHandler.Update)
	syscodesGroup.Delete("/:id", syscodeHandler.Delete)
	syscodesGroup.Get("/type/:type", syscodeHandler.GetByType)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "4000"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		},
	})
}
