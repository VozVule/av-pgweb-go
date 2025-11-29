SET search_path TO company;

-- Create core tables inside the company schema
CREATE TABLE IF NOT EXISTS departments (
    department_id SERIAL PRIMARY KEY,
    name          TEXT        NOT NULL UNIQUE,
    slug          TEXT        NOT NULL UNIQUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS employees (
    employee_id   SERIAL PRIMARY KEY,
    department_id INTEGER     NOT NULL REFERENCES departments(department_id) ON DELETE RESTRICT,
    full_name     TEXT        NOT NULL,
    email         TEXT        NOT NULL UNIQUE,
    hired_at      DATE        NOT NULL DEFAULT CURRENT_DATE
);

CREATE TABLE IF NOT EXISTS projects (
    project_id    SERIAL PRIMARY KEY,
    department_id INTEGER NOT NULL REFERENCES departments(department_id) ON DELETE CASCADE,
    name          TEXT    NOT NULL,
    code          TEXT    NOT NULL UNIQUE,
    lead_id       INTEGER REFERENCES employees(employee_id) ON DELETE SET NULL,
    starts_on     DATE    NOT NULL,
    ends_on       DATE
);

CREATE TABLE IF NOT EXISTS project_assignments (
    project_id   INTEGER NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    employee_id  INTEGER NOT NULL REFERENCES employees(employee_id) ON DELETE CASCADE,
    assigned_on  DATE    NOT NULL DEFAULT CURRENT_DATE,
    role         TEXT    NOT NULL DEFAULT 'member',
    PRIMARY KEY (project_id, employee_id)
);
