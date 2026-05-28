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
	"github.com/nova/backend/internal/infrastructure/middleware"
	"github.com/nova/backend/internal/infrastructure/wire"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := wire.InitializeDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("Connected to database")

	// Wire dependencies
	c := wire.NewContainer(database.Pool, cfg)

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

	// Auth middleware
	authMw := middleware.NewAuthMiddleware(cfg.JWT)

	// Routes
	api := app.Group("/api/v1")

	// Public routes (no auth required)
	authGroup := api.Group("/auth")
	authGroup.Post("/login", c.AuthHandler.Login)
	authGroup.Post("/register", c.AuthHandler.Register)
	authGroup.Post("/refresh", c.AuthHandler.Refresh)

	// Protected routes (auth required)
	protected := api.Group("", authMw.Authenticate())

	// Auth routes requiring auth
	protected.Post("/logout", c.AuthHandler.Logout)
	protected.Get("/auth/me", c.AuthHandler.Me)

	// Users CRUD
	usersGroup := protected.Group("/users")
	usersGroup.Get("/", c.UserHandler.List)
	usersGroup.Get("/:id", c.UserHandler.Get)
	usersGroup.Post("/", c.UserHandler.Create)
	usersGroup.Put("/:id", c.UserHandler.Update)
	usersGroup.Delete("/:id", c.UserHandler.Delete)

	// Organizations CRUD
	orgsGroup := protected.Group("/organizations")
	orgsGroup.Get("/", c.OrgHandler.List)
	orgsGroup.Get("/:id", c.OrgHandler.Get)
	orgsGroup.Post("/", c.OrgHandler.Create)
	orgsGroup.Put("/:id", c.OrgHandler.Update)
	orgsGroup.Delete("/:id", c.OrgHandler.Delete)

	// Objects CRUD
	objectsGroup := protected.Group("/objects")
	objectsGroup.Get("/", c.ObjectHandler.List)
	objectsGroup.Get("/:id", c.ObjectHandler.Get)
	objectsGroup.Post("/", c.ObjectHandler.Create)
	objectsGroup.Put("/:id", c.ObjectHandler.Update)
	objectsGroup.Delete("/:id", c.ObjectHandler.Delete)
	objectsGroup.Get("/:id/children", c.ObjectHandler.GetChildren)

	// Structure CRUD
	structureGroup := protected.Group("/structure")
	structureGroup.Get("/", c.StructureHandler.List)
	structureGroup.Post("/", c.StructureHandler.Create)
	structureGroup.Put("/:id", c.StructureHandler.Update)
	structureGroup.Delete("/:id", c.StructureHandler.Delete)

	// Parts CRUD
	partsGroup := protected.Group("/parts")
	partsGroup.Get("/", c.PartHandler.List)
	partsGroup.Get("/:id", c.PartHandler.Get)
	partsGroup.Post("/", c.PartHandler.Create)
	partsGroup.Put("/:id", c.PartHandler.Update)
	partsGroup.Delete("/:id", c.PartHandler.Delete)

	// Stores CRUD
	storesGroup := protected.Group("/stores")
	storesGroup.Get("/", c.StoreHandler.List)
	storesGroup.Get("/:id", c.StoreHandler.Get)
	storesGroup.Post("/", c.StoreHandler.Create)
	storesGroup.Put("/:id", c.StoreHandler.Update)
	storesGroup.Delete("/:id", c.StoreHandler.Delete)

	// Stocks CRUD
	stocksGroup := protected.Group("/stocks")
	stocksGroup.Get("/", c.StockHandler.List)
	stocksGroup.Get("/low-stock", c.StockHandler.GetLowStock)
	stocksGroup.Get("/:id", c.StockHandler.Get)
	stocksGroup.Post("/", c.StockHandler.Create)
	stocksGroup.Put("/:id", c.StockHandler.Update)
	stocksGroup.Delete("/:id", c.StockHandler.Delete)
	stocksGroup.Post("/:id/adjust", c.StockHandler.AdjustQuantity)

	// Bin Stocks
	binStocksGroup := protected.Group("/bin-stocks")
	binStocksGroup.Get("/", c.StockHandler.ListBinStocks)
	binStocksGroup.Post("/", c.StockHandler.CreateBinStock)
	binStocksGroup.Put("/:id", c.StockHandler.UpdateBinStock)
	binStocksGroup.Delete("/:id", c.StockHandler.DeleteBinStock)

	// Events CRUD
	eventsGroup := protected.Group("/events")
	eventsGroup.Get("/", c.EventHandler.List)
	eventsGroup.Get("/:id", c.EventHandler.Get)
	eventsGroup.Post("/", c.EventHandler.Create)
	eventsGroup.Put("/:id", c.EventHandler.Update)
	eventsGroup.Delete("/:id", c.EventHandler.Delete)
	eventsGroup.Get("/object/:code/:org", c.EventHandler.GetByObject)
	eventsGroup.Put("/:id/status", c.EventHandler.UpdateStatus)

	// SysCodes CRUD
	syscodesGroup := protected.Group("/syscodes")
	syscodesGroup.Get("/", c.SyscodeHandler.List)
	syscodesGroup.Get("/:id", c.SyscodeHandler.Get)
	syscodesGroup.Post("/", c.SyscodeHandler.Create)
	syscodesGroup.Put("/:id", c.SyscodeHandler.Update)
	syscodesGroup.Delete("/:id", c.SyscodeHandler.Delete)
	syscodesGroup.Get("/type/:type", c.SyscodeHandler.GetByType)

	// Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdown
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
