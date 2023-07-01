package handlers

import "github.com/gofiber/fiber/v2"

type PluginsHandler struct {
}

func NewPluginsHandler() *PluginsHandler {
	return &PluginsHandler{}
}

func (h *PluginsHandler) FetchPlugins(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"data": []string{},
	})
}
