CREATE TABLE academic_student (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    student_no TEXT NOT NULL,
    name TEXT NOT NULL,
    email TEXT,
    program TEXT NOT NULL,
    intake TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'graduated', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, student_no),
    UNIQUE (id, organization_id)
);

CREATE TABLE academic_semester (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    academic_year INTEGER NOT NULL CHECK (academic_year BETWEEN 2000 AND 2200),
    sequence INTEGER NOT NULL CHECK (sequence BETWEEN 1 AND 20),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('planned', 'active', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, code),
    UNIQUE (id, organization_id)
);

CREATE TABLE academic_course (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    credits DOUBLE PRECISION NOT NULL CHECK (credits > 0 AND credits <= 30),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, code),
    UNIQUE (id, organization_id)
);

CREATE TABLE academic_grade_scale (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    letter TEXT NOT NULL,
    min_score DOUBLE PRECISION NOT NULL CHECK (min_score >= 0 AND min_score <= 100),
    max_score DOUBLE PRECISION NOT NULL CHECK (max_score >= 0 AND max_score <= 100),
    grade_point DOUBLE PRECISION NOT NULL CHECK (grade_point >= 0 AND grade_point <= 4),
    passing BOOLEAN NOT NULL,
    sort_order INTEGER NOT NULL,
    CHECK (min_score <= max_score),
    UNIQUE (organization_id, letter),
    UNIQUE (organization_id, sort_order)
);

CREATE TABLE academic_result (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    student_id TEXT NOT NULL,
    semester_id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    score DOUBLE PRECISION NOT NULL CHECK (score >= 0 AND score <= 100),
    grade TEXT NOT NULL,
    grade_point DOUBLE PRECISION NOT NULL CHECK (grade_point >= 0 AND grade_point <= 4),
    credits DOUBLE PRECISION NOT NULL CHECK (credits > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (student_id, organization_id) REFERENCES academic_student(id, organization_id) ON DELETE CASCADE,
    FOREIGN KEY (semester_id, organization_id) REFERENCES academic_semester(id, organization_id) ON DELETE CASCADE,
    FOREIGN KEY (course_id, organization_id) REFERENCES academic_course(id, organization_id) ON DELETE RESTRICT,
    UNIQUE (student_id, semester_id, course_id)
);

CREATE INDEX academic_student_org_idx ON academic_student(organization_id, name);
CREATE INDEX academic_semester_org_idx ON academic_semester(organization_id, academic_year, sequence);
CREATE INDEX academic_course_org_idx ON academic_course(organization_id, code);
CREATE INDEX academic_result_student_idx ON academic_result(student_id, semester_id);

CREATE OR REPLACE FUNCTION seed_academic_grade_scale(target_organization_id TEXT)
RETURNS VOID AS $$
BEGIN
    INSERT INTO academic_grade_scale
        (id, organization_id, letter, min_score, max_score, grade_point, passing, sort_order)
    VALUES
        (gen_random_uuid()::TEXT, target_organization_id, 'A',  80, 100, 4.00, true,  1),
        (gen_random_uuid()::TEXT, target_organization_id, 'A-', 75, 79.99, 3.67, true,  2),
        (gen_random_uuid()::TEXT, target_organization_id, 'B+', 70, 74.99, 3.33, true,  3),
        (gen_random_uuid()::TEXT, target_organization_id, 'B',  65, 69.99, 3.00, true,  4),
        (gen_random_uuid()::TEXT, target_organization_id, 'B-', 60, 64.99, 2.67, true,  5),
        (gen_random_uuid()::TEXT, target_organization_id, 'C+', 55, 59.99, 2.33, true,  6),
        (gen_random_uuid()::TEXT, target_organization_id, 'C',  50, 54.99, 2.00, true,  7),
        (gen_random_uuid()::TEXT, target_organization_id, 'D',  40, 49.99, 1.00, true,  8),
        (gen_random_uuid()::TEXT, target_organization_id, 'F',   0, 39.99, 0.00, false, 9)
    ON CONFLICT (organization_id, letter) DO NOTHING;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION seed_academic_grade_scale_for_organization()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM seed_academic_grade_scale(NEW.id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER organization_academic_grade_scale_trigger
AFTER INSERT ON organization
FOR EACH ROW EXECUTE FUNCTION seed_academic_grade_scale_for_organization();

SELECT seed_academic_grade_scale(id) FROM organization;
