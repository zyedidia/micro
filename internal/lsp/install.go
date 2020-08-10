package lsp

import (
	"errors"
	"io"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
)

var ErrManualInstall = errors.New("Requires manual installation")

type Config struct {
	Languages map[string]Language `toml:"language"`
}

type Language struct {
	Command string     `toml:"command"`
	Args    []string   `toml:"args"`
	Install [][]string `toml:"install"`
}

var conf *Config

func GetLanguage(lang string) (Language, bool) {
	l, ok := conf.Languages[lang]
	return l, ok
}

func init() {
	conf, _ = LoadConfig([]byte(servers))
}

func LoadConfig(data []byte) (*Config, error) {
	var conf Config
	if _, err := toml.Decode(string(data), &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (l *Language) Installed() bool {
	_, err := exec.LookPath(l.Command)
	if err != nil {
		return false
	}

	return true
}

func (l *Language) DoInstall(w io.Writer) error {
	if l.Installed() {
		return nil
	}

	if len(l.Install) == 0 {
		return ErrManualInstall
	}

	for _, c := range l.Install {
		io.WriteString(w, strings.Join(c, " ")+"\n")
		cmd := exec.Command(c[0], c[1:]...)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
