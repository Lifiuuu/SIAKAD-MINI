package helper

import (
	"github.com/gofiber/fiber/v2"
	"siakad-mini/app/model"
)

// SuccessResponse for standardized success output
func SuccessResponse(c *fiber.Ctx, status int, message string, data any, meta any) error {
	resp := model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	}
	return c.Status(status).JSON(resp)
}

// ErrorResponsePayload for standardized error output
func ErrorResponsePayload(c *fiber.Ctx, status int, message string, errors any) error {
	resp := model.WebResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	}
	return c.Status(status).JSON(resp)
}

// Original helper functions for compatibility (if needed by other files not rewritten)
func Success(c *fiber.Ctx, status int, message string, data any) error {
	return SuccessResponse(c, status, message, data, nil)
}

func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return SuccessResponse(c, fiber.StatusCreated, message, data, nil)
}
