package parser

import (
	"archive/zip"
	"fmt"
	"mime/multipart"

	"gopkg.in/yaml.v3"
)

type Plugin struct {
	Name             string         `yaml:"name" json:"name"`
	Version          string         `yaml:"version" json:"version"`
	Dependencies     []string       `yaml:"depend" json:"depend"`
	SoftDependencies []string       `yaml:"softdepend" json:"softdepend"`
	IsMain           bool           `yaml:"-" json:"isMain"`
	File             multipart.File `yaml:"-" json:"-"`
}

func ParsePlugin(file *multipart.FileHeader, isMain bool) (*Plugin, error) {
	zipReader, err := toZipReader(file)
	if err != nil {
		return nil, err
	}

	pluginYaml, err := findPluginYaml(zipReader)
	if err != nil {
		return nil, err
	}

	if pluginYaml == nil {
		return nil, fmt.Errorf("plugin.yml not found in %s", file.Filename)
	}

	plugin, err := parsePluginYaml(pluginYaml)
	if err != nil {
		return nil, err
	}

	multiFile, err := file.Open()
	if err != nil {
		return nil, err
	}

	plugin.File = multiFile
	plugin.IsMain = isMain

	return plugin, nil
}

func toZipReader(file *multipart.FileHeader) (*zip.Reader, error) {
	data, err := file.Open()
	if err != nil {
		return nil, err
	}

	return zip.NewReader(data, file.Size)
}

func findPluginYaml(zipReader *zip.Reader) (*zip.File, error) {
	for _, file := range zipReader.File {
		if file.Name == "plugin.yml" {
			return file, nil
		}
	}

	return nil, nil
}

func parsePluginYaml(file *zip.File) (*Plugin, error) {
	data, err := file.Open()
	if err != nil {
		return nil, err
	}

	pluginYAML := &Plugin{}
	err = yaml.NewDecoder(data).Decode(pluginYAML)
	if err != nil {
		return nil, err
	}

	return pluginYAML, nil
}
