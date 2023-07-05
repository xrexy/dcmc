package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/xrexy/dmc/docker"
	"github.com/xrexy/dmc/parser"
	"github.com/xrexy/dmc/utils"
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

	var dependencies []*multipart.FileHeader
	for formFieldName, fileHeaders := range form.File {
		for _, file := range fileHeaders {
			if formFieldName == "plugin" || formFieldName != "dependency" {
				continue
			}

			if file.Header.Get("Content-Type") == "application/java-archive" {
				dependencies = append(dependencies, file)
			}
		}
	}

	fmt.Println("Dependencies: ", dependencies)

	plugins := make([]*parser.Plugin, 0)

	var wg sync.WaitGroup

	plugin, err := parser.ParsePlugin(form.File["plugin"][0], true)
	if err != nil {
		fmt.Println("Error parsing plugin", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Error parsing plugin. Check if the plugin.yml is valid",
		})
	}

	requiredDeps := plugin.Dependencies // required takes into account all plugins's "depend"s
	softDeps := plugin.SoftDependencies //  softDeps takes only the main plugin's "softdepend"s

	plugins = append(plugins, plugin)

	for _, dep := range dependencies {
		wg.Add(1)

		go func(dep *multipart.FileHeader) {
			defer wg.Done()

			plugin, err := parser.ParsePlugin(dep, false)
			if err != nil {
				fmt.Println("Error parsing plugin", err)
				return
			}

			if plugin.File == nil {
				fmt.Println("Error parsing plugin. Has no File associated with it", err)
				return
			}

			// Check if the plugin is required by the main plugin
			if utils.ContainsString(requiredDeps, plugin.Name) || utils.ContainsString(softDeps, plugin.Name) {
				plugins = append(plugins, plugin)
				if plugin.Dependencies != nil {
					requiredDeps = append(requiredDeps, plugin.Dependencies...)
				}

				return
			}
		}(dep)
	}

	wg.Wait()

	uuid := uuid.New().String()

	err = docker.StartContainer(plugins, uuid)
	if err != nil {
		fmt.Println("Error starting container", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error starting container",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "OK",
		"plugins": plugins,
		"uuid":    uuid,
		"dependencies": fiber.Map{
			"required": requiredDeps,
			"soft":     softDeps,
		},
	})
}

type StopContainerRequest struct {
	UUID string `json:"uuid"`
}

func (h *PluginsHandler) StopContainer(c *fiber.Ctx) error {
	var req StopContainerRequest
	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid body",
		})
	}

	if req.UUID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing 'uuid' field",
		})
	}

	err = docker.StopContainer(req.UUID)
	if err != nil {
		fmt.Println("Error stopping container", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "OK",
		"uuid":   req.UUID,
	})
}
