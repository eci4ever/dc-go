package academic

import (
	"context"
	"errors"

	db "github.com/eci4ever/dc-go/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ q *db.Queries }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{q: db.New(pool)} }

func (r *Repository) MemberRole(ctx context.Context, orgID, userID string) (string, error) {
	m, err := r.q.GetAcademicMember(ctx, db.GetAcademicMemberParams{OrganizationID: orgID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return m.Role, nil
}

func (r *Repository) GlobalUserRole(ctx context.Context, userID string) (string, error) {
	return r.q.GetAcademicUserRole(ctx, userID)
}

func (r *Repository) CreateStudent(ctx context.Context, orgID string, req CreateStudentRequest) (Student, error) {
	row, err := r.q.CreateAcademicStudent(ctx, db.CreateAcademicStudentParams{ID: uuid.NewString(), OrganizationID: orgID, StudentNo: req.StudentNo, Name: req.Name, Email: req.Email, Program: req.Program, Intake: req.Intake})
	if isUnique(err) {
		return Student{}, ErrConflict
	}
	return mapStudent(row), err
}

func (r *Repository) ListStudents(ctx context.Context, orgID string) ([]Student, error) {
	rows, err := r.q.ListAcademicStudents(ctx, orgID)
	if err != nil {
		return nil, err
	}
	items := make([]Student, len(rows))
	for i, row := range rows {
		items[i] = mapStudent(row)
	}
	return items, nil
}

func (r *Repository) GetStudent(ctx context.Context, orgID, id string) (Student, error) {
	row, err := r.q.GetAcademicStudent(ctx, db.GetAcademicStudentParams{ID: id, OrganizationID: orgID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Student{}, ErrNotFound
	}
	return mapStudent(row), err
}

func (r *Repository) CreateSemester(ctx context.Context, orgID string, req CreateSemesterRequest) (Semester, error) {
	row, err := r.q.CreateAcademicSemester(ctx, db.CreateAcademicSemesterParams{ID: uuid.NewString(), OrganizationID: orgID, Code: req.Code, Name: req.Name, AcademicYear: req.AcademicYear, Sequence: req.Sequence, Status: req.Status})
	if isUnique(err) {
		return Semester{}, ErrConflict
	}
	return mapSemester(row), err
}

func (r *Repository) ListSemesters(ctx context.Context, orgID string) ([]Semester, error) {
	rows, err := r.q.ListAcademicSemesters(ctx, orgID)
	if err != nil {
		return nil, err
	}
	items := make([]Semester, len(rows))
	for i, row := range rows {
		items[i] = mapSemester(row)
	}
	return items, nil
}

func (r *Repository) GetSemester(ctx context.Context, orgID, id string) (Semester, error) {
	row, err := r.q.GetAcademicSemester(ctx, db.GetAcademicSemesterParams{ID: id, OrganizationID: orgID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Semester{}, ErrNotFound
	}
	return mapSemester(row), err
}

func (r *Repository) CreateCourse(ctx context.Context, orgID string, req CreateCourseRequest) (Course, error) {
	row, err := r.q.CreateAcademicCourse(ctx, db.CreateAcademicCourseParams{ID: uuid.NewString(), OrganizationID: orgID, Code: req.Code, Name: req.Name, Credits: req.Credits})
	if isUnique(err) {
		return Course{}, ErrConflict
	}
	return mapCourse(row), err
}

func (r *Repository) ListCourses(ctx context.Context, orgID string) ([]Course, error) {
	rows, err := r.q.ListAcademicCourses(ctx, orgID)
	if err != nil {
		return nil, err
	}
	items := make([]Course, len(rows))
	for i, row := range rows {
		items[i] = mapCourse(row)
	}
	return items, nil
}

func (r *Repository) GetCourse(ctx context.Context, orgID, id string) (Course, error) {
	row, err := r.q.GetAcademicCourse(ctx, db.GetAcademicCourseParams{ID: id, OrganizationID: orgID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Course{}, ErrNotFound
	}
	return mapCourse(row), err
}

func (r *Repository) GradeForScore(ctx context.Context, orgID string, score float64) (GradeScale, error) {
	row, err := r.q.FindAcademicGrade(ctx, db.FindAcademicGradeParams{OrganizationID: orgID, MinScore: score})
	if errors.Is(err, pgx.ErrNoRows) {
		return GradeScale{}, ErrNotFound
	}
	return mapGrade(row), err
}

func (r *Repository) ListGradeScale(ctx context.Context, orgID string) ([]GradeScale, error) {
	rows, err := r.q.ListAcademicGradeScale(ctx, orgID)
	if err != nil {
		return nil, err
	}
	items := make([]GradeScale, len(rows))
	for i, row := range rows {
		items[i] = mapGrade(row)
	}
	return items, nil
}

func (r *Repository) UpsertResult(ctx context.Context, orgID string, req UpsertResultRequest, grade GradeScale, credits float64) (Result, error) {
	row, err := r.q.UpsertAcademicResult(ctx, db.UpsertAcademicResultParams{ID: uuid.NewString(), OrganizationID: orgID, StudentID: req.StudentID, SemesterID: req.SemesterID, CourseID: req.CourseID, Score: req.Score, Grade: grade.Letter, GradePoint: grade.GradePoint, Credits: credits})
	if err != nil {
		return Result{}, err
	}
	return mapResult(row), nil
}

func (r *Repository) ListResults(ctx context.Context, orgID, studentID string) ([]db.ListAcademicResultsByStudentRow, error) {
	return r.q.ListAcademicResultsByStudent(ctx, db.ListAcademicResultsByStudentParams{OrganizationID: orgID, StudentID: studentID})
}

func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func mapStudent(row db.AcademicStudent) Student {
	var email *string
	if row.Email.Valid {
		email = &row.Email.String
	}
	return Student{ID: row.ID, OrgID: row.OrganizationID, StudentNo: row.StudentNo, Name: row.Name, Email: email, Program: row.Program, Intake: row.Intake, Status: row.Status, CreatedAt: row.CreatedAt.Time.Format(time3339)}
}

func mapSemester(row db.AcademicSemester) Semester {
	return Semester{ID: row.ID, OrgID: row.OrganizationID, Code: row.Code, Name: row.Name, AcademicYear: row.AcademicYear, Sequence: row.Sequence, Status: row.Status}
}

func mapCourse(row db.AcademicCourse) Course {
	return Course{ID: row.ID, OrgID: row.OrganizationID, Code: row.Code, Name: row.Name, Credits: row.Credits, Active: row.Active}
}

func mapGrade(row db.AcademicGradeScale) GradeScale {
	return GradeScale{Letter: row.Letter, MinScore: row.MinScore, MaxScore: row.MaxScore, GradePoint: row.GradePoint, Passing: row.Passing}
}

func mapResult(row db.AcademicResult) Result {
	return Result{ID: row.ID, StudentID: row.StudentID, SemesterID: row.SemesterID, CourseID: row.CourseID, Score: row.Score, Grade: row.Grade, GradePoint: row.GradePoint, Credits: row.Credits, QualityPoint: round2(row.Credits * row.GradePoint)}
}
