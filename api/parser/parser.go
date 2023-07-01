package parser

import (
	"archive/zip"
	"errors"
	"fmt"
	"mime/multipart"

	"gopkg.in/yaml.v2"
)

type ParsedPluginDependency struct {
	PluginName    string
	PluginVersion string
}

type ParsedPlugin struct {
	PluginName    string
	PluginVersion string

	Dependencies []ParsedPluginDependency
}

type PartialPluginYAML struct {
	Name             string   `yaml:"name"`
	Version          string   `yaml:"version"`
	HardDependencies []string `yaml:"depend"`
	SoftDependencies []string `yaml:"softdepend"`
}

func ParsePluginFile(file *multipart.FileHeader, dependencies []*multipart.FileHeader) (*ParsedPlugin, error) {
	// parse plugin file
	data, err := file.Open()
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(data, file.Size)
	if err != nil {
		return nil, err
	}

	for _, zipFile := range zipReader.File {
		if zipFile.Name == "plugin.yml" {
			fmt.Println("Found plugin.yml, parsing...")
			plugin, err := parsePlugin(zipFile)
			if err != nil {
				return nil, err
			}

			fmt.Println("Plugin YAML:", plugin)

			break
		}
	}

	return nil, errors.New("plugin.yml not found")
}

func parsePlugin(file *zip.File) (*PartialPluginYAML, error) {
	data, err := file.Open()
	if err != nil {
		return nil, err
	}

	pluginYAML := &PartialPluginYAML{}
	err = yaml.NewDecoder(data).Decode(pluginYAML)
	if err != nil {
		return nil, err
	}

	return pluginYAML, nil
}
