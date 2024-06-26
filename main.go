package main

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/xrexy/dmc/handlers"
	"go.uber.org/fx"
)

func scaffoldFiberServer(lc fx.Lifecycle, pluginsHandler *handlers.PluginsHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024, // 50 MB
	})

	app.Use(cors.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "OK",
		})
	})

	pluginsGroup := app.Group("/plugins")
	pluginsGroup.Get("/container/new", pluginsHandler.CreateContainer)
	// stop
	pluginsGroup.Post("/container/stop", pluginsHandler.StopContainer)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				port := ":8080"
				fmt.Printf("Listening on port %s\n", port)
				if err := app.Listen(port); err != nil {
					panic(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Shutdown()
		},
	})

	return app
}

func main() {
	fx.New(
		fx.Provide(
			handlers.NewPluginsHandler,
			// docker.NewDocker,
		),
		fx.Invoke(
			scaffoldFiberServer,
		),
	).Run()
}
