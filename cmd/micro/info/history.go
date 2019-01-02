package info

import (
	"encoding/gob"
	"os"

	"github.com/zyedidia/micro/cmd/micro/config"
)

// LoadHistory attempts to load user history from configDir/buffers/history
// into the history map
// The savehistory option must be on
func (i *InfoBuf) LoadHistory() {
	if config.GetGlobalOption("savehistory").(bool) {
		file, err := os.Open(config.ConfigDir + "/buffers/history")
		defer file.Close()
		var decodedMap map[string][]string
		if err == nil {
			decoder := gob.NewDecoder(file)
			err = decoder.Decode(&decodedMap)

			if err != nil {
				i.Error("Error loading history:", err)
				return
			}
		}

		if decodedMap != nil {
			i.History = decodedMap
		} else {
			i.History = make(map[string][]string)
		}
	} else {
		i.History = make(map[string][]string)
	}
}

// SaveHistory saves the user's command history to configDir/buffers/history
// only if the savehistory option is on
func (i *InfoBuf) SaveHistory() {
	if config.GetGlobalOption("savehistory").(bool) {
		// Don't save history past 100
		for k, v := range i.History {
			if len(v) > 100 {
				i.History[k] = v[len(i.History[k])-100:]
			}
		}

		file, err := os.Create(config.ConfigDir + "/buffers/history")
		defer file.Close()
		if err == nil {
			encoder := gob.NewEncoder(file)

			err = encoder.Encode(i.History)
			if err != nil {
				i.Error("Error saving history:", err)
				return
			}
		}
	}
}
