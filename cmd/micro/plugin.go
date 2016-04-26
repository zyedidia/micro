package main

import (
	"io/ioutil"
)

var loadedPlugins []string

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
}
