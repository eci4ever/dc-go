package academic

type CreateStudentRequest struct {
	StudentNo string `json:"studentNo" validate:"required,min=2,max=50"`
	Name      string `json:"name" validate:"required,min=2,max=150"`
	Email     string `json:"email" validate:"omitempty,email,max=255"`
	Program   string `json:"program" validate:"required,min=2,max=150"`
	Intake    string `json:"intake" validate:"required,min=2,max=50"`
}

type CreateSemesterRequest struct {
	Code         string `json:"code" validate:"required,min=2,max=30"`
	Name         string `json:"name" validate:"required,min=2,max=100"`
	AcademicYear int32  `json:"academicYear" validate:"required,min=2000,max=2200"`
	Sequence     int32  `json:"sequence" validate:"required,min=1,max=20"`
	Status       string `json:"status" validate:"required,oneof=planned active closed"`
}

type CreateCourseRequest struct {
	Code    string  `json:"code" validate:"required,min=2,max=30"`
	Name    string  `json:"name" validate:"required,min=2,max=150"`
	Credits float64 `json:"credits" validate:"required,gt=0,lte=30"`
}

type UpsertResultRequest struct {
	StudentID  string  `json:"studentId" validate:"required"`
	SemesterID string  `json:"semesterId" validate:"required"`
	CourseID   string  `json:"courseId" validate:"required"`
	Score      float64 `json:"score" validate:"gte=0,lte=100"`
}
