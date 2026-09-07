DROP INDEX IF EXISTS idx_employees_deleted_at;

ALTER TABLE employees
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS created_at;