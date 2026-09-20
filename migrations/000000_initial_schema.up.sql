CREATE SEQUENCE employees_id_seq;

CREATE TABLE employees (
    id BIGINT NOT NULL DEFAULT nextval('employees_id_seq'),
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    department TEXT,
    CONSTRAINT employees_pkey PRIMARY KEY (id),
    CONSTRAINT uni_employees_email UNIQUE (email)
);

ALTER SEQUENCE employees_id_seq
    OWNED BY employees.id;


CREATE SEQUENCE projects_id_seq;

CREATE TABLE projects (
    id BIGINT NOT NULL DEFAULT nextval('projects_id_seq'),
    name TEXT NOT NULL,
    description TEXT,
    CONSTRAINT projects_pkey PRIMARY KEY (id)
);

ALTER SEQUENCE projects_id_seq
    OWNED BY projects.id;


CREATE SEQUENCE tasks_id_seq;

CREATE TABLE tasks (
    id BIGINT NOT NULL DEFAULT nextval('tasks_id_seq'),
    project_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    CONSTRAINT tasks_pkey PRIMARY KEY (id),
    CONSTRAINT fk_projects_tasks
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
);

ALTER SEQUENCE tasks_id_seq
    OWNED BY tasks.id;
