package academic

import (
	"context"
	"math"

	db "github.com/eci4ever/dc-go/internal/db"
)

const (
	permissionStudents  = "academic.students.manage"
	permissionStructure = "academic.structure.manage"
	permissionResults   = "academic.results.manage"
)

type Service struct{ repo *Repository }

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

func (s *Service) CreateStudent(ctx context.Context, orgID, actorID string, req CreateStudentRequest) (Student, error) {
	if err := s.authorize(ctx, orgID, actorID, permissionStudents, true); err != nil {
		return Student{}, err
	}
	return s.repo.CreateStudent(ctx, orgID, req)
}

func (s *Service) ListStudents(ctx context.Context, orgID, actorID string) ([]Student, error) {
	if err := s.authorize(ctx, orgID, actorID, permissionStudents, false); err != nil {
		return nil, err
	}
	return s.repo.ListStudents(ctx, orgID)
}

func (s *Service) CreateSemester(ctx context.Context, orgID, actorID string, req CreateSemesterRequest) (Semester, error) {
	if err := s.authorize(ctx, orgID, actorID, permissionStructure, true); err != nil {
		return Semester{}, err
	}
	return s.repo.CreateSemester(ctx, orgID, req)
}

func (s *Service) ListSemesters(ctx context.Context, orgID, actorID string) ([]Semester, error) {
	if err := s.authorize(ctx, orgID, actorID, permissionStructure, false); err != nil {
		return nil, err
	}
	return s.repo.ListSemesters(ctx, orgID)
}

func (s *Service) CreateCourse(ctx context.Context, orgID, actorID string, req CreateCourseRequest) (Course, error) {
	if err := s.authorize(ctx, orgID, actorID, permissionStructure, true); err != nil {
		return Course{}, err
	}
	return s.repo.CreateCourse(ctx, orgID, req)
}

func (s *Service) ListCourses(ctx context.Context, orgID, actorID string) ([]Course, error) {
	if err := s.authorize(ctx, orgID, actorID, permissionStructure, false); err != nil {
		return nil, err
	}
	return s.repo.ListCourses(ctx, orgID)
}

func (s *Service) ListGradeScale(ctx context.Context, orgID, actorID string) ([]GradeScale, error) {
	if err := s.authorize(ctx, orgID, actorID, permissionStructure, false); err != nil {
		return nil, err
	}
	return s.repo.ListGradeScale(ctx, orgID)
}

func (s *Service) UpsertResult(ctx context.Context, orgID, actorID string, req UpsertResultRequest) (Result, error) {
	if err := s.authorize(ctx, orgID, actorID, permissionResults, true); err != nil {
		return Result{}, err
	}
	if _, err := s.repo.GetStudent(ctx, orgID, req.StudentID); err != nil {
		return Result{}, err
	}
	if _, err := s.repo.GetSemester(ctx, orgID, req.SemesterID); err != nil {
		return Result{}, err
	}
	course, err := s.repo.GetCourse(ctx, orgID, req.CourseID)
	if err != nil {
		return Result{}, err
	}
	grade, err := s.repo.GradeForScore(ctx, orgID, req.Score)
	if err != nil {
		return Result{}, err
	}
	return s.repo.UpsertResult(ctx, orgID, req, grade, course.Credits)
}

func (s *Service) Transcript(ctx context.Context, orgID, studentID, actorID string) (Transcript, error) {
	if err := s.authorize(ctx, orgID, actorID, permissionResults, false); err != nil {
		return Transcript{}, err
	}
	student, err := s.repo.GetStudent(ctx, orgID, studentID)
	if err != nil {
		return Transcript{}, err
	}
	rows, err := s.repo.ListResults(ctx, orgID, studentID)
	if err != nil {
		return Transcript{}, err
	}
	return calculateTranscript(student, rows), nil
}

func (s *Service) manager(ctx context.Context, orgID, actorID string) error {
	return s.authorize(ctx, orgID, actorID, permissionResults, false)
}

func (s *Service) authorize(ctx context.Context, orgID, actorID, permission string, write bool) error {
	if write {
		status, err := s.repo.OrganizationStatus(ctx, orgID)
		if err != nil {
			return err
		}
		if status != "active" {
			return ErrOrganizationLocked
		}
	}
	globalRole, err := s.repo.GlobalUserRole(ctx, actorID)
	if err != nil {
		return err
	}
	if globalRole == "admin" {
		return nil
	}
	access, err := s.repo.MemberAccess(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if access.Role == "owner" || access.Role == "admin" {
		return nil
	}
	for _, granted := range access.Permissions {
		if granted == permission {
			return nil
		}
		if !write && granted == permissionResults &&
			(permission == permissionStudents || permission == permissionStructure) {
			return nil
		}
	}
	return ErrForbidden
}

func calculateTranscript(student Student, rows []db.ListAcademicResultsByStudentRow) Transcript {
	out := Transcript{Student: student, Semesters: []SemesterResult{}}
	var cumulativeCredits, cumulativePoints float64
	for _, row := range rows {
		if len(out.Semesters) == 0 || out.Semesters[len(out.Semesters)-1].SemesterID != row.SemesterID {
			out.Semesters = append(out.Semesters, SemesterResult{SemesterID: row.SemesterID, SemesterCode: row.SemesterCode, SemesterName: row.SemesterName, AcademicYear: row.AcademicYear, Sequence: row.SemesterSequence, Results: []Result{}})
		}
		semester := &out.Semesters[len(out.Semesters)-1]
		qualityPoint := row.Credits * row.GradePoint
		semester.Results = append(semester.Results, Result{ID: row.ID, StudentID: row.StudentID, SemesterID: row.SemesterID, CourseID: row.CourseID, CourseCode: row.CourseCode, CourseName: row.CourseName, Score: row.Score, Grade: row.Grade, GradePoint: row.GradePoint, Credits: row.Credits, QualityPoint: round2(qualityPoint)})
		semester.TotalCredits += row.Credits
		semester.TotalPoints += qualityPoint
		cumulativeCredits += row.Credits
		cumulativePoints += qualityPoint
		semester.GPA = divide(semester.TotalPoints, semester.TotalCredits)
		semester.CGPA = divide(cumulativePoints, cumulativeCredits)
	}
	if len(out.Semesters) > 0 {
		out.CGPA = out.Semesters[len(out.Semesters)-1].CGPA
	}
	for i := range out.Semesters {
		out.Semesters[i].TotalCredits = round2(out.Semesters[i].TotalCredits)
		out.Semesters[i].TotalPoints = round2(out.Semesters[i].TotalPoints)
	}
	return out
}

func divide(value, divisor float64) float64 {
	if divisor == 0 {
		return 0
	}
	return round2(value / divisor)
}

func round2(value float64) float64 { return math.Round(value*100) / 100 }
