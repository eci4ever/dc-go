package academic

type Student struct {
	ID        string  `json:"id"`
	OrgID     string  `json:"organizationId"`
	StudentNo string  `json:"studentNo"`
	Name      string  `json:"name"`
	Email     *string `json:"email"`
	Program   string  `json:"program"`
	Intake    string  `json:"intake"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"createdAt"`
}

type Semester struct {
	ID           string `json:"id"`
	OrgID        string `json:"organizationId"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	AcademicYear int32  `json:"academicYear"`
	Sequence     int32  `json:"sequence"`
	Status       string `json:"status"`
}

type Course struct {
	ID      string  `json:"id"`
	OrgID   string  `json:"organizationId"`
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Credits float64 `json:"credits"`
	Active  bool    `json:"active"`
}

type GradeScale struct {
	Letter     string  `json:"letter"`
	MinScore   float64 `json:"minScore"`
	MaxScore   float64 `json:"maxScore"`
	GradePoint float64 `json:"gradePoint"`
	Passing    bool    `json:"passing"`
}

type Result struct {
	ID           string  `json:"id"`
	StudentID    string  `json:"studentId"`
	SemesterID   string  `json:"semesterId"`
	CourseID     string  `json:"courseId"`
	CourseCode   string  `json:"courseCode,omitempty"`
	CourseName   string  `json:"courseName,omitempty"`
	Score        float64 `json:"score"`
	Grade        string  `json:"grade"`
	GradePoint   float64 `json:"gradePoint"`
	Credits      float64 `json:"credits"`
	QualityPoint float64 `json:"qualityPoint"`
}

type SemesterResult struct {
	SemesterID   string   `json:"semesterId"`
	SemesterCode string   `json:"semesterCode"`
	SemesterName string   `json:"semesterName"`
	AcademicYear int32    `json:"academicYear"`
	Sequence     int32    `json:"sequence"`
	Results      []Result `json:"results"`
	TotalCredits float64  `json:"totalCredits"`
	TotalPoints  float64  `json:"totalPoints"`
	GPA          float64  `json:"gpa"`
	CGPA         float64  `json:"cgpa"`
}

type Transcript struct {
	Student   Student          `json:"student"`
	Semesters []SemesterResult `json:"semesters"`
	CGPA      float64          `json:"cgpa"`
}

const time3339 = "2006-01-02T15:04:05Z07:00"
