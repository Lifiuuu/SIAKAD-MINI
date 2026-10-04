package config

import (
	"errors"
	"log/slog"

	"siakad-mini/app/model"
	"siakad-mini/helper"
	"siakad-mini/middleware"
	"siakad-mini/route"

	"github.com/gofiber/fiber/v2"
)

// NewApp merakit aplikasi: membuat instance Fiber, memasang middleware,
// lalu mendaftarkan route.
func NewApp(
	logger *slog.Logger,
	deps route.Dependencies,
	allowedOrigins string,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "SIAKAD Mini"),
		ErrorHandler: newErrorHandler(logger),
		// Membatasi ukuran body mencegah denial of service.
		BodyLimit: 1 * 1024 * 1024, // 1 MB
	})

	middleware.Register(app, logger, allowedOrigins)
	route.Register(app, deps)

	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})
	return app
}

// newErrorHandler adalah SATU-SATUNYA tempat error berubah menjadi response HTTP.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID := helper.RequestID(c)
		var appErr *helper.AppError
		switch {
		case errors.As(err, &appErr):
			// Kegagalan yang sudah kita rencanakan.
		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status:  fiber.StatusRequestEntityTooLarge,
				Code:    "PAYLOAD_TOO_LARGE",
				Message: "ukuran body melebihi batas yang diizinkan",
			}
		default:
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status:  fiberErr.Code,
					Code:    "HTTP_ERROR",
					Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}
		if appErr.Status >= fiber.StatusInternalServerError {
			logger.Error("request_failed",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
				slog.String("error", appErr.Error()))
		} else {
			logger.Warn("request_rejected",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status))
		}
		var errs any
		if len(appErr.Fields) > 0 {
			errMap := make(map[string][]string)
			for k, v := range appErr.Fields {
				errMap[k] = append(errMap[k], v)
			}
			errs = errMap
		}

		return c.Status(appErr.Status).JSON(model.WebResponse{
			Success: false,
			Message: appErr.Message,
			Errors:  errs,
		})
	}
}
