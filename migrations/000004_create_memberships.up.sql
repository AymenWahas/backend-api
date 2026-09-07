CREATE TABLE IF NOT EXISTS memberships (
    employee_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (employee_id, project_id),

    CONSTRAINT fk_memberships_employee
        FOREIGN KEY (employee_id)
        REFERENCES employees(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_memberships_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE
);

