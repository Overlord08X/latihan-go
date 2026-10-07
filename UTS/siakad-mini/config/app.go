package config

import (
	"errors"

	"siakad-mini/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func NewApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "SIAKAD Mini API - UTS PBE",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var appErr *helper.AppError
			if errors.As(err, &appErr) {
				if appErr.Errors != nil {
					return helper.ValidationError(c, appErr.Message, appErr.Errors)
				}
				return helper.Error(c, appErr.Status, appErr.Message)
			}

			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				return helper.Error(c, fiberErr.Code, fiberErr.Message)
			}

			return helper.Error(c, fiber.StatusInternalServerError, "Terjadi kesalahan internal server")
		},
	})

	app.Use(recover.New())
	app.Use(RequestLogger())

	return app
}
