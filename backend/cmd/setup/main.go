package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/nova/backend/internal/config"
)

var cfg *config.Config

func main() {
	tenant := flag.String("tenant", "", "Tenant code (required)")
	configPath := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	if *tenant == "" {
		fmt.Println("Usage: cleanup -tenant=CODE [-config=PATH]")
		flag.Usage()
		os.Exit(1)
	}

	var err error
	cfg, err = config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to nova database
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/nova",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port,
	)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Failed to connect to nova: %v", err)
	}
	defer conn.Close(context.Background())

	// Tables to KEEP in public (global tables)
	keepTables := map[string]bool{
		"eamtenants":          true,
		"eamtenant_customers": true,
	}

	// Get all tables in public
	rows, err := conn.Query(context.Background(),
		"SELECT tablename FROM pg_tables WHERE schemaname = 'public'")
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}

	var tablesToDrop []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		if !keepTables[name] {
			tablesToDrop = append(tablesToDrop, name)
		}
	}
	rows.Close()

	// Drop tables that don't belong in public
	fmt.Println("Cleaning up public schema...")
	for _, table := range tablesToDrop {
		fmt.Printf("  Dropping %s (not a global table)\n", table)
		_, err = conn.Exec(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			log.Fatalf("Failed to drop %s: %v", table, err)
		}
	}

	// Verify public schema
	rows2, err := conn.Query(context.Background(),
		"SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename")
	if err != nil {
		log.Fatalf("Failed to verify tables: %v", err)
	}
	defer rows2.Close()

	fmt.Println("\nTables remaining in public:")
	for rows2.Next() {
		var name string
		rows2.Scan(&name)
		fmt.Printf("  - %s\n", name)
	}

	// Create tenant schema and enable uuid extension
	fmt.Printf("Creating tenant schema: tenant_%s\n", *tenant)
	_, err = conn.Exec(context.Background(), fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS tenant_%s", *tenant))
	if err != nil {
		log.Fatalf("Failed to create tenant schema: %v", err)
	}

	// Create uuid-ossp extension in tenant schema
	fmt.Println("Creating uuid extension in tenant schema...")
	_, err = conn.Exec(context.Background(), fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" SCHEMA tenant_%s", *tenant))
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			log.Printf("Warning: extension creation: %v", err)
		}
	}

	// Connect to tenant schema
	content, err := os.ReadFile("migrations/tenant/001_init_tenant.up.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Remove extension creation and comments
	var filteredLines []string
	lines := strings.Split(string(content), "\n")
	skipNextExtension := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "-- Enable UUID extension" {
			skipNextExtension = true
			continue
		}
		if skipNextExtension && trimmed == "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" {
			skipNextExtension = false
			continue
		}
		skipNextExtension = false
		if trimmed != "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" && !strings.HasPrefix(trimmed, "--") {
			filteredLines = append(filteredLines, line)
		}
	}
	filteredSQL := strings.Join(filteredLines, "\n")

	// Connect to tenant schema
	tenantConnStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/nova?search_path=tenant_%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, *tenant,
	)

	tenantConn, err := pgx.Connect(context.Background(), tenantConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to tenant schema: %v", err)
	}
	defer tenantConn.Close(context.Background())

	// CRITICAL: Set search_path to include both tenant AND public (for functions like uuid_generate_v4)
	_, err = tenantConn.Exec(context.Background(), fmt.Sprintf("SET search_path TO tenant_%s, public", *tenant))
	if err != nil {
		log.Fatalf("Failed to set search_path: %v", err)
	}

	// Execute statements
	fmt.Printf("Executing tenant tables in tenant_%s schema...\n", *tenant)

	statements := strings.Split(filteredSQL, ";")
	executed := 0
	skipped := 0
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		stmt = strings.TrimRight(stmt, "\n")

		_, err = tenantConn.Exec(context.Background(), stmt)
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "already exists") ||
				strings.Contains(errStr, "duplicate key") ||
				strings.Contains(errStr, "duplicate object") ||
				strings.Contains(errStr, "duplicate index") ||
				strings.Contains(errStr, "duplicate relation") {
				skipped++
				continue
			}
			log.Fatalf("[%d] Statement failed: %v\nStatement: %s", i+1, err, stmt)
		}
		executed++
	}

	fmt.Printf("\nDone! Tenant %s created with %d tables\n", *tenant, executed)
}
