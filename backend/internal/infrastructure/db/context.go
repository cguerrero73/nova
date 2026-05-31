package db

import (
	"context"
	"fmt"

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

const connLocalsKey = "dbConn"

// SetTenantSchema acquires a connection from the pool, sets the search_path
// to the tenant schema, and stores it in fasthttp.RequestCtx.UserValue.
// Callers must ensure the connection is released after the request completes.
func SetTenantSchema(ctx *fasthttp.RequestCtx, pool *pgxpool.Pool, tenant string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquiring connection: %w", err)
	}

	schemaName := fmt.Sprintf("tenant_%s", tenant)
	_, err = conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", schemaName))
	if err != nil {
		conn.Release()
		return fmt.Errorf("setting search_path: %w", err)
	}

	ctx.SetUserValue(connLocalsKey, conn)
	return nil
}

// ReleaseTenantConn releases the connection stored in fasthttp.RequestCtx, if any.
func ReleaseTenantConn(ctx *fasthttp.RequestCtx) {
	if conn, ok := ctx.UserValue(connLocalsKey).(*pgxpool.Conn); ok {
		conn.Release()
		ctx.RemoveUserValue(connLocalsKey)
	}
}

// GetQueryEngine returns a tenant-scoped connection when available (stored in
// the fasthttp request context by the tenant middleware), falling back to the
// pool. This allows repositories to transparently use the per-request
// connection (with the correct search_path).
func GetQueryEngine(ctx context.Context, pool *pgxpool.Pool) QueryEngine {
	// The context passed by Fiber handlers is *fasthttp.RequestCtx.
	// If the middleware stored a connection there, use it.
	if fctx, ok := ctx.(*fasthttp.RequestCtx); ok {
		if conn, ok := fctx.UserValue(connLocalsKey).(*pgxpool.Conn); ok {
			return conn
		}
	}
	return pool
}
