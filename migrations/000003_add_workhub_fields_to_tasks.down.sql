DROP INDEX IF EXISTS idx_tasks_deleted_at;

ALTER TABLE tasks
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at;

