SET search_path TO company;

-- Seed sample data
INSERT INTO departments (name, slug)
VALUES
    ('Engineering', 'engineering'),
    ('Product', 'product'),
    ('Operations', 'operations')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO employees (department_id, full_name, email)
SELECT d.department_id, v.full_name, v.email
FROM (
    VALUES
        ('engineering', 'Ada Lovelace', 'ada@example.com'),
        ('engineering', 'Linus Torvalds', 'linus@example.com'),
        ('product', 'Grace Hopper', 'grace@example.com'),
        ('operations', 'Edsger Dijkstra', 'edsger@example.com')
) AS v(slug, full_name, email)
JOIN departments d ON d.slug = v.slug
ON CONFLICT (email) DO NOTHING;

INSERT INTO projects (department_id, name, code, lead_id, starts_on, ends_on)
SELECT d.department_id,
       v.name,
       v.code,
       e.employee_id,
       v.starts_on,
       v.ends_on
FROM (
    VALUES
        ('engineering', 'Telemetry Platform', 'ENG-TEL-001', 'linus@example.com', '2025-01-01', NULL),
        ('product', 'Mobile Redesign', 'PRO-MOB-002', 'grace@example.com', '2025-02-15', '2025-08-30')
) AS v(slug, name, code, lead_email, starts_on, ends_on)
JOIN departments d ON d.slug = v.slug
LEFT JOIN employees e ON e.email = v.lead_email
ON CONFLICT (code) DO NOTHING;

INSERT INTO project_assignments (project_id, employee_id, role)
SELECT p.project_id, e.employee_id, v.role
FROM (
    VALUES
        ('ENG-TEL-001', 'linus@example.com', 'lead'),
        ('ENG-TEL-001', 'ada@example.com', 'member'),
        ('PRO-MOB-002', 'grace@example.com', 'lead'),
        ('PRO-MOB-002', 'ada@example.com', 'advisor')
) AS v(code, email, role)
JOIN projects p ON p.code = v.code
JOIN employees e ON e.email = v.email
ON CONFLICT (project_id, employee_id) DO NOTHING;
