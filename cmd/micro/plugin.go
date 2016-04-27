package main

import (
	"io/ioutil"
)

var loadedPlugins []string

var preInstalledPlugins = []string{
	"go",
}

// LoadPlugins loads the pre-installed plugins and the plugins located in ~/.config/micro/plugins
func LoadPlugins() {
	files, _ := ioutil.ReadDir(configDir + "/plugins")
	for _, plugin := range files {
		if plugin.IsDir() {
			pluginName := plugin.Name()
			files, _ := ioutil.ReadDir(configDir + "/plugins/" + pluginName)
			for _, f := range files {
				if f.Name() == pluginName+".lua" {
					if err := L.DoFile(configDir + "/plugins/" + pluginName + "/" + f.Name()); err != nil {
						TermMessage(err)
						continue
					}
					loadedPlugins = append(loadedPlugins, pluginName)
				}
			}
		}
	}

	for _, pluginName := range preInstalledPlugins {
		plugin := "runtime/plugins/" + pluginName + "/" + pluginName + ".lua"
		data, err := Asset(plugin)
		if err != nil {
			TermMessage("Error loading pre-installed plugin: " + pluginName)
			continue
		}
		if err := L.DoString(string(data)); err != nil {
			TermMessage(err)
			continue
		}
		loadedPlugins = append(loadedPlugins, pluginName)
	}
}
