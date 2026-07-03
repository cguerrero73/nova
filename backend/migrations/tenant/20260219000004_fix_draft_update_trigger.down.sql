-- Rollback: restore the original statement-level immutability trigger
DROP TRIGGER IF EXISTS trg_no_update_versions ON eamform_layout_versions;

CREATE TRIGGER trg_no_update_versions
    BEFORE UPDATE ON eamform_layout_versions
    FOR EACH STATEMENT
    EXECUTE FUNCTION raise_immutable();

DROP FUNCTION IF EXISTS check_version_update();
