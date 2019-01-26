package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"path"

	"github.com/blang/semver"
	"github.com/zyedidia/micro/cmd/micro/config"
	git "gopkg.in/src-d/go-git.v4"
)

var (
	ErrMissingName     = errors.New("Missing or empty name field")
	ErrMissingDesc     = errors.New("Missing or empty description field")
	ErrMissingSite     = errors.New("Missing or empty website field")
	ErrMissingRepo     = errors.New("Missing or empty repository field")
	ErrMissingVersions = errors.New("Missing or empty versions field")
	ErrMissingTag      = errors.New("Missing or empty tag field")
	ErrMissingRequire  = errors.New("Missing or empty require field")
)

const (
	infojson    = "plugin.json"
	versionfile = "version.lock"
)

type Plugin struct {
	info    *PluginInfo
	dir     string
	repo    *git.Repository
	version semver.Version // currently installed version
}

func (p *Plugin) GetRequires() *PluginVersion {
	for _, v := range p.info.Versions {
		if p.version.Equals(v.Vers) {
			return &v
		}
	}
	return nil
}

// PluginVersion describes a version for a plugin as well as any dependencies that
// version might have
// This marks a tag that corresponds to the version in the git repo
type PluginVersion struct {
	Vers    semver.Version
	Vstr    string            `json:"version"`
	Tag     string            `json:"tag"`
	Require map[string]string `json:"require"`
}

// PluginInfo contains all the needed info about a plugin
type PluginInfo struct {
	Name     string          `json:"name"`
	Desc     string          `json:"description"`
	Site     string          `json:"website"`
	Repo     string          `json:"repository"`
	Versions []PluginVersion `json:"versions"`
}

// NewPluginInfo parses a JSON input into a valid PluginInfo struct
// Returns an error if there are any missing fields or any invalid fields
// There are no optional fields in a plugin info json file
func NewPluginInfo(data []byte) (*PluginInfo, error) {
	var info PluginInfo

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // Force errors

	if err := dec.Decode(&info); err != nil {
		return nil, err
	}

	if len(info.Name) == 0 {
		return nil, ErrMissingName
	} else if len(info.Desc) == 0 {
		return nil, ErrMissingDesc
	} else if len(info.Site) == 0 {
		return nil, ErrMissingSite
	} else if len(info.Repo) == 0 {
		return nil, ErrMissingRepo
	} else if err := info.makeVersions(); err != nil {
		return nil, err
	}

	return &info, nil
}

func (i *PluginInfo) makeVersions() error {
	if len(i.Versions) == 0 {
		return ErrMissingVersions
	}

	for _, v := range i.Versions {
		sv, err := semver.Make(v.Vstr)
		if err != nil {
			return err
		}
		v.Vers = sv
		if len(v.Tag) == 0 {
			return ErrMissingTag
		} else if v.Require == nil {
			return ErrMissingRequire
		}
	}

	return nil
}

// ListInstalledPlugins searches the config directory for all installed plugins
// and returns the list of plugin infos corresponding to them
func ListInstalledPlugins() ([]*Plugin, error) {
	pdir := path.Join(config.ConfigDir, "plugin")

	files, err := ioutil.ReadDir(pdir)
	if err != nil {
		return nil, err
	}

	var plugins []*Plugin

	for _, dir := range files {
		if dir.IsDir() {
			files, err := ioutil.ReadDir(path.Join(pdir, dir.Name()))
			if err != nil {
				return nil, err
			}

			for _, f := range files {
				if f.Name() == infojson {
					dat, err := ioutil.ReadFile(path.Join(pdir, dir.Name(), infojson))
					if err != nil {
						return nil, err
					}
					info, err := NewPluginInfo(dat)
					if err != nil {
						return nil, err
					}

					versiondat, err := ioutil.ReadFile(path.Join(pdir, dir.Name(), versionfile))
					if err != nil {
						return nil, err
					}
					sv, err := semver.Make(string(versiondat))
					if err != nil {
						return nil, err
					}

					dirname := path.Join(pdir, dir.Name())
					r, err := git.PlainOpen(dirname)
					if err != nil {
						return nil, err
					}

					p := &Plugin{
						info:    info,
						dir:     dirname,
						repo:    r,
						version: sv,
					}

					plugins = append(plugins, p)
				}
			}
		}
	}
	return plugins, nil
}
