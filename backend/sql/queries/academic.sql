-- name: GetAcademicMember :one
SELECT * FROM member WHERE organization_id = $1 AND user_id = $2;

-- name: GetAcademicUserRole :one
SELECT role FROM "user" WHERE id = $1;

-- name: GetAcademicOrganizationStatus :one
SELECT status FROM organization WHERE id = $1;

-- name: CreateAcademicStudent :one
INSERT INTO academic_student (id, organization_id, student_no, name, email, program, intake)
VALUES (
    sqlc.arg(id), sqlc.arg(organization_id), sqlc.arg(student_no), sqlc.arg(name),
    NULLIF(sqlc.arg(email)::TEXT, ''), sqlc.arg(program), sqlc.arg(intake)
)
RETURNING *;

-- name: ListAcademicStudents :many
SELECT * FROM academic_student WHERE organization_id = $1 ORDER BY name, student_no;

-- name: GetAcademicStudent :one
SELECT * FROM academic_student WHERE id = $1 AND organization_id = $2;

-- name: CreateAcademicSemester :one
INSERT INTO academic_semester (id, organization_id, code, name, academic_year, sequence, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAcademicSemesters :many
SELECT * FROM academic_semester WHERE organization_id = $1 ORDER BY academic_year DESC, sequence DESC;

-- name: GetAcademicSemester :one
SELECT * FROM academic_semester WHERE id = $1 AND organization_id = $2;

-- name: CreateAcademicCourse :one
INSERT INTO academic_course (id, organization_id, code, name, credits)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListAcademicCourses :many
SELECT * FROM academic_course WHERE organization_id = $1 ORDER BY code;

-- name: GetAcademicCourse :one
SELECT * FROM academic_course WHERE id = $1 AND organization_id = $2;

-- name: ListAcademicGradeScale :many
SELECT * FROM academic_grade_scale WHERE organization_id = $1 ORDER BY sort_order;

-- name: FindAcademicGrade :one
SELECT * FROM academic_grade_scale
WHERE organization_id = $1 AND $2 BETWEEN min_score AND max_score
ORDER BY sort_order LIMIT 1;

-- name: UpsertAcademicResult :one
INSERT INTO academic_result
    (id, organization_id, student_id, semester_id, course_id, score, grade, grade_point, credits)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (student_id, semester_id, course_id) DO UPDATE SET
    score = EXCLUDED.score,
    grade = EXCLUDED.grade,
    grade_point = EXCLUDED.grade_point,
    credits = EXCLUDED.credits,
    updated_at = NOW()
RETURNING *;

-- name: ListAcademicResultsByStudent :many
SELECT
    r.*,
    s.code AS semester_code,
    s.name AS semester_name,
    s.academic_year,
    s.sequence AS semester_sequence,
    c.code AS course_code,
    c.name AS course_name
FROM academic_result r
JOIN academic_semester s ON s.id = r.semester_id
JOIN academic_course c ON c.id = r.course_id
WHERE r.organization_id = $1 AND r.student_id = $2
ORDER BY s.academic_year, s.sequence, c.code;
