package utils

import "github.com/xrexy/dmc/parser"

func GetPluginNames(plugin *parser.Plugin) []string {
	var names []string
	names = append(names, plugin.Dependencies...)
	names = append(names, plugin.SoftDependencies...)
	return names
}

func GetPluginNamesMultiple(plugins []*parser.Plugin) []string {
	var names []string
	for _, plugin := range plugins {
		names = append(names, GetPluginNames(plugin)...)
	}
	return names
}
