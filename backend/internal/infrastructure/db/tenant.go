package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantDB represents a database connection scoped to a specific tenant
type TenantDB struct {
	Pool   *pgxpool.Pool
	Schema string
	Tenant string
}

// GetTenantPool returns a connection pool configured for the specified tenant schema
func (db *PostgresDB) GetTenantPool(tenantCode string) (*TenantDB, error) {
	schema := "tenant_" + tenantCode

	poolConfig := db.Pool.Config().Copy()

	// Override search_path for this tenant
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schema

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return &TenantDB{
		Pool:   pool,
		Schema: schema,
		Tenant: tenantCode,
	}, nil
}

// Close closes the tenant connection pool
func (tdb *TenantDB) Close() {
	tdb.Pool.Close()
}

// Ping tests the connection to the tenant database
func (tdb *TenantDB) Ping(ctx context.Context) error {
	return tdb.Pool.Ping(ctx)
}

// CreateTenantSchema creates a new schema for a tenant
func (db *PostgresDB) CreateTenantSchema(tenantCode string) error {
	schemaName := "tenant_" + tenantCode
	_, err := db.Pool.Exec(context.Background(), "CREATE SCHEMA IF NOT EXISTS "+schemaName)
	return err
}

// DeleteTenantSchema deletes a tenant schema
func (db *PostgresDB) DeleteTenantSchema(tenantCode string) error {
	schemaName := "tenant_" + tenantCode
	_, err := db.Pool.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schemaName+" CASCADE")
	return err
}
