package academic

import (
	"errors"

	"github.com/eci4ever/dc-go/pkg/response"
	"github.com/eci4ever/dc-go/pkg/validator"
	"github.com/gofiber/fiber/v2"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) CreateStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	item, err := h.svc.CreateStudent(c.UserContext(), c.Params("orgID"), actorID(c), req)
	if err != nil {
		return academicError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(response.Created(item))
}

func (h *Handler) ListStudents(c *fiber.Ctx) error {
	items, err := h.svc.ListStudents(c.UserContext(), c.Params("orgID"), actorID(c))
	if err != nil {
		return academicError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) CreateSemester(c *fiber.Ctx) error {
	var req CreateSemesterRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	item, err := h.svc.CreateSemester(c.UserContext(), c.Params("orgID"), actorID(c), req)
	if err != nil {
		return academicError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(response.Created(item))
}

func (h *Handler) ListSemesters(c *fiber.Ctx) error {
	items, err := h.svc.ListSemesters(c.UserContext(), c.Params("orgID"), actorID(c))
	if err != nil {
		return academicError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) CreateCourse(c *fiber.Ctx) error {
	var req CreateCourseRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	item, err := h.svc.CreateCourse(c.UserContext(), c.Params("orgID"), actorID(c), req)
	if err != nil {
		return academicError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(response.Created(item))
}

func (h *Handler) ListCourses(c *fiber.Ctx) error {
	items, err := h.svc.ListCourses(c.UserContext(), c.Params("orgID"), actorID(c))
	if err != nil {
		return academicError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) ListGradeScale(c *fiber.Ctx) error {
	items, err := h.svc.ListGradeScale(c.UserContext(), c.Params("orgID"), actorID(c))
	if err != nil {
		return academicError(c, err)
	}
	return c.JSON(response.OK(items))
}

func (h *Handler) UpsertResult(c *fiber.Ctx) error {
	var req UpsertResultRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	item, err := h.svc.UpsertResult(c.UserContext(), c.Params("orgID"), actorID(c), req)
	if err != nil {
		return academicError(c, err)
	}
	return c.JSON(response.OK(item))
}

func (h *Handler) Transcript(c *fiber.Ctx) error {
	item, err := h.svc.Transcript(c.UserContext(), c.Params("orgID"), c.Params("studentID"), actorID(c))
	if err != nil {
		return academicError(c, err)
	}
	return c.JSON(response.OK(item))
}

func bind(c *fiber.Ctx, req any) error {
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err.Error()))
	}
	return nil
}

func actorID(c *fiber.Ctx) string { return c.Locals("user_id").(string) }

func academicError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrForbidden):
		return c.Status(fiber.StatusForbidden).JSON(response.Error("forbidden"))
	case errors.Is(err, ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(response.NotFound())
	case errors.Is(err, ErrConflict):
		return c.Status(fiber.StatusConflict).JSON(response.Error("record already exists"))
	case errors.Is(err, ErrOrganizationLocked):
		return c.Status(fiber.StatusLocked).JSON(response.Error(err.Error()))
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
	}
}
