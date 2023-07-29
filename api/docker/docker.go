package docker

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	natting "github.com/docker/go-connections/nat"
	"github.com/xrexy/dmc/parser"
	"github.com/xrexy/dmc/utils"
)

type Gamemode string
type Version string

const (
	CREATIVE  Gamemode = "CREATIVE"
	SURVIVAL  Gamemode = "SURVIVAL"
	SPECTATOR Gamemode = "SPECTATOR"
	ADVENTURE Gamemode = "ADVENTURE"
)

func makeEnv(version string, gamemode Gamemode, spigetResources []uint32) []string {
	return []string{
		"EULA=TRUE",
		"TYPE=SPIGOT",
		fmt.Sprintf("VERSION=%s", version),
		fmt.Sprintf("MODE=%s", gamemode),
		// fmt.Sprintf("SPIGET_RESOURCES=%s", strings.Trim(strings.Replace(fmt.Sprint(spigetResources), " ", ",", -1), "[]")),
	}
}

func StartSpigetContainer(uuid string, pluginIds []uint32) error {
	// TODO don't create a new client every time
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	defer cli.Close()

	// create the folder
	dirName := fmt.Sprintf("docker/storage/%s/plugins", uuid)
	if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create container folder")
	}

	env := makeEnv("1.20.1", CREATIVE, pluginIds)
	return runContainer(cli, uuid, "25565", env)
}

func StartContainer(uuid string, plugins []*parser.Plugin) error {
	dirName := fmt.Sprintf("docker/storage/%s/plugins", uuid)

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
				return
			}
			fmt.Println("Saved file", plugin.Name, plugin.Version)
		}(plugin)
	}

	wg.Wait()

	// start container
	fmt.Println("Starting container")

	// TODO don't create a new client every time
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	defer cli.Close()

	return runContainer(cli, uuid, "25565", makeEnv("1.20.1", CREATIVE, []uint32{}))
}

func runContainer(client *client.Client, uuid string, port string, env []string) error {
	// Define a PORT opening
	newport, err := natting.NewPort("tcp", port)
	if err != nil {
		fmt.Println("Unable to create docker port")
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Unable to get working directory")
		return err
	}

	fmt.Println(uuid)

	files, err := os.ReadDir(fmt.Sprintf("%s/docker/storage/%s", pwd, uuid))
	if err != nil {
		fmt.Println("Unable to read directory")
		return err
	}

	for _, file := range files {
		fmt.Println(file.Name())
	}

	// Configured hostConfig:
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	hostConfig := &container.HostConfig{
		PortBindings: natting.PortMap{
			newport: []natting.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
		Mounts: []mount.Mount{
			{
				Type: mount.TypeBind,
				// Source: fmt.Sprintf("storage/%s", uuid),
				// absolute path, using pwd
				Source: fmt.Sprintf("%s/docker/storage/%s", pwd, uuid),

				Target: "/data",
			},
		},
	}

	// Define Network config (why isn't PORT in here...?:
	// https://godoc.org/github.com/docker/docker/api/types/network#NetworkingConfig
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	gatewayConfig := &network.EndpointSettings{
		Gateway: "gatewayname",
	}
	networkConfig.EndpointsConfig["bridge"] = gatewayConfig

	// Define ports to be exposed (has to be same as hostconfig.portbindings.newport)
	exposedPorts := map[natting.Port]struct{}{
		newport: {},
	}

	// Configuration
	// https://godoc.org/github.com/docker/docker/api/types/container#Config
	config := &container.Config{
		// Image:        "itzg/minecraft-server:java11-jdk",
		Image:        "itzg/minecraft-server:java17",
		ExposedPorts: exposedPorts,
		Env:          env,
	}

	// Creating the actual container. This is "nil,nil,nil" in every example.
	cont, err := client.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		networkConfig, nil,
		fmt.Sprintf("dmc-%s", uuid),
	)

	if err != nil {
		log.Println(err)
		return err
	}

	// Run the actual container
	client.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	log.Printf("Container %s is created", cont.ID)

	return nil
}

func StopContainer(uuid string) error {
	dirName := fmt.Sprintf("docker/storage/%s", uuid)

	// check if folder exists
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		return fmt.Errorf("we don't have a container with that uuid")
	}

	// remove folder
	if err := os.RemoveAll(dirName); err != nil {
		return fmt.Errorf("unable to remove folder")
	}

	// disabled docker client
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("unable to create docker client")
	}

	defer cli.Close()

	// check if container exists
	_, err = cli.ContainerInspect(context.Background(), fmt.Sprintf("dmc-%s", uuid))
	if err != nil {
		return fmt.Errorf("no container with that uuid exists")
	}

	err = cli.ContainerRemove(context.Background(), fmt.Sprintf("dmc-%s", uuid), types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})

	if err != nil {
		return fmt.Errorf("unable to remove container")
	}

	// remove folder
	return nil
}
