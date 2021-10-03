package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zyedidia/micro/v2/internal/buffer"
	"github.com/zyedidia/micro/v2/internal/config"
)

func shouldContinue() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Continue [Y/n]: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return false
	}

	text = strings.TrimRight(text, "\r\n")

	return len(text) == 0 || strings.ToLower(text)[0] == 'y'
}

// CleanConfig performs cleanup in the user's configuration directory
func CleanConfig() {
	fmt.Println("Cleaning your configuration directory at", config.ConfigDir)
	fmt.Printf("Please consider backing up %s before continuing\n", config.ConfigDir)

	if !shouldContinue() {
		fmt.Println("Stopping early")
		return
	}

	fmt.Println("Cleaning default settings")
	config.WriteSettings(filepath.Join(config.ConfigDir, "settings.json"))

	// detect unused options
	var unusedOptions []string
	defaultSettings := config.DefaultAllSettings()
	for k := range config.GlobalSettings {
		if _, ok := defaultSettings[k]; !ok {
			valid := false
			for _, p := range config.Plugins {
				if strings.HasPrefix(k, p.Name+".") || k == p.Name {
					valid = true
				}
			}
			if !valid {
				unusedOptions = append(unusedOptions, k)
			}
		}
	}

	if len(unusedOptions) > 0 {
		fmt.Println("The following options are unused:")

		sort.Strings(unusedOptions)

		for _, s := range unusedOptions {
			fmt.Printf("%s (value: %v)\n", s, config.GlobalSettings[s])
		}

		fmt.Printf("These options will be removed from %s\n", filepath.Join(config.ConfigDir, "settings.json"))

		if shouldContinue() {
			for _, s := range unusedOptions {
				delete(config.GlobalSettings, s)
			}

			err := config.OverwriteSettings(filepath.Join(config.ConfigDir, "settings.json"))
			if err != nil {
				fmt.Println("Error writing settings.json file: " + err.Error())
			}

			fmt.Println("Removed unused options")
			fmt.Print("\n\n")
		}
	}

	// detect incorrectly formatted buffer/ files
	files, err := ioutil.ReadDir(filepath.Join(config.ConfigDir, "buffers"))
	if err == nil {
		var badFiles []string
		var buffer buffer.SerializedBuffer
		for _, f := range files {
			fname := filepath.Join(config.ConfigDir, "buffers", f.Name())
			file, e := os.Open(fname)

			if e == nil {
				decoder := gob.NewDecoder(file)
				err = decoder.Decode(&buffer)

				if err != nil && f.Name() != "history" {
					badFiles = append(badFiles, fname)
				}
				file.Close()
			}
		}

		if len(badFiles) > 0 {
			fmt.Printf("Detected %d files with an invalid format in %s\n", len(badFiles), filepath.Join(config.ConfigDir, "buffers"))
			fmt.Println("These files store cursor and undo history.")
			fmt.Printf("Removing badly formatted files in %s\n", filepath.Join(config.ConfigDir, "buffers"))

			if shouldContinue() {
				removed := 0
				for _, f := range badFiles {
					err := os.Remove(f)
					if err != nil {
						fmt.Println(err)
						continue
					}
					removed++
				}

				if removed == 0 {
					fmt.Println("Failed to remove files")
				} else {
					fmt.Printf("Removed %d badly formatted files\n", removed)
				}
				fmt.Print("\n\n")
			}
		}
	}

	// detect plugins/ directory
	plugins := filepath.Join(config.ConfigDir, "plugins")
	if stat, err := os.Stat(plugins); err == nil && stat.IsDir() {
		fmt.Printf("Found directory %s\n", plugins)
		fmt.Printf("Plugins should now be stored in %s\n", filepath.Join(config.ConfigDir, "plug"))
		fmt.Printf("Removing %s\n", plugins)

		if shouldContinue() {
			os.RemoveAll(plugins)
		}

		fmt.Print("\n\n")
	}

	fmt.Println("Done cleaning")
}
