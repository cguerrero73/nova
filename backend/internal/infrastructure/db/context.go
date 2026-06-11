package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyala/fasthttp"
)

// QueryEngine is the subset of pgx operations used by all repositories.
// Both *pgxpool.Pool and *pgxpool.Conn implement this interface.
type QueryEngine interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

var _ QueryEngine = (*pgxpool.Pool)(nil)
var _ QueryEngine = (*pgxpool.Conn)(nil)

const (
	tenantKey = "tenant"
	connKey   = "dbConn"
)

// TenantContextKey is the type used as context key for tenant.
// This MUST match the type used in the middleware.
// Using a struct type as key to avoid collisions.
type TenantContextKey struct{}

// RunInTenantTx executes a function within a tenant-scoped transaction.
// It automatically sets the search_path for the transaction.
// This is the recommended way to execute queries when using PgBouncer in transaction mode.
func RunInTenantTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tenant := extractTenantCode(ctx)

	// DEBUG: log ctx type
	log.Printf("[RunInTenantTx] ctx type=%T tenant='%s'", ctx, tenant)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	// Set search_path al inicio de la transacción si hay tenant
	if tenant != "" {
		schemaName := fmt.Sprintf("tenant_%s", tenant)
		if _, err := tx.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", schemaName)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("setting search_path: %w", err)
		}
		log.Printf("[TX] [%s] BEGIN + SET search_path TO %s, public", tenant, schemaName)
	} else {
		log.Printf("[TX] [no-tenant] BEGIN (using public schema)")
	}

	// Ejecutar la función con la transacción
	if err := fn(tx); err != nil {
		tx.Rollback(ctx)
		if tenant != "" {
			log.Printf("[TX] [%s] ROLLBACK (error: %v)", tenant, err)
		} else {
			log.Printf("[TX] [no-tenant] ROLLBACK (error: %v)", err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		if tenant != "" {
			log.Printf("[TX] [%s] COMMIT ERROR: %v", tenant, err)
		} else {
			log.Printf("[TX] [no-tenant] COMMIT ERROR: %v", err)
		}
		return err
	}

	if tenant != "" {
		log.Printf("[TX] [%s] COMMIT OK", tenant)
	} else {
		log.Printf("[TX] [no-tenant] COMMIT OK")
	}
	return nil
}

// RunInTx executes a function within a transaction without tenant scope.
// Use this for global operations (e.g., tenant creation) that don't need isolation.
func RunInTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// GetEngine returns a tenant-scoped connection when available (stored in
// the fasthttp request context by the tenant middleware), falling back to the
// pool. This is for backward compatibility with session-mode connections.
// For new code, prefer RunInTenantTx.
func GetEngine(ctx context.Context, pool *pgxpool.Pool) QueryEngine {
	conn := extractTenantConn(ctx)
	tenant := extractTenantCode(ctx)

	if conn != nil {
		return conn
	}

	// Log warning if tenant is configured but no connection was found.
	if tenant != "" {
		log.Printf("[WARN] Tenant '%s' set but no tenant-scoped connection available. "+
			"Using pool without tenant isolation.", tenant)
	}

	return pool
}

// SetTenantSchema acquires a connection from the pool, sets the search_path
// to the tenant schema, and stores it in fasthttp.RequestCtx.UserValue.
// This is for session mode. For transaction mode, use RunInTenantTx.
func SetTenantSchema(ctx *fasthttp.RequestCtx, pool *pgxpool.Pool, tenant string) error {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("acquiring connection: %w", err)
	}

	schemaName := fmt.Sprintf("tenant_%s", tenant)
	_, err = conn.Exec(context.Background(), fmt.Sprintf("SET search_path TO %s, public", schemaName))
	if err != nil {
		conn.Release()
		return fmt.Errorf("setting search_path: %w", err)
	}

	ctx.SetUserValue(tenantKey, tenant)
	ctx.SetUserValue(connKey, conn)
	log.Printf("[SetTenantSchema] tenant='%s' stored in ctx.UserValue", tenant)
	return nil
}

// ReleaseTenantConn releases the connection stored in fasthttp.RequestCtx, if any.
func ReleaseTenantConn(ctx *fasthttp.RequestCtx) {
	if conn, ok := ctx.UserValue(connKey).(*pgxpool.Conn); ok {
		conn.Release()
		ctx.RemoveUserValue(connKey)
		ctx.RemoveUserValue(tenantKey)
	}
}

// extractTenantConn safely extracts the tenant connection from the context.
func extractTenantConn(ctx context.Context) *pgxpool.Conn {
	if fctx, ok := ctx.(*fasthttp.RequestCtx); ok {
		if conn, ok := fctx.UserValue(connKey).(*pgxpool.Conn); ok {
			return conn
		}
	}
	return nil
}

// extractTenantCode extracts the tenant code from Go context.
// It first tries the Go context (set via context.WithValue in middleware),
// then falls back to fasthttp.RequestCtx.UserValue for backward compatibility.
func extractTenantCode(ctx context.Context) string {
	// First try Go context (primary method)
	if tenant := ctx.Value(TenantContextKey{}); tenant != nil {
		if s, ok := tenant.(string); ok {
			return s
		}
	}

	// Fallback to fasthttp.RequestCtx (backward compatibility)
	if fctx, ok := ctx.(*fasthttp.RequestCtx); ok {
		if tenant, ok := fctx.UserValue(tenantKey).(string); ok {
			return tenant
		}
	}
	return ""
}
