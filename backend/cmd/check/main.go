package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/nova/backend/internal/config"
)

var cfg *config.Config

func main() {
	_ = "acme" // placeholder

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to tenant schema with search_path including public
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/nova",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port,
	)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close(context.Background())

	// Set search_path to include both tenant and public (for uuid function)
	fmt.Println("Setting search_path to tenant_acme, public...")
	_, err = conn.Exec(context.Background(), "SET search_path TO tenant_acme, public")
	if err != nil {
		log.Fatalf("Failed to set search_path: %v", err)
	}

	// Verify search_path
	var currentPath string
	err = conn.QueryRow(context.Background(), "SHOW search_path").Scan(&currentPath)
	if err != nil {
		log.Fatalf("Failed to verify search_path: %v", err)
	}
	fmt.Printf("Current search_path: %s\n", currentPath)

	// Test uuid
	fmt.Println("\nTesting uuid_generate_v4()...")
	var uuidTest string
	err = conn.QueryRow(context.Background(), "SELECT uuid_generate_v4()::text").Scan(&uuidTest)
	if err != nil {
		fmt.Printf("  Failed: %v\n", err)
	} else {
		fmt.Printf("  Works: %s\n", uuidTest)
	}

	// Check tables in tenant schema
	fmt.Println("\nTables in tenant_acme:")
	rows, err := conn.Query(context.Background(),
		"SELECT tablename FROM pg_tables WHERE schemaname = 'tenant_acme' ORDER BY tablename")
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Printf("  - %s\n", name)
	}
}
