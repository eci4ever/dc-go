package academic

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, authMw, csrfMw fiber.Handler) {
	g := router.Group("/organizations/:orgID/academic", authMw)
	g.Get("/students", h.ListStudents)
	g.Post("/students", csrfMw, h.CreateStudent)
	g.Get("/semesters", h.ListSemesters)
	g.Post("/semesters", csrfMw, h.CreateSemester)
	g.Get("/courses", h.ListCourses)
	g.Post("/courses", csrfMw, h.CreateCourse)
	g.Get("/grade-scale", h.ListGradeScale)
	g.Put("/results", csrfMw, h.UpsertResult)
	g.Get("/students/:studentID/transcript", h.Transcript)
}
