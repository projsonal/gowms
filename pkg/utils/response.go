package utils

import "github.com/gofiber/fiber/v2"

type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func OK(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{Success: true, Message: message, Data: data})
}

func OKWithMeta(c *fiber.Ctx, message string, data interface{}, meta interface{}) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{Success: true, Message: message, Data: data, Meta: meta})
}

func Created(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{Success: true, Message: message, Data: data})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Fail(c *fiber.Ctx, status int, message string, errs interface{}) error {
	return c.Status(status).JSON(Envelope{Success: false, Message: message, Errors: errs})
}

func NotFound(c *fiber.Ctx) error {
	return Fail(c, fiber.StatusNotFound, "resource tidak ditemukan", nil)
}
