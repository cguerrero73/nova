-- Nova EAM - Dynamic Form Builder: Core Tables
-- Timestamp: 20260219000001
-- Description: Form definitions, layouts, versions, role assignments, audit log.
--              No tenant_id columns — isolation via search_path (conventions §1).

-- ============================================
-- Helper: raise_immutable()
-- ============================================
CREATE OR REPLACE FUNCTION raise_immutable() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'table % is immutable — rows cannot be updated or deleted', TG_TABLE_NAME;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Helper: check_default_layout_exists()
-- Rejects DELETE on eamform_definitions if a layout named 'default' still
-- references the form (prevents orphaned default layouts).
-- ============================================
CREATE OR REPLACE FUNCTION check_default_layout_exists() RETURNS trigger AS $$
DECLARE
    v_count INT;
BEGIN
    SELECT COUNT(*) INTO v_count
    FROM eamform_layouts
    WHERE fl_form_id = OLD.frm_id
      AND fl_name = 'default';

    IF v_count > 0 THEN
        RAISE EXCEPTION 'cannot delete form %: default layout still references it', OLD.frm_key;
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- eamform_definitions
-- ============================================
CREATE TABLE eamform_definitions (
    frm_id          BIGSERIAL PRIMARY KEY,
    frm_key         VARCHAR(100) NOT NULL,
    frm_name        VARCHAR(255) NOT NULL,
    frm_description TEXT,
    frm_status      VARCHAR(20) NOT NULL DEFAULT 'active',
    frm_created_by  VARCHAR(50),
    frm_created_at  TIMESTAMP NOT NULL DEFAULT now(),
    frm_updated_at  TIMESTAMP NOT NULL DEFAULT now(),

    CONSTRAINT chk_frm_status CHECK (frm_status IN ('active', 'archived'))
);

CREATE UNIQUE INDEX uq_frm_key ON eamform_definitions(frm_key);
CREATE INDEX idx_frm_status ON eamform_definitions(frm_status);

COMMENT ON TABLE eamform_definitions IS 'Logical form definitions per tenant';
COMMENT ON COLUMN eamform_definitions.frm_key IS 'Unique form identifier (slug)';
COMMENT ON COLUMN eamform_definitions.frm_status IS 'active | archived';

-- ============================================
-- eamform_layouts
-- ============================================
CREATE TABLE eamform_layouts (
    fl_id                   BIGSERIAL PRIMARY KEY,
    fl_form_id              BIGINT NOT NULL REFERENCES eamform_definitions(frm_id),
    fl_name                 VARCHAR(100) NOT NULL,
    fl_display_name         VARCHAR(255) NOT NULL,
    fl_description          TEXT,
    fl_status               VARCHAR(20) NOT NULL DEFAULT 'active',
    fl_draft_version_id     BIGINT,
    fl_published_version_id BIGINT,
    fl_created_by           VARCHAR(50),
    fl_created_at           TIMESTAMP NOT NULL DEFAULT now(),
    fl_updated_at           TIMESTAMP NOT NULL DEFAULT now(),

    CONSTRAINT chk_fl_status CHECK (fl_status IN ('active', 'archived'))
);

CREATE UNIQUE INDEX uq_fl_form_name ON eamform_layouts(fl_form_id, fl_name);
CREATE INDEX idx_fl_form_id ON eamform_layouts(fl_form_id);

COMMENT ON TABLE eamform_layouts IS 'Named layouts per form (role audience variants)';
COMMENT ON COLUMN eamform_layouts.fl_name IS 'Layout slug; "default" is reserved and auto-created';
COMMENT ON COLUMN eamform_layouts.fl_draft_version_id IS 'Points to current draft version (NULL if none)';
COMMENT ON COLUMN eamform_layouts.fl_published_version_id IS 'Points to current published version (NULL if never published)';

-- ============================================
-- eamform_layout_versions
-- ============================================
CREATE TABLE eamform_layout_versions (
    flv_id            BIGSERIAL PRIMARY KEY,
    flv_layout_id     BIGINT NOT NULL REFERENCES eamform_layouts(fl_id),
    flv_version_number INT NOT NULL,
    flv_kind          VARCHAR(20) NOT NULL,
    flv_description   TEXT,
    flv_definition    JSONB NOT NULL DEFAULT '{}'::jsonb,
    flv_created_by    VARCHAR(50),
    flv_created_at    TIMESTAMP NOT NULL DEFAULT now(),
    flv_published_at  TIMESTAMP,

    CONSTRAINT chk_flv_kind CHECK (flv_kind IN ('draft', 'published', 'archived'))
);

CREATE INDEX idx_flv_layout_id ON eamform_layout_versions(flv_layout_id);
CREATE INDEX idx_flv_layout_kind ON eamform_layout_versions(flv_layout_id, flv_kind);

-- Partial unique: at most one draft per layout
CREATE UNIQUE INDEX uq_flv_one_draft ON eamform_layout_versions(flv_layout_id)
    WHERE flv_kind = 'draft';

COMMENT ON TABLE eamform_layout_versions IS 'Append-only version snapshots per layout';
COMMENT ON COLUMN eamform_layout_versions.flv_kind IS 'draft | published | archived';
COMMENT ON COLUMN eamform_layout_versions.flv_definition IS 'Complete layout JSON snapshot';

-- Immutability: no UPDATE on versions
CREATE TRIGGER trg_no_update_versions
    BEFORE UPDATE ON eamform_layout_versions
    FOR EACH STATEMENT
    EXECUTE FUNCTION raise_immutable();

-- ============================================
-- Wire draft/published pointers
-- ============================================
ALTER TABLE eamform_layouts
    ADD CONSTRAINT fk_fl_draft_version
    FOREIGN KEY (fl_draft_version_id) REFERENCES eamform_layout_versions(flv_id);

ALTER TABLE eamform_layouts
    ADD CONSTRAINT fk_fl_published_version
    FOREIGN KEY (fl_published_version_id) REFERENCES eamform_layout_versions(flv_id);

-- ============================================
-- eamform_role_assignments
-- ============================================
CREATE TABLE eamform_role_assignments (
    fra_id          BIGSERIAL PRIMARY KEY,
    fra_form_id     BIGINT NOT NULL REFERENCES eamform_definitions(frm_id),
    fra_layout_id   BIGINT NOT NULL REFERENCES eamform_layouts(fl_id),
    fra_role_name   VARCHAR(50) NOT NULL,
    fra_assigned_at TIMESTAMP NOT NULL DEFAULT now(),
    fra_revoked_at  TIMESTAMP,
    fra_assigned_by VARCHAR(50)
);

CREATE INDEX idx_fra_form_id ON eamform_role_assignments(fra_form_id);
CREATE INDEX idx_fra_form_role ON eamform_role_assignments(fra_form_id, fra_role_name);

-- Partial unique: at most one active assignment per (form, role)
CREATE UNIQUE INDEX uq_fra_active ON eamform_role_assignments(fra_form_id, fra_role_name)
    WHERE fra_revoked_at IS NULL;

COMMENT ON TABLE eamform_role_assignments IS 'Maps tenant roles to specific layouts per form';
COMMENT ON COLUMN eamform_role_assignments.fra_revoked_at IS 'NULL = active; set = revoked';

-- ============================================
-- eamform_audit_log
-- ============================================
CREATE TABLE eamform_audit_log (
    fal_id            BIGSERIAL PRIMARY KEY,
    fal_actor_user_id VARCHAR(50) NOT NULL,
    fal_action        VARCHAR(50) NOT NULL,
    fal_entity_type   VARCHAR(50) NOT NULL,
    fal_entity_id     BIGINT NOT NULL,
    fal_metadata      JSONB NOT NULL DEFAULT '{}'::jsonb,
    fal_note          TEXT,
    fal_created_at    TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_fal_entity ON eamform_audit_log(fal_entity_type, fal_entity_id);
CREATE INDEX idx_fal_action ON eamform_audit_log(fal_action);
CREATE INDEX idx_fal_created ON eamform_audit_log(fal_created_at);

COMMENT ON TABLE eamform_audit_log IS 'Append-only audit log for all form builder mutations';
COMMENT ON COLUMN eamform_audit_log.fal_action IS 'form.create, form.archive, layout.create, layout.archive, layout.assign, layout.unassign, version.publish, version.revert, version.draft_save';

-- Immutability: no UPDATE on audit
CREATE TRIGGER trg_no_update_audit
    BEFORE UPDATE ON eamform_audit_log
    FOR EACH STATEMENT
    EXECUTE FUNCTION raise_immutable();

-- Immutability: no DELETE on audit
CREATE TRIGGER trg_no_delete_audit
    BEFORE DELETE ON eamform_audit_log
    FOR EACH STATEMENT
    EXECUTE FUNCTION raise_immutable();

-- ============================================
-- Protect default layout from orphaned form deletion
-- ============================================
CREATE TRIGGER trg_protect_default_layout
    BEFORE DELETE ON eamform_definitions
    FOR EACH ROW
    EXECUTE FUNCTION check_default_layout_exists();
