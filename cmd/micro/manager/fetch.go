package manager

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/zyedidia/micro/cmd/micro/config"
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
		info: i,
		dir:  dir,
		repo: r,
	}

	err = p.ResolveVersion()
	if err != nil {
		return err
	}
	err = p.WriteVersion()

	return err
}

func (p *Plugin) ResolveVersion() error {
	return nil
}

func (p *Plugin) WriteVersion() error {
	return ioutil.WriteFile(path.Join(p.dir, versionfile), []byte(p.version.String()), os.ModePerm)
}

func (p *Plugin) FetchDeps() error {
	_, err := ListInstalledPlugins()
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) PostInstallHooks() error {
	return nil
}
