package parser

import (
	"archive/zip"
	"errors"
	"fmt"
	"mime/multipart"

	"gopkg.in/yaml.v2"
)

type ParsedPluginDependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ParsedPlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`

	Dependencies []ParsedPluginDependency `json:"dependencies"`
}

type PartialPluginYAML struct {
	Name             string   `yaml:"name" json:"name"`
	Version          string   `yaml:"version" json:"version"`
	HardDependencies []string `yaml:"depend" json:"depend"`
	SoftDependencies []string `yaml:"softdepend" json:"softdepend"`
}

func ParsePluginFile(file *multipart.FileHeader, dependencies []*multipart.FileHeader) (*ParsedPlugin, *PartialPluginYAML, error) {
	zip, err := toZipReader(file)
	if err != nil {
		return nil, nil, err
	}

	for _, zipFile := range zip.File {
		if zipFile.Name == "plugin.yml" {
			fmt.Println("Found plugin.yml, parsing...")
			pluginYaml, err := parsePlugin(zipFile)
			if err != nil {
				return nil, nil, err
			}

			fmt.Println("Plugin YAML:", pluginYaml)

			fmt.Printf("Found %d dependencies\n", len(dependencies))
			fmt.Println("Hard dependencies:", pluginYaml.HardDependencies)
			fmt.Println("Soft dependencies:", pluginYaml.SoftDependencies)

			if len(pluginYaml.HardDependencies) > len(dependencies) {
				return nil, nil, fmt.Errorf("plugin requires %d dependencies, but only %d were provided", len(pluginYaml.HardDependencies), len(dependencies))
			}

			if len(pluginYaml.HardDependencies) != 0 && len(dependencies) == 0 {
				return nil, pluginYaml, fmt.Errorf("plugin has required dependencies but no dependencies were provided")
			}

			pdeps := make([]ParsedPluginDependency, 0)
			if len(dependencies) > 0 {
				fmt.Println("Parsing dependencies...")
				pdeps, err = parsePluginDependencies(pluginYaml, dependencies)
				if err != nil {
					return nil, pluginYaml, err
				}
			}

			return &ParsedPlugin{
				Name:         pluginYaml.Name,
				Version:      pluginYaml.Version,
				Dependencies: pdeps,
			}, pluginYaml, nil
		}
	}

	return nil, nil, errors.New("plugin.yml not found")
}

func parsePluginDependencies(plugin *PartialPluginYAML, dependencies []*multipart.FileHeader) ([]ParsedPluginDependency, error) {
	var parsedDependencies []ParsedPluginDependency

	hardFound := 0
	for _, dependency := range dependencies {
		dependencyPlugin, _, err := ParsePluginFile(dependency, nil)
		if err != nil {
			continue
		}

		foundHard := false
		if contains(plugin.HardDependencies, dependencyPlugin.Name) {
			foundHard = true
			hardFound++
		}

		// check if deps contains this plugin
		if (!foundHard && !contains(plugin.HardDependencies, dependencyPlugin.Name)) && !contains(plugin.SoftDependencies, dependencyPlugin.Name) {
			continue
		}

		parsedDependencies = append(parsedDependencies, ParsedPluginDependency{
			Name:    dependencyPlugin.Name,
			Version: dependencyPlugin.Version,
		})
	}

	if hardFound != len(plugin.HardDependencies) {
		// return nil, errors.New("hard dependencies missmatch; expected " + len(plugin.HardDependencies)) + " but found " + hardFound) + " dependencies")
		return nil, fmt.Errorf("hard dependencies missmatch; expected %d but found %d dependencies", len(plugin.HardDependencies), hardFound)
	}

	return parsedDependencies, nil
}

func toZipReader(file *multipart.FileHeader) (*zip.Reader, error) {
	data, err := file.Open()
	if err != nil {
		return nil, err
	}

	return zip.NewReader(data, file.Size)
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

func contains(slice []string, element string) bool {
	for _, sliceElement := range slice {
		if sliceElement == element {
			return true
		}
	}

	return false
}
