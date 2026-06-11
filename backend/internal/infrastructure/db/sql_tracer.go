package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/valyala/fasthttp"
)

// QueryTracer implements pgx.QueryTracer to log all SQL queries
type QueryTracer struct{}

// extractTenantFromCtx extracts tenant code from Go context (supports fasthttp.RequestCtx)
func extractTenantFromCtx(ctx context.Context) string {
	if fctx, ok := ctx.(*fasthttp.RequestCtx); ok {
		if tenant, ok := fctx.UserValue(tenantKey).(string); ok {
			return tenant
		}
	}
	return ""
}

func (t *QueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	tenant := extractTenantFromCtx(ctx)
	if tenant != "" {
		log.Printf("[SQL] [%s] --> %s | args=%v", tenant, data.SQL, data.Args)
	} else {
		log.Printf("[SQL] --> %s | args=%v", data.SQL, data.Args)
	}
	return ctx
}

func (t *QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	tenant := extractTenantFromCtx(ctx)
	if data.Err != nil {
		if tenant != "" {
			log.Printf("[SQL] [%s] <-- ERROR: %v", tenant, data.Err)
		} else {
			log.Printf("[SQL] <-- ERROR: %v", data.Err)
		}
	} else {
		if tenant != "" {
			log.Printf("[SQL] [%s] <-- OK rows=%d cmd=%s", tenant, data.CommandTag.RowsAffected(), data.CommandTag.String())
		} else {
			log.Printf("[SQL] <-- OK rows=%d cmd=%s", data.CommandTag.RowsAffected(), data.CommandTag.String())
		}
	}
}
