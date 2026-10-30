package apihandler

import (
	"errors"
	"log"
	"regexp"
	"simplemavenrepo/internal/repository"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

var namePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

func validateName(kind, name string) error {
	if !namePattern.MatchString(name) {
		return fiber.NewError(fiber.StatusBadRequest,
			kind+" must be 1-64 characters of letters, digits, '.', '_' or '-' and start with a letter or digit")
	}
	return nil
}

// repoError converts a repository error into an HTTP error. what names the entity for the error message.
func repoError(err error, what string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, what+" not found")
	case errors.Is(err, repository.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, what+" already exists")
	}
	log.Printf("Database error: %v", err)
	return fiber.ErrInternalServerError
}

func idParam(c fiber.Ctx, key string) (int64, error) {
	id, err := strconv.ParseInt(c.Params(key), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+key)
	}
	return id, nil
}

func bindJSON(c fiber.Ctx, out any) error {
	if err := c.Bind().JSON(out); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
	}
	return nil
}
