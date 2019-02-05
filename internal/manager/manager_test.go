package manager

import (
	"testing"

	"github.com/blang/semver"
	"github.com/zyedidia/micro/internal/config"
	"github.com/zyedidia/micro/internal/util"
)

func init() {
	config.InitConfigDir("./")
	util.Version = "1.3.1"
	util.SemVersion, _ = semver.Make(util.Version)
}

var sampleJson = []byte(`{
    "name": "comment",
    "description": "Plugin to auto comment or uncomment lines",
    "website": "https://github.com/micro-editor/comment-plugin",
	"repository": "https://github.com/micro-editor/comment-plugin",
    "versions": [
        {
            "version": "1.0.6",
            "tag": "v1.0.6",
            "require": {
                "micro": ">=1.1.0"
            }
        },
        {
            "version": "1.0.5",
            "tag": "v1.0.5",
            "require": {
                "micro": ">=1.0.0"
            }
        },
        {
            "version": "1.0.6-dev",
            "tag": "nightly",
            "require": {
                "micro": ">=1.3.1"
            }
        }
    ]
}`)

func TestParse(t *testing.T) {
	_, err := NewPluginInfo(sampleJson)
	if err != nil {
		t.Error(err)
	}
}

func TestFetch(t *testing.T) {
	i, err := NewPluginInfoFromUrl("http://zbyedidia.webfactional.com/micro/test.json")
	if err != nil {
		t.Error(err)
	}

	err = i.FetchRepo()
	if err != nil {
		t.Error(err)
	}
}

// func TestList(t *testing.T) {
// 	is, err := ListInstalledPlugins()
// 	if err != nil {
// 		t.Error(err)
// 	}
//
// 	for _, i := range is {
// 		fmt.Println(i.dir)
// 	}
// }
