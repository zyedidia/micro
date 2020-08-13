package lsp

import (
	"errors"
	"io"
	"log"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

var ErrManualInstall = errors.New("Requires manual installation")

type Config struct {
	Languages map[string]Language `yaml:"language"`
}

type Language struct {
	Command string     `yaml:"command"`
	Args    []string   `yaml:"args"`
	Install [][]string `yaml:"install"`
}

var conf *Config

func GetLanguage(lang string) (Language, bool) {
	if conf != nil {
		l, ok := conf.Languages[lang]
		return l, ok
	}
	return Language{}, false
}

func init() {
	var err error
	conf, err = LoadConfig([]byte(servers))
	if err != nil {
		log.Println("[micro-lsp]", err)
	}
}

func LoadConfig(data []byte) (*Config, error) {
	var conf Config
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (l Language) Installed() bool {
	_, err := exec.LookPath(l.Command)
	if err != nil {
		return false
	}

	return true
}

func (l Language) DoInstall(w io.Writer) error {
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
