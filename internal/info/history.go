package info

import (
	"encoding/gob"
	"os"
	"path/filepath"

	"github.com/zyedidia/micro/v2/internal/config"
)

// LoadHistory attempts to load user history from configDir/buffers/history
// into the history map
// The savehistory option must be on
func (i *InfoBuf) LoadHistory() {
	if config.GetGlobalOption("savehistory").(bool) {
		file, err := os.Open(filepath.Join(config.ConfigDir, "buffers", "history"))
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

		file, err := os.Create(filepath.Join(config.ConfigDir, "buffers", "history"))
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

// AddToHistory adds a new item to the history for the prompt type `ptype`.
// This function is not used by micro itself. It is useful for plugins
// which add their own items to the history, bypassing the infobar command line.
func (i *InfoBuf) AddToHistory(ptype string, item string) {
	if i.HasPrompt && i.PromptType == ptype {
		return
	}

	if _, ok := i.History[ptype]; !ok {
		i.History[ptype] = []string{item}
	} else {
		i.History[ptype] = append(i.History[ptype], item)

		// avoid duplicates
		h := i.History[ptype]
		for j := len(h) - 2; j >= 0; j-- {
			if h[j] == h[len(h)-1] {
				i.History[ptype] = append(h[:j], h[j+1:]...)
				break
			}
		}
	}
}

// UpHistory fetches the previous item in the history
func (i *InfoBuf) UpHistory(history []string) {
	if i.HistoryNum > 0 && i.HasPrompt && !i.HasYN {
		i.HistoryNum--
		i.Replace(i.Start(), i.End(), history[i.HistoryNum])
		i.Buffer.GetActiveCursor().GotoLoc(i.End())
	}
}

// DownHistory fetches the next item in the history
func (i *InfoBuf) DownHistory(history []string) {
	if i.HistoryNum < len(history)-1 && i.HasPrompt && !i.HasYN {
		i.HistoryNum++
		i.Replace(i.Start(), i.End(), history[i.HistoryNum])
		i.Buffer.GetActiveCursor().GotoLoc(i.End())
	}
}
