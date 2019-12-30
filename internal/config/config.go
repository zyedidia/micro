package config

import (
	"errors"
	"os"

	homedir "github.com/mitchellh/go-homedir"
)

var ConfigDir string

// InitConfigDir finds the configuration directory for micro according to the XDG spec.
// If no directory is found, it creates one.
func InitConfigDir(flagConfigDir string) error {
	var e error

	microHome := os.Getenv("MICRO_CONFIG_HOME")
	if microHome == "" {
		// The user has not set $MICRO_CONFIG_HOME so we'll try $XDG_CONFIG_HOME
		xdgHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgHome == "" {
			// The user has not set $XDG_CONFIG_HOME so we should act like it was set to ~/.config
			home, err := homedir.Dir()
			if err != nil {
				return errors.New("Error finding your home directory\nCan't load config files")
			}
			xdgHome = home + "/.config"
		}
		microHome = xdgHome + "/micro"
	}
	ConfigDir = microHome

	if len(flagConfigDir) > 0 {
		if _, err := os.Stat(flagConfigDir); os.IsNotExist(err) {
			e = errors.New("Error: " + flagConfigDir + " does not exist. Defaulting to " + ConfigDir + ".")
		} else {
			ConfigDir = flagConfigDir
			return nil
		}
	}

	if _, err := os.Stat(microHome); os.IsNotExist(err) {
		// If the microHome doesn't exist we should create it
		err = os.Mkdir(microHome, os.ModePerm)
		if err != nil {
			return errors.New("Error creating XDG_CONFIG_HOME directory: " + err.Error())
		}
	}

	if _, err := os.Stat(ConfigDir); os.IsNotExist(err) {
		// If the micro specific config directory doesn't exist we should create that too
		err = os.Mkdir(ConfigDir, os.ModePerm)
		if err != nil {
			return errors.New("Error creating configuration directory: " + err.Error())
		}
	}

	return e
}
