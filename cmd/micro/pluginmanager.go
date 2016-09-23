package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/blang/semver"
	"github.com/yuin/gopher-lua"
)

var Repositories []PluginRepository = []PluginRepository{}

type PluginRepository string

type PluginPackage struct {
	Name        string
	Description string
	Author      string
	Tags        []string
	Versions    PluginVersions
}

type PluginPackages []*PluginPackage

type PluginVersion struct {
	pack    *PluginPackage
	Version semver.Version
	Url     string
	Require PluginDependencies
}
type PluginVersions []*PluginVersion

type PluginDependency struct {
	Name  string
	Range semver.Range
}
type PluginDependencies []*PluginDependency

func (pv *PluginVersion) UnmarshalJSON(data []byte) error {
	var values struct {
		Version semver.Version
		Url     string
		Require map[string]string
	}

	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	pv.Version = values.Version
	pv.Url = values.Url
	pv.Require = make(PluginDependencies, 0)

	for k, v := range values.Require {
		if vRange, err := semver.ParseRange(v); err == nil {
			pv.Require = append(pv.Require, &PluginDependency{k, vRange})
		}
	}
	return nil
}

func (pv *PluginVersion) String() string {
	return fmt.Sprintf("%s (%s)", pv.pack.Name, pv.Version)
}

func (pd *PluginDependency) String() string {
	return pd.Name
}

func (pp *PluginPackage) UnmarshalJSON(data []byte) error {
	var values struct {
		Name        string
		Description string
		Author      string
		Tags        []string
		Versions    PluginVersions
	}
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	pp.Name = values.Name
	pp.Description = values.Description
	pp.Author = values.Author
	pp.Tags = values.Tags
	pp.Versions = values.Versions
	for _, v := range pp.Versions {
		v.pack = pp
	}
	return nil
}

func (pv PluginVersions) Find(name string) *PluginVersion {
	for _, v := range pv {
		if v.pack.Name == name {
			return v
		}
	}
	return nil
}
func (pv PluginVersions) Len() int {
	return len(pv)
}

func (pv PluginVersions) Swap(i, j int) {
	pv[i], pv[j] = pv[j], pv[i]
}

func (s PluginVersions) Less(i, j int) bool {
	// sort descending
	return s[i].Version.GT(s[j].Version)
}

func (pr PluginRepository) Query() <-chan *PluginPackage {
	resChan := make(chan *PluginPackage)
	go func() {
		defer close(resChan)

		resp, err := http.Get(string(pr))
		if err != nil {
			TermMessage("Failed to query plugin repository:\n", err)
			return
		}
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)

		var plugins PluginPackages
		if err := decoder.Decode(&plugins); err != nil {
			TermMessage("Failed to decode repository data:\n", err)
			return
		}
		for _, p := range plugins {
			resChan <- p
		}
	}()
	return resChan
}

func (pp *PluginPackage) GetInstallableVersion() *PluginVersion {
	matching := make(PluginVersions, 0)

versionLoop:
	for _, pv := range pp.Versions {
		for _, req := range pv.Require {
			curVersion := GetInstalledVersion(req.Name)
			if curVersion == nil || !req.Range(*curVersion) {
				continue versionLoop
			}
		}
		matching = append(matching, pv)
	}
	if len(matching) > 0 {
		sort.Sort(matching)
		return matching[0]
	}
	return nil
}

func (pp PluginPackage) Match(text string) bool {
	// ToDo: improve matching.
	text = "(?i)" + text
	if r, err := regexp.Compile(text); err == nil {
		return r.MatchString(pp.Name)
	}
	return false
}

func SearchPlugin(text string) (plugins []*PluginPackage) {
	wgQuery := new(sync.WaitGroup)
	wgQuery.Add(len(Repositories))
	results := make(chan *PluginPackage)

	wgDone := new(sync.WaitGroup)
	wgDone.Add(1)
	for _, repo := range Repositories {
		go func(repo PluginRepository) {
			res := repo.Query()
			for r := range res {
				results <- r
			}
			wgQuery.Done()
		}(repo)
	}
	go func() {
		for res := range results {
			if res.GetInstallableVersion() != nil && res.Match(text) {
				plugins = append(plugins, res)
			}
		}
		wgDone.Done()
	}()
	wgQuery.Wait()
	close(results)
	wgDone.Wait()
	return
}

func GetInstalledVersion(name string) *semver.Version {
	versionStr := ""
	if name == "micro" {
		versionStr = Version

	} else {
		plugin := L.GetGlobal(name)
		if plugin == lua.LNil {
			return nil
		}
		version := L.GetField(plugin, "VERSION")
		if str, ok := version.(lua.LString); ok {
			versionStr = string(str)
		}
	}

	if v, err := semver.Parse(versionStr); err != nil {
		return nil
	} else {
		return &v
	}
}

func (pv *PluginVersion) Install() {
	resp, err := http.Get(pv.Url)
	if err == nil {
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		zipbuf := bytes.NewReader(data)
		z, err := zip.NewReader(zipbuf, zipbuf.Size())
		if err == nil {
			targetDir := filepath.Join(configDir, "plugins", pv.pack.Name)
			dirPerm := os.FileMode(0755)
			if err = os.MkdirAll(targetDir, dirPerm); err == nil {
				for _, f := range z.File {
					targetName := filepath.Join(targetDir, filepath.Join(strings.Split(f.Name, "/")...))
					if f.FileInfo().IsDir() {
						err = os.MkdirAll(targetName, dirPerm)
					} else {
						content, err := f.Open()
						if err == nil {
							defer content.Close()
							if target, err := os.Create(targetName); err == nil {
								defer target.Close()
								_, err = io.Copy(target, content)
							}
						}
					}
					if err != nil {
						break
					}
				}
			}
		}
	}
	if err != nil {
		TermMessage("Failed to install plugin:", err)
	}
}

func UninstallPlugin(name string) {
	os.RemoveAll(filepath.Join(configDir, name))
}

// Updates...

func (pl PluginPackages) Get(name string) *PluginPackage {
	for _, p := range pl {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (pl PluginPackages) GetAllVersions(name string) PluginVersions {
	result := make(PluginVersions, 0)
	p := pl.Get(name)
	if p != nil {
		for _, v := range p.Versions {
			result = append(result, v)
		}
	}
	return result
}

func (req PluginDependencies) Join(other PluginDependencies) PluginDependencies {
	m := make(map[string]*PluginDependency)
	for _, r := range req {
		m[r.Name] = r
	}
	for _, o := range other {
		cur, ok := m[o.Name]
		if ok {
			m[o.Name] = &PluginDependency{
				o.Name,
				o.Range.AND(cur.Range),
			}
		} else {
			m[o.Name] = o
		}
	}
	result := make(PluginDependencies, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

func (all PluginPackages) ResolveStep(selectedVersions PluginVersions, open PluginDependencies) (PluginVersions, error) {
	if len(open) == 0 {
		return selectedVersions, nil
	}
	currentRequirement, stillOpen := open[0], open[1:]
	if currentRequirement != nil {
		if selVersion := selectedVersions.Find(currentRequirement.Name); selVersion != nil {
			if currentRequirement.Range(selVersion.Version) {
				return all.ResolveStep(selectedVersions, stillOpen)
			}
			return nil, fmt.Errorf("unable to find a matching version for \"%s\"", currentRequirement.Name)
		} else {
			availableVersions := all.GetAllVersions(currentRequirement.Name)
			sort.Sort(availableVersions)

			for _, version := range availableVersions {
				if currentRequirement.Range(version.Version) {
					resolved, err := all.ResolveStep(append(selectedVersions, version), stillOpen.Join(version.Require))

					if err == nil {
						return resolved, nil
					}
				}
			}
			return nil, fmt.Errorf("unable to find a matching version for \"%s\"", currentRequirement.Name)
		}
	} else {
		return selectedVersions, nil
	}
}
