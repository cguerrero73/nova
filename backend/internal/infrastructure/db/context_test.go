package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

// TestRunInTenantTx_TenantPropagation pins the contract that RunInTenantTx
// inspects the supplied context for the tenant and emits the per-tenant
// `SET search_path` statement inside the transaction before the user fn runs.
//
// This is the regression guard for the production bug fixed in commits 9791fbe
// through 2a602b7: eleven API handlers used to forward c.Context() (the raw
// *fasthttp.RequestCtx) to the service layer, which bypassed the wrapped Go
// context set by the tenant middleware. With c.UserContext() now in use,
// every handler reaches RunInTenantTx with a context carrying
// TenantContextKey{}, and the SQL emitted must include
// `SET search_path TO tenant_<code>, public`.
//
// Two subtests cover the two ctx shapes RunInTenantTx must handle:
//   - wrapped Go context with TenantContextKey{} (the new, correct handler path)
//   - empty context (no tenant -> runs against public schema only)
func TestRunInTenantTx_TenantPropagation(t *testing.T) {
	const userSQL = "SELECT 1"
	const setPathSQL = "SET search_path TO tenant_acme, public"

	tests := []struct {
		name          string
		ctx           context.Context
		expectSetPath bool
	}{
		{
			// Positive: simulates the tenant middleware wrapping the request
			// context with TenantContextKey{} before the handler calls
			// c.UserContext(). RunInTenantTx MUST emit SET search_path before
			// any user SQL, otherwise the per-tenant tables (e.g. eamqueries)
			// are looked up against `public` and the query fails with 42P01.
			name:          "tenant_wrapped_ctx_emits_set_search_path",
			ctx:           context.WithValue(context.Background(), TenantContextKey{}, "acme"),
			expectSetPath: true,
		},
		{
			// Negative (regression guard): when the context carries no tenant,
			// RunInTenantTx MUST NOT emit SET search_path. The transaction
			// runs against the public schema only. This guards against a
			// future change that would silently apply the wrong tenant's
			// schema (or crash with "schema does not exist") to requests
			// that legitimately do not carry a tenant (e.g. global setup
			// flows like tenant provisioning).
			name:          "no_tenant_uses_public_schema",
			ctx:           context.Background(),
			expectSetPath: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer mock.Close()

			mock.ExpectBegin()

			if tt.expectSetPath {
				mock.ExpectExec(setPathSQL).
					WillReturnResult(pgconn.NewCommandTag("SET"))
			}

			mock.ExpectExec(userSQL).
				WillReturnResult(pgconn.NewCommandTag("SELECT 1"))

			mock.ExpectCommit()

			err = RunInTenantTx(tt.ctx, mock, func(tx pgx.Tx) error {
				_, execErr := tx.Exec(tt.ctx, userSQL)
				return execErr
			})

			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
