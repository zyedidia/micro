package manager

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/blang/semver"
	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/util"
	git "gopkg.in/src-d/go-git.v4"
)

// NewPluginInfoFromUrl creates a new PluginInfo from a URL by fetching
// the data at that URL and parsing the JSON (running a GET request at
// the URL should return the JSON for a plugin info)
func NewPluginInfoFromUrl(url string) (*PluginInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	dat, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return NewPluginInfo(dat)
}

// FetchRepo downloads this plugin's git repository
func (i *PluginInfo) FetchRepo() error {
	dir := path.Join(config.ConfigDir, "plugin", i.Name)
	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      i.Repo,
		Progress: nil,
	})

	if err != nil {
		return err
	}

	p := &Plugin{
		Info: i,
		Dir:  dir,
		Repo: r,
	}

	err = p.ResolveVersion()
	if err != nil {
		return err
	}
	err = p.WriteVersion()

	return err
}

func (p *Plugin) ResolveVersion() error {
	i := p.Info
	vs := i.Versions

	for _, v := range vs {
		microrange, err := semver.ParseRange(v.Require["micro"])
		if err != nil {
			return err
		}
		if microrange(util.SemVersion) {
			p.Version = v.Vers
			fmt.Println("resolve version to ", v.Vstr)
			return nil
		}
	}

	return ErrRequireUnsat
}

func (p *Plugin) WriteVersion() error {
	return ioutil.WriteFile(path.Join(p.Dir, versionfile), []byte(p.Version.String()), 0644)
}

func (p *Plugin) PostInstallHooks() error {
	return nil
}
