-- Nova EAM - Dynamic Form Builder: Fix draft update trigger
-- Timestamp: 20260219000004
-- Description: The original trg_no_update_versions blocked ALL updates on
--   eamform_layout_versions, including draft mutations needed by SaveDraft.
--   This migration replaces the statement-level trigger with a row-level
--   trigger that allows UPDATE only when the row being modified is a draft
--   (OLD.flv_kind = 'draft'). Published and archived versions remain immutable.

-- ============================================
-- New function: allow draft updates, block others
-- ============================================
CREATE OR REPLACE FUNCTION check_version_update() RETURNS trigger AS $$
BEGIN
    -- Allow updates on draft rows (SaveDraft, kind transition during publish)
    IF OLD.flv_kind = 'draft' THEN
        RETURN NEW;
    END IF;
    -- Block updates on published and archived rows (immutable snapshots)
    RAISE EXCEPTION 'cannot update non-draft version (id=%, kind=%)', OLD.flv_id, OLD.flv_kind;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Replace the old statement-level trigger with row-level
-- ============================================
DROP TRIGGER IF EXISTS trg_no_update_versions ON eamform_layout_versions;

CREATE TRIGGER trg_no_update_versions
    BEFORE UPDATE ON eamform_layout_versions
    FOR EACH ROW
    EXECUTE FUNCTION check_version_update();
