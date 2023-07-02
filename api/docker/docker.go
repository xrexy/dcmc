package docker

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/xrexy/dmc/parser"
	"github.com/xrexy/dmc/utils"
)

// type Docker struct {
// 	client *client.Client
// }

// func NewDocker() *Docker {
// 	cli, err := client.NewClientWithOpts(client.FromEnv)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return &Docker{
// 		client: cli,
// 	}
// }

func StartContainer(plugins []*parser.Plugin, uuid string) {
	dirName := fmt.Sprintf("storage/%s/plugins", uuid)

	// create folders
	if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	// upload plugins
	for _, plugin := range plugins {
		wg.Add(1)
		go func(plugin *parser.Plugin) {
			defer wg.Done()

			if err := utils.SaveFile(fmt.Sprintf("%s-%s.jar", plugin.Name, plugin.Version), plugin.File, dirName); err != nil {
				panic(err)
			}
		}(plugin)
	}

	wg.Wait()

	// start container
	fmt.Println("Starting container")

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	imageName := "itzg/minecraft-server:latest"
	containerName := fmt.Sprintf("dmc-%s", uuid)

	_, err = cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Volumes: map[string]struct{}{
			"/data": {},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"25565/tcp": []nat.PortBinding{
				{
					HostIP:   " ",
					HostPort: "25565",
				},
			},
		},
	}, nil, nil, containerName)

	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, containerName, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println("Container started")
}

func StopContainer(uuid string) {
	dirName := fmt.Sprintf("storage/%s", uuid)

	if err := os.RemoveAll(dirName); err != nil {
		panic(err)
	}
}
