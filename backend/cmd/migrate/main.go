package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/nova/backend/internal/config"
)

var (
	cfg        *config.Config
	configPath string
	tenantCode string
)

func main() {
	migType := flag.String("type", "", "global|tenant (required)")
	tenant := flag.String("tenant", "", "Tenant code (required for tenant type)")
	direction := flag.String("direction", "up", "up|down|status")
	steps := flag.Int("steps", -1, "Number of migrations (-1 = all)")
	seed := flag.Bool("seed", false, "Run seeds")
	bootstrap := flag.Bool("bootstrap", false, "Create schema + migrate + seed")
	configPathFlag := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	configPath = *configPathFlag
	tenantCode = *tenant

	// Handle positional arguments (e.g., "nova-migrate -type=global status")
	// This allows "status" as shorthand instead of "-direction=status"
	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "up", "down", "status":
			*direction = args[0]
		default:
			// Unknown positional arg, ignore or could error
		}
	}

	// Load config
	var err error
	cfg, err = config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate required flags
	if *migType == "" {
		fmt.Println("Usage: nova-migrate -type=global|tenant [options]")
		fmt.Println("")
		fmt.Println("Global migrations:")
		fmt.Println("  nova-migrate -type=global up                    # Apply all global migrations")
		fmt.Println("  nova-migrate -type=global down                 # Rollback last global migration")
		fmt.Println("  nova-migrate -type=global status              # Show global migration status")
		fmt.Println("  nova-migrate -type=global seed                # Run global seeds")
		fmt.Println("")
		fmt.Println("Tenant migrations:")
		fmt.Println("  nova-migrate -type=tenant -tenant=CODE up      # Apply tenant migrations")
		fmt.Println("  nova-migrate -type=tenant -tenant=CODE down   # Rollback last tenant migration")
		fmt.Println("  nova-migrate -type=tenant -tenant=CODE status  # Show tenant migration status")
		fmt.Println("  nova-migrate -type=tenant -tenant=CODE seed    # Run tenant seeds")
		fmt.Println("  nova-migrate -type=tenant -tenant=CODE bootstrap # Create schema + migrate + seed")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  -steps=N     Number of migrations to apply/rollback (-1 = all)")
		fmt.Println("  -config=PATH Path to config file (default: config.yaml)")
		flag.Usage()
		os.Exit(1)
	}

	if *migType == "tenant" && tenantCode == "" {
		log.Fatal("Tenant code is required for tenant migrations (-tenant=CODE)")
	}

	switch *migType {
	case "global":
		runGlobal(*direction, *steps, *seed)
	case "tenant":
		runTenant(tenantCode, *direction, *steps, *seed, *bootstrap)
	default:
		log.Fatalf("Unknown migration type: %s (use global or tenant)", *migType)
	}
}

// runGlobal executes global migrations and/or seeds
func runGlobal(direction string, steps int, runSeed bool) {
	fmt.Println("=== Global Migrations ===")

	// Build connection string for public schema
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?search_path=public&sslmode=disable",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database,
	)

	// Run migrations
	switch direction {
	case "up", "down":
		if err := runMigrations(connStr, "migrations/global", direction, steps); err != nil {
			log.Printf("Migration error: %v", err)
		}
	case "status":
		showStatus(connStr, "migrations/global")
	}

	// Run seeds if requested
	if runSeed {
		fmt.Println("\n=== Global Seeds ===")
		runSeeds(connStr, "seeds/global")
	}
}

// runTenant executes tenant migrations and/or seeds
func runTenant(tenant string, direction string, steps int, runSeed bool, bootstrap bool) {
	fmt.Printf("=== Tenant: %s ===\n", tenant)

	schemaName := fmt.Sprintf("tenant_%s", tenant)

	// Step 1: Create schema if not exists (bootstrap or up)
	if bootstrap || direction == "up" {
		fmt.Printf("\nCreating schema %s if not exists...\n", schemaName)
		if err := createSchemaIfNotExists(schemaName); err != nil {
			log.Fatalf("Failed to create schema: %v", err)
		}
	}

	// Build connection string with search_path including tenant schema and public
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?search_path=%s,public&sslmode=disable",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database, schemaName,
	)

	// Step 2: Run migrations
	if direction == "up" || direction == "down" {
		if err := runMigrations(connStr, "migrations/tenant", direction, steps); err != nil {
			log.Printf("Migration error: %v", err)
		}
	} else if direction == "status" {
		showStatus(connStr, "migrations/tenant")
	}

	// Step 3: Run seeds if requested
	if runSeed || bootstrap {
		fmt.Printf("\n=== Tenant Seeds: %s ===\n", tenant)
		runSeeds(connStr, "seeds/tenant")
	}
}

// createSchemaIfNotExists creates the tenant schema
func createSchemaIfNotExists(schemaName string) error {
	// Connect to postgres to create schema
	adminConnStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/postgres",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port,
	)

	conn, err := pgx.Connect(context.Background(), adminConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer conn.Close(context.Background())

	// Create schema
	_, err = conn.Exec(context.Background(), fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName))
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	fmt.Printf("Schema %s created or already exists\n", schemaName)
	return nil
}

// runMigrations executes migrations using golang-migrate
func runMigrations(connStr string, migrationsPath string, direction string, steps int) error {
	// Get absolute path for migrations directory (not glob pattern)
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create migrate instance with file source
	// The file driver expects: file:///path/to/migrations (directory, not glob)
	sourceURL := fmt.Sprintf("file://%s", absPath)

	m, err := migrate.New(sourceURL, connStr)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Execute based on direction
	switch direction {
	case "up":
		fmt.Printf("Running UP migrations from %s...\n", migrationsPath)
		if steps > 0 {
			err = m.Steps(steps)
			fmt.Printf("Applied %d migration(s)\n", steps)
		} else {
			err = m.Up()
			if err != nil {
				if err == migrate.ErrNoChange {
					fmt.Println("No more migrations to apply")
				} else {
					fmt.Printf("Error: %v\n", err)
				}
			} else {
				fmt.Println("All migrations applied successfully")
			}
		}
	case "down":
		fmt.Printf("Running DOWN migrations from %s...\n", migrationsPath)
		if steps > 0 {
			err = m.Steps(-steps)
			fmt.Printf("Rolled back %d migration(s)\n", steps)
		} else {
			err = m.Down()
			if err != nil {
				if err == migrate.ErrNoChange {
					fmt.Println("No more migrations to rollback")
				} else {
					fmt.Printf("Error: %v\n", err)
				}
			} else {
				fmt.Println("Last migration rolled back")
			}
		}
	}

	// Check for errors (but ignore NoChange)
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// showStatus shows migration status
func showStatus(connStr string, migrationsPath string) {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		log.Printf("Failed to get absolute path: %v", err)
		return
	}

	sourceURL := fmt.Sprintf("file://%s", absPath)

	m, err := migrate.New(sourceURL, connStr)
	if err != nil {
		log.Printf("Failed to create migrate instance: %v", err)
		return
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			fmt.Println("No migrations applied yet")
		} else {
			log.Printf("Failed to get version: %v", err)
		}
		return
	}

	fmt.Printf("Current version: %d, dirty: %v\n", version, dirty)
}

// runSeeds executes seed files
func runSeeds(connStr string, seedsPath string) {
	// Check if seeds directory exists
	if _, err := os.Stat(seedsPath); os.IsNotExist(err) {
		fmt.Printf("Seeds path %s does not exist, skipping\n", seedsPath)
		return
	}

	// Connect to database
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Printf("Failed to connect for seeds: %v", err)
		return
	}
	defer conn.Close(context.Background())

	// Find all SQL files in seeds directory
	files, err := os.ReadDir(seedsPath)
	if err != nil {
		log.Printf("Failed to read seeds directory: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		filePath := filepath.Join(seedsPath, file.Name())
		fmt.Printf("Running seed: %s\n", file.Name())

		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read seed file %s: %v", file.Name(), err)
			continue
		}

		// Split by semicolon and execute each statement
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" || strings.HasPrefix(stmt, "--") {
				continue
			}
			_, err := conn.Exec(context.Background(), stmt)
			if err != nil {
				// Ignore "already exists" errors for seeds
				errStr := err.Error()
				if strings.Contains(errStr, "duplicate key") ||
					strings.Contains(errStr, "duplicate entry") ||
					strings.Contains(errStr, "already exists") {
					continue
				}
				log.Printf("Seed statement failed in %s: %v", file.Name(), err)
			}
		}
		fmt.Printf("  Seed %s completed\n", file.Name())
	}
}
