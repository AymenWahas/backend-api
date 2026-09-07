ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS fk_projects_owner;

DROP INDEX IF EXISTS idx_projects_deleted_at;
DROP INDEX IF EXISTS idx_projects_owner_id;

ALTER TABLE projects
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS owner_id;
