package handlers

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/xrexy/dmc/docker"
)

type PluginsHandler struct {
}

func NewPluginsHandler() *PluginsHandler {
	return &PluginsHandler{}
}

func (h *PluginsHandler) CreateContainer(c *fiber.Ctx) error {
	if c.Query("spiget") == "true" {
		uuid := uuid.New().String()

		// string ids
		var pluginsIds []uint32
		err := c.BodyParser(&pluginsIds)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid body",
			})
		}

		err = docker.StartSpigetContainer(uuid, pluginsIds)
		if err != nil {
			fmt.Println("Error starting spiget container", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error starting container",
			})
		}

		return c.JSON(fiber.Map{
			"uuid": uuid,
			"data": fiber.Map{"pluginsIds": pluginsIds, "mode": "spiget"},
		})
	}

	return c.JSON(fiber.Map{
		"error": "Invalid mode; Only spiget is currently supported.",
	})

	// --- works fine, don't want to expose to frontend yet.

	// form, err := c.MultipartForm()
	// if err != nil {
	// 	return c.Status(http.StatusBadRequest).JSON(fiber.Map{
	// 		"error": "Invalid multipart form",
	// 	})
	// }

	// if form.File["plugin"] == nil {
	// 	return c.Status(http.StatusBadRequest).JSON(fiber.Map{
	// 		"error": "Missing part 'plugin' in multipart form",
	// 	})
	// }

	// var dependencies []*multipart.FileHeader
	// for formFieldName, fileHeaders := range form.File {
	// 	for _, file := range fileHeaders {
	// 		if formFieldName == "plugin" || formFieldName != "dependency" {
	// 			continue
	// 		}

	// 		if file.Header.Get("Content-Type") == "application/java-archive" {
	// 			dependencies = append(dependencies, file)
	// 		}
	// 	}
	// }

	// fmt.Println("Dependencies: ", dependencies)

	// plugins := make([]*parser.Plugin, 0)

	// var wg sync.WaitGroup

	// plugin, err := parser.ParsePlugin(form.File["plugin"][0])
	// if err != nil {
	// 	fmt.Println("Error parsing plugin", err)
	// 	return c.Status(http.StatusBadRequest).JSON(fiber.Map{
	// 		"error": "Error parsing plugin. Check if the plugin.yml is valid",
	// 	})
	// }

	// requiredDeps := plugin.Dependencies // required takes into account all plugins's "depend"s
	// softDeps := plugin.SoftDependencies //  softDeps takes only the main plugin's "softdepend"s

	// if len(requiredDeps) > len(dependencies) {
	// 	return c.Status(http.StatusBadRequest).JSON(fiber.Map{
	// 		"error":    fmt.Sprintf("Missing dependencies. Required: %v, has %v", len(requiredDeps), len(dependencies)),
	// 		"required": requiredDeps,
	// 	})
	// }

	// plugins = append(plugins, plugin)

	// for _, dep := range dependencies {
	// 	wg.Add(1)

	// 	go func(dep *multipart.FileHeader) {
	// 		defer wg.Done()

	// 		plugin, err := parser.ParsePlugin(dep)
	// 		if err != nil {
	// 			fmt.Println("Error parsing plugin", err)
	// 			return
	// 		}

	// 		if plugin.File == nil {
	// 			fmt.Println("Error parsing plugin. Has no File associated with it", err)
	// 			return
	// 		}

	// 		if utils.ContainsString(requiredDeps, plugin.Name) || utils.ContainsString(softDeps, plugin.Name) {
	// 			plugins = append(plugins, plugin)
	// 		} else {
	// 			return
	// 		}

	// 		if plugin.Dependencies != nil {
	// 			requiredDeps = append(requiredDeps, plugin.Dependencies...)
	// 		}

	// 	}(dep)
	// }

	// wg.Wait()

	// // Make sure required dependencies are satisfied
	// if len(requiredDeps) > len(plugins) {
	// 	return c.Status(http.StatusBadRequest).JSON(fiber.Map{
	// 		"error":    fmt.Sprintf("Too little dependencies. Required: %v, has %v", len(requiredDeps), len(plugins)),
	// 		"required": requiredDeps,
	// 	})
	// }

	// // Make sure all required dependencies match the ones in the plugins slice
	// found := 0
	// for _, dep := range requiredDeps {
	// 	for _, plugin := range plugins {
	// 		fmt.Println("Plugin name: ", plugin.Name, "Dep: ", dep)
	// 		if plugin.Name == dep {
	// 			found++
	// 			break
	// 		}
	// 	}
	// }

	// if found != len(requiredDeps) {
	// 	return c.Status(http.StatusBadRequest).JSON(fiber.Map{
	// 		"error":    fmt.Sprintf("Missing required dependencies. Required: %v, has %v", len(requiredDeps), found),
	// 		"required": requiredDeps,
	// 	})
	// }

	// uuid := uuid.New().String()

	// err = docker.StartContainer(uuid, plugins)
	// if err != nil {
	// 	fmt.Println("Error starting container", err)
	// 	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
	// 		"error": "Error starting container",
	// 	})
	// }

	// return c.JSON(fiber.Map{
	// 	"uuid": uuid,
	// 	"data": fiber.Map{
	// 		"plugins": plugins,
	// 		"mode":    "local",
	// 		"dependencies": fiber.Map{
	// 			"required": requiredDeps,
	// 			"soft":     softDeps,
	// 		},
	// 	},
	// })
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
