//go:build integration

package formbuilder_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nova/backend/internal/infrastructure/db"
)

// Integration tests for form builder immutability triggers and audit endpoint.
// These require a real PostgreSQL database with the form builder migrations applied.
//
// Run with: NOVA_TEST_DB_DSN="postgres://..." go test -tags=integration ./internal/adapters/db/formbuilder/...

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	dsn := os.Getenv("NOVA_TEST_DB_DSN")
	if dsn == "" {
		fmt.Println("SKIP: NOVA_TEST_DB_DSN not set — integration tests require a real database")
		os.Exit(0)
	}

	ctx := context.Background()
	var err error
	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	defer testPool.Close()

	// Verify connectivity
	if err := testPool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to ping test database: %v\n", err)
		os.Exit(1)
	}

	// Set search_path to a test tenant schema
	_, err = testPool.Exec(ctx, "SET search_path TO tenant_test, public")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set search_path: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestImmutability_AuditLog_UpdateRaises(t *testing.T) {
	ctx := context.Background()

	// Insert a test audit entry
	var auditID int64
	err := testPool.QueryRow(ctx, `
		INSERT INTO eamform_audit_log (fal_actor_user_id, fal_action, fal_entity_type, fal_entity_id, fal_metadata, fal_created_at)
		VALUES ('test-user', 'test.action', 'form', 999999, '{}'::jsonb, now())
		RETURNING fal_id
	`).Scan(&auditID)
	if err != nil {
		t.Fatalf("failed to insert test audit entry: %v", err)
	}

	// Attempt UPDATE — should raise exception
	_, err = testPool.Exec(ctx, `
		UPDATE eamform_audit_log SET fal_note = 'modified' WHERE fal_id = $1
	`, auditID)
	if err == nil {
		t.Fatal("expected UPDATE on eamform_audit_log to raise exception, but it succeeded")
	}

	// Clean up (this will also fail due to trigger, so we use a workaround)
	// The DELETE trigger will also block, so we disable triggers temporarily for cleanup
	_, _ = testPool.Exec(ctx, `
		ALTER TABLE eamform_audit_log DISABLE TRIGGER trg_no_delete_audit;
		DELETE FROM eamform_audit_log WHERE fal_id = $1;
		ALTER TABLE eamform_audit_log ENABLE TRIGGER trg_no_delete_audit;
	`, auditID)
}

func TestImmutability_AuditLog_DeleteRaises(t *testing.T) {
	ctx := context.Background()

	// Insert a test audit entry
	var auditID int64
	err := testPool.QueryRow(ctx, `
		INSERT INTO eamform_audit_log (fal_actor_user_id, fal_action, fal_entity_type, fal_entity_id, fal_metadata, fal_created_at)
		VALUES ('test-user', 'test.action', 'form', 999999, '{}'::jsonb, now())
		RETURNING fal_id
	`).Scan(&auditID)
	if err != nil {
		t.Fatalf("failed to insert test audit entry: %v", err)
	}

	// Attempt DELETE — should raise exception
	_, err = testPool.Exec(ctx, `
		DELETE FROM eamform_audit_log WHERE fal_id = $1
	`, auditID)
	if err == nil {
		t.Fatal("expected DELETE on eamform_audit_log to raise exception, but it succeeded")
	}

	// Clean up
	_, _ = testPool.Exec(ctx, `
		ALTER TABLE eamform_audit_log DISABLE TRIGGER trg_no_delete_audit;
		DELETE FROM eamform_audit_log WHERE fal_id = $1;
		ALTER TABLE eamform_audit_log ENABLE TRIGGER trg_no_delete_audit;
	`, auditID)
}

func TestImmutability_PublishedVersion_UpdateRaises(t *testing.T) {
	ctx := context.Background()

	// We need a form and layout to create a published version
	// Insert test data
	var formID, layoutID, versionID int64

	err := testPool.QueryRow(ctx, `
		INSERT INTO eamform_definitions (frm_key, frm_name, frm_status, frm_created_by)
		VALUES ('test-immut-form', 'Test Immutability', 'active', 'test-user')
		RETURNING frm_id
	`).Scan(&formID)
	if err != nil {
		t.Fatalf("failed to insert test form: %v", err)
	}

	err = testPool.QueryRow(ctx, `
		INSERT INTO eamform_layouts (fl_form_id, fl_name, fl_display_name, fl_status, fl_created_by)
		VALUES ($1, 'default', 'Default', 'active', 'test-user')
		RETURNING fl_id
	`, formID).Scan(&layoutID)
	if err != nil {
		t.Fatalf("failed to insert test layout: %v", err)
	}

	err = testPool.QueryRow(ctx, `
		INSERT INTO eamform_layout_versions (flv_layout_id, flv_version_number, flv_kind, flv_definition, flv_created_by)
		VALUES ($1, 1, 'published', '{}'::jsonb, 'test-user')
		RETURNING flv_id
	`, layoutID).Scan(&versionID)
	if err != nil {
		t.Fatalf("failed to insert test published version: %v", err)
	}

	// Attempt UPDATE on published version — should raise exception
	_, err = testPool.Exec(ctx, `
		UPDATE eamform_layout_versions SET flv_description = 'modified' WHERE flv_id = $1
	`, versionID)
	if err == nil {
		t.Fatal("expected UPDATE on published version to raise exception, but it succeeded")
	}

	// Clean up (disable triggers for cleanup)
	_, _ = testPool.Exec(ctx, `
		ALTER TABLE eamform_layout_versions DISABLE TRIGGER trg_no_update_versions;
		DELETE FROM eamform_layout_versions WHERE flv_id = $1;
		ALTER TABLE eamform_layout_versions ENABLE TRIGGER trg_no_update_versions;
	`, versionID)
	_, _ = testPool.Exec(ctx, `DELETE FROM eamform_layouts WHERE fl_id = $1`, layoutID)
	_, _ = testPool.Exec(ctx, `DELETE FROM eamform_definitions WHERE frm_id = $1`, formID)
}

func TestImmutability_DraftVersion_UpdateAllowed(t *testing.T) {
	ctx := context.Background()

	// Verify that draft versions CAN be updated (positive test)
	var layoutID, versionID int64

	err := testPool.QueryRow(ctx, `
		INSERT INTO eamform_layouts (fl_form_id, fl_name, fl_display_name, fl_status, fl_created_by)
		VALUES (
			(SELECT frm_id FROM eamform_definitions WHERE frm_key = 'test-immut-form' LIMIT 1),
			'draft-test-layout', 'Draft Test', 'active', 'test-user'
		)
		RETURNING fl_id
	`).Scan(&layoutID)
	if err != nil {
		// Form might not exist from previous test cleanup; skip
		t.Skip("skipping: test form not available")
	}

	err = testPool.QueryRow(ctx, `
		INSERT INTO eamform_layout_versions (flv_layout_id, flv_version_number, flv_kind, flv_definition, flv_created_by)
		VALUES ($1, 1, 'draft', '{}'::jsonb, 'test-user')
		RETURNING flv_id
	`, layoutID).Scan(&versionID)
	if err != nil {
		t.Fatalf("failed to insert test draft version: %v", err)
	}

	// UPDATE on draft version — should succeed
	_, err = testPool.Exec(ctx, `
		UPDATE eamform_layout_versions SET flv_definition = '{"updated":true}'::jsonb WHERE flv_id = $1
	`, versionID)
	if err != nil {
		t.Fatalf("expected UPDATE on draft version to succeed, but got error: %v", err)
	}

	// Clean up
	_, _ = testPool.Exec(ctx, `
		ALTER TABLE eamform_layout_versions DISABLE TRIGGER trg_no_update_versions;
		DELETE FROM eamform_layout_versions WHERE flv_id = $1;
		ALTER TABLE eamform_layout_versions ENABLE TRIGGER trg_no_update_versions;
	`, versionID)
	_, _ = testPool.Exec(ctx, `DELETE FROM eamform_layouts WHERE fl_id = $1`, layoutID)
}

// Ensure the db package import is used (for RunInTenantTx reference)
var _ = db.TenantContextKey{}
