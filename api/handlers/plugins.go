package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/xrexy/dmc/parser"
)

type PluginsHandler struct {
}

func NewPluginsHandler() *PluginsHandler {
	return &PluginsHandler{}
}

func (h *PluginsHandler) CreateContainer(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid multipart form",
		})
	}

	if form.File["plugin"] == nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing part 'plugin' in multipart form",
		})
	}

	plugin := form.File["plugin"][0]
	fmt.Println(plugin.Filename)

	var dependencies []*multipart.FileHeader
	for formFieldName, fileHeaders := range form.File {
		for _, file := range fileHeaders {
			if formFieldName == "plugin" {
				continue
			}

			if file.Header.Get("Content-Type") == "application/java-archive" {
				dependencies = append(dependencies, file)
			}
		}
	}

	go parser.ParsePluginFile(plugin, dependencies)

	return c.JSON(fiber.Map{
		"status": "OK",
	})
}
