package academic

import (
	"testing"

	db "github.com/eci4ever/dc-go/internal/db"
)

func TestCalculateTranscript(t *testing.T) {
	student := Student{ID: "student-1", Name: "Ali"}
	rows := []db.ListAcademicResultsByStudentRow{
		{ID: "r1", StudentID: student.ID, SemesterID: "s1", SemesterCode: "2026-1", SemesterName: "Semester 1", AcademicYear: 2026, SemesterSequence: 1, CourseID: "c1", CourseCode: "MAT101", CourseName: "Math", Credits: 3, GradePoint: 4, Grade: "A", Score: 85},
		{ID: "r2", StudentID: student.ID, SemesterID: "s1", SemesterCode: "2026-1", SemesterName: "Semester 1", AcademicYear: 2026, SemesterSequence: 1, CourseID: "c2", CourseCode: "ENG101", CourseName: "English", Credits: 2, GradePoint: 3, Grade: "B", Score: 67},
		{ID: "r3", StudentID: student.ID, SemesterID: "s2", SemesterCode: "2026-2", SemesterName: "Semester 2", AcademicYear: 2026, SemesterSequence: 2, CourseID: "c3", CourseCode: "TEC101", CourseName: "Technology", Credits: 3, GradePoint: 2, Grade: "C", Score: 52},
	}

	got := calculateTranscript(student, rows)
	if len(got.Semesters) != 2 {
		t.Fatalf("semesters = %d, want 2", len(got.Semesters))
	}
	if got.Semesters[0].GPA != 3.6 {
		t.Errorf("semester 1 GPA = %.2f, want 3.60", got.Semesters[0].GPA)
	}
	if got.Semesters[1].GPA != 2.0 {
		t.Errorf("semester 2 GPA = %.2f, want 2.00", got.Semesters[1].GPA)
	}
	if got.CGPA != 3.0 {
		t.Errorf("CGPA = %.2f, want 3.00", got.CGPA)
	}
}
