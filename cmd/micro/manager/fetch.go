package manager

import (
	"io/ioutil"
	"net/http"
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
	_, err := git.PlainClone(path.Join(config.ConfigDir, "plugin", i.Name), false, &git.CloneOptions{
		URL:      i.Repo,
		Progress: nil,
	})

	return err
}

func (i *PluginInfo) FetchDeps() error {
	return nil
}

func (i *PluginInfo) PostInstallHooks() error {
	return nil
}
