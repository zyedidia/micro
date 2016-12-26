package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/blang/semver"
	"github.com/yuin/gopher-lua"
	"github.com/zyedidia/json5/encoding/json5"
)

var (
	allPluginPackages PluginPackages
)

// CorePluginName is a plugin dependency name for the micro core.
const CorePluginName = "micro"

// PluginChannel contains an url to a json list of PluginRepository
type PluginChannel string

// PluginChannels is a slice of PluginChannel
type PluginChannels []PluginChannel

// PluginRepository contains an url to json file containing PluginPackages
type PluginRepository string

// PluginPackage contains the meta-data of a plugin and all available versions
type PluginPackage struct {
	Name        string
	Description string
	Author      string
	Tags        []string
	Versions    PluginVersions
}

// PluginPackages is a list of PluginPackage instances.
type PluginPackages []*PluginPackage

// PluginVersion descripes a version of a PluginPackage. Containing a version, download url and also dependencies.
type PluginVersion struct {
	pack    *PluginPackage
	Version semver.Version
	Url     string
	Require PluginDependencies
}

// PluginVersions is a slice of PluginVersion
type PluginVersions []*PluginVersion

// PluginDependency descripes a dependency to another plugin or micro itself.
type PluginDependency struct {
	Name  string
	Range semver.Range
}

// PluginDependencies is a slice of PluginDependency
type PluginDependencies []*PluginDependency

func (pp *PluginPackage) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString("Plugin: ")
	buf.WriteString(pp.Name)
	buf.WriteRune('\n')
	if pp.Author != "" {
		buf.WriteString("Author: ")
		buf.WriteString(pp.Author)
		buf.WriteRune('\n')
	}
	if pp.Description != "" {
		buf.WriteRune('\n')
		buf.WriteString(pp.Description)
	}
	return buf.String()
}

func fetchAllSources(count int, fetcher func(i int) PluginPackages) PluginPackages {
	wgQuery := new(sync.WaitGroup)
	wgQuery.Add(count)

	results := make(chan PluginPackages)

	wgDone := new(sync.WaitGroup)
	wgDone.Add(1)
	var packages PluginPackages
	for i := 0; i < count; i++ {
		go func(i int) {
			results <- fetcher(i)
			wgQuery.Done()
		}(i)
	}
	go func() {
		packages = make(PluginPackages, 0)
		for res := range results {
			packages = append(packages, res...)
		}
		wgDone.Done()
	}()
	wgQuery.Wait()
	close(results)
	wgDone.Wait()
	return packages
}

// Fetch retrieves all available PluginPackages from the given channels
func (pc PluginChannels) Fetch() PluginPackages {
	return fetchAllSources(len(pc), func(i int) PluginPackages {
		return pc[i].Fetch()
	})
}

// Fetch retrieves all available PluginPackages from the given channel
func (pc PluginChannel) Fetch() PluginPackages {
	// messenger.AddLog(fmt.Sprintf("Fetching channel: %q", string(pc)))
	resp, err := http.Get(string(pc))
	if err != nil {
		TermMessage("Failed to query plugin channel:\n", err)
		return PluginPackages{}
	}
	defer resp.Body.Close()
	decoder := json5.NewDecoder(resp.Body)

	var repositories []PluginRepository
	if err := decoder.Decode(&repositories); err != nil {
		TermMessage("Failed to decode channel data:\n", err)
		return PluginPackages{}
	}
	return fetchAllSources(len(repositories), func(i int) PluginPackages {
		return repositories[i].Fetch()
	})
}

// Fetch retrieves all available PluginPackages from the given repository
func (pr PluginRepository) Fetch() PluginPackages {
	// messenger.AddLog(fmt.Sprintf("Fetching repository: %q", string(pr)))
	resp, err := http.Get(string(pr))
	if err != nil {
		TermMessage("Failed to query plugin repository:\n", err)
		return PluginPackages{}
	}
	defer resp.Body.Close()
	decoder := json5.NewDecoder(resp.Body)

	var plugins PluginPackages
	if err := decoder.Decode(&plugins); err != nil {
		TermMessage("Failed to decode repository data:\n", err)
		return PluginPackages{}
	}
	if len(plugins) > 0 {
		return PluginPackages{plugins[0]}
	}
	return nil
	// return plugins
}

// UnmarshalJSON unmarshals raw json to a PluginVersion
func (pv *PluginVersion) UnmarshalJSON(data []byte) error {
	var values struct {
		Version semver.Version
		Url     string
		Require map[string]string
	}

	if err := json5.Unmarshal(data, &values); err != nil {
		return err
	}
	pv.Version = values.Version
	pv.Url = values.Url
	pv.Require = make(PluginDependencies, 0)

	for k, v := range values.Require {
		// don't add the dependency if it's the core and
		// we have a unknown version number.
		// in that case just accept that dependency (which equals to not adding it.)
		if k != CorePluginName || !isUnknownCoreVersion() {
			if vRange, err := semver.ParseRange(v); err == nil {
				pv.Require = append(pv.Require, &PluginDependency{k, vRange})
			}
		}
	}
	return nil
}

// UnmarshalJSON unmarshals raw json to a PluginPackage
func (pp *PluginPackage) UnmarshalJSON(data []byte) error {
	var values struct {
		Name        string
		Description string
		Author      string
		Tags        []string
		Versions    PluginVersions
	}
	if err := json5.Unmarshal(data, &values); err != nil {
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

// GetAllPluginPackages gets all PluginPackages which may be available.
func GetAllPluginPackages() PluginPackages {
	if allPluginPackages == nil {
		getOption := func(name string) []string {
			data := GetOption(name)
			if strs, ok := data.([]string); ok {
				return strs
			}
			if ifs, ok := data.([]interface{}); ok {
				result := make([]string, len(ifs))
				for i, urlIf := range ifs {
					if url, ok := urlIf.(string); ok {
						result[i] = url
					} else {
						return nil
					}
				}
				return result
			}
			return nil
		}

		channels := PluginChannels{}
		for _, url := range getOption("pluginchannels") {
			channels = append(channels, PluginChannel(url))
		}
		repos := []PluginRepository{}
		for _, url := range getOption("pluginrepos") {
			repos = append(repos, PluginRepository(url))
		}
		allPluginPackages = fetchAllSources(len(repos)+1, func(i int) PluginPackages {
			if i == 0 {
				return channels.Fetch()
			}
			return repos[i-1].Fetch()
		})
	}
	return allPluginPackages
}

func (pv PluginVersions) find(ppName string) *PluginVersion {
	for _, v := range pv {
		if v.pack.Name == ppName {
			return v
		}
	}
	return nil
}

// Len returns the number of pluginversions in this slice
func (pv PluginVersions) Len() int {
	return len(pv)
}

// Swap two entries of the slice
func (pv PluginVersions) Swap(i, j int) {
	pv[i], pv[j] = pv[j], pv[i]
}

// Less returns true if the version at position i is greater then the version at position j (used for sorting)
func (pv PluginVersions) Less(i, j int) bool {
	return pv[i].Version.GT(pv[j].Version)
}

// Match returns true if the package matches a given search text
func (pp PluginPackage) Match(text string) bool {
	text = strings.ToLower(text)
	for _, t := range pp.Tags {
		if strings.ToLower(t) == text {
			return true
		}
	}
	if strings.Contains(strings.ToLower(pp.Name), text) {
		return true
	}

	if strings.Contains(strings.ToLower(pp.Description), text) {
		return true
	}

	return false
}

// IsInstallable returns true if the package can be installed.
func (pp PluginPackage) IsInstallable() error {
	_, err := GetAllPluginPackages().Resolve(GetInstalledVersions(true), PluginDependencies{
		&PluginDependency{
			Name:  pp.Name,
			Range: semver.Range(func(v semver.Version) bool { return true }),
		}})
	return err
}

// SearchPlugin retrieves a list of all PluginPackages which match the given search text and
// could be or are already installed
func SearchPlugin(texts []string) (plugins PluginPackages) {
	plugins = make(PluginPackages, 0)

pluginLoop:
	for _, pp := range GetAllPluginPackages() {
		for _, text := range texts {
			if !pp.Match(text) {
				continue pluginLoop
			}
		}

		if err := pp.IsInstallable(); err == nil {
			plugins = append(plugins, pp)
		}
	}
	return
}

func isUnknownCoreVersion() bool {
	_, err := semver.ParseTolerant(Version)
	return err != nil
}

func newStaticPluginVersion(name, version string) *PluginVersion {
	vers, err := semver.ParseTolerant(version)

	if err != nil {
		if vers, err = semver.ParseTolerant("0.0.0-" + version); err != nil {
			vers = semver.MustParse("0.0.0-unknown")
		}
	}
	pl := &PluginPackage{
		Name: name,
	}
	pv := &PluginVersion{
		pack:    pl,
		Version: vers,
	}
	pl.Versions = PluginVersions{pv}
	return pv
}

// GetInstalledVersions returns a list of all currently installed plugins including an entry for
// micro itself. This can be used to resolve dependencies.
func GetInstalledVersions(withCore bool) PluginVersions {
	result := PluginVersions{}
	if withCore {
		result = append(result, newStaticPluginVersion(CorePluginName, Version))
	}

	for name, lpname := range loadedPlugins {
		version := GetInstalledPluginVersion(lpname)
		if pv := newStaticPluginVersion(name, version); pv != nil {
			result = append(result, pv)
		}
	}

	return result
}

// GetInstalledPluginVersion returns the string of the exported VERSION variable of a loaded plugin
func GetInstalledPluginVersion(name string) string {
	plugin := L.GetGlobal(name)
	if plugin != lua.LNil {
		version := L.GetField(plugin, "VERSION")
		if str, ok := version.(lua.LString); ok {
			return string(str)

		}
	}
	return ""
}

// DownloadAndInstall downloads and installs the given plugin and version
func (pv *PluginVersion) DownloadAndInstall() error {
	messenger.AddLog(fmt.Sprintf("Downloading %q (%s) from %q", pv.pack.Name, pv.Version, pv.Url))
	resp, err := http.Get(pv.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	zipbuf := bytes.NewReader(data)
	z, err := zip.NewReader(zipbuf, zipbuf.Size())
	if err != nil {
		return err
	}
	targetDir := filepath.Join(configDir, "plugins", pv.pack.Name)
	dirPerm := os.FileMode(0755)
	if err = os.MkdirAll(targetDir, dirPerm); err != nil {
		return err
	}

	// Check if all files in zip are in the same directory.
	// this might be the case if the plugin zip contains the whole plugin dir
	// instead of its content.
	var prefix string
	allPrefixed := false
	for i, f := range z.File {
		parts := strings.Split(f.Name, "/")
		if i == 0 {
			prefix = parts[0]
		} else if parts[0] != prefix {
			allPrefixed = false
			break
		} else {
			// switch to true since we have at least a second file
			allPrefixed = true
		}
	}

	for _, f := range z.File {
		parts := strings.Split(f.Name, "/")
		if allPrefixed {
			parts = parts[1:]
		}

		targetName := filepath.Join(targetDir, filepath.Join(parts...))
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetName, dirPerm); err != nil {
				return err
			}
		} else {
			content, err := f.Open()
			if err != nil {
				return err
			}
			defer content.Close()
			target, err := os.Create(targetName)
			if err != nil {
				return err
			}
			defer target.Close()
			if _, err = io.Copy(target, content); err != nil {
				return err
			}
		}
	}
	return nil
}

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

// Resolve resolves dependencies between different plugins
func (all PluginPackages) Resolve(selectedVersions PluginVersions, open PluginDependencies) (PluginVersions, error) {
	if len(open) == 0 {
		return selectedVersions, nil
	}
	currentRequirement, stillOpen := open[0], open[1:]
	if currentRequirement != nil {
		if selVersion := selectedVersions.find(currentRequirement.Name); selVersion != nil {
			if currentRequirement.Range(selVersion.Version) {
				return all.Resolve(selectedVersions, stillOpen)
			}
			return nil, fmt.Errorf("unable to find a matching version for \"%s\"", currentRequirement.Name)
		}
		availableVersions := all.GetAllVersions(currentRequirement.Name)
		sort.Sort(availableVersions)

		for _, version := range availableVersions {
			if currentRequirement.Range(version.Version) {
				resolved, err := all.Resolve(append(selectedVersions, version), stillOpen.Join(version.Require))

				if err == nil {
					return resolved, nil
				}
			}
		}
		return nil, fmt.Errorf("unable to find a matching version for \"%s\"", currentRequirement.Name)
	}
	return selectedVersions, nil
}

func (pv PluginVersions) install() {
	anyInstalled := false
	currentlyInstalled := GetInstalledVersions(true)

	for _, sel := range pv {
		if sel.pack.Name != CorePluginName {
			shouldInstall := true
			if pv := currentlyInstalled.find(sel.pack.Name); pv != nil {
				if pv.Version.NE(sel.Version) {
					messenger.AddLog(fmt.Sprint("Uninstalling %q", sel.pack.Name))
					UninstallPlugin(sel.pack.Name)
				} else {
					shouldInstall = false
				}
			}

			if shouldInstall {
				if err := sel.DownloadAndInstall(); err != nil {
					messenger.Error(err)
					return
				}
				anyInstalled = true
			}
		}
	}
	if anyInstalled {
		messenger.Message("One or more plugins installed. Please restart micro.")
	} else {
		messenger.AddLog("Nothing to install / update")
	}
}

// UninstallPlugin deletes the plugin folder of the given plugin
func UninstallPlugin(name string) {
	if err := os.RemoveAll(filepath.Join(configDir, "plugins", name)); err != nil {
		messenger.Error(err)
		return
	}
	delete(loadedPlugins, name)
}

// Install installs the plugin
func (pl PluginPackage) Install() {
	selected, err := GetAllPluginPackages().Resolve(GetInstalledVersions(true), PluginDependencies{
		&PluginDependency{
			Name:  pl.Name,
			Range: semver.Range(func(v semver.Version) bool { return true }),
		}})
	if err != nil {
		TermMessage(err)
		return
	}
	selected.install()
}

// UpdatePlugins updates the given plugins
func UpdatePlugins(plugins []string) {
	// if no plugins are specified, update all installed plugins.
	if len(plugins) == 0 {
		for name := range loadedPlugins {
			plugins = append(plugins, name)
		}
	}

	messenger.AddLog("Checking for plugin updates")
	microVersion := PluginVersions{
		newStaticPluginVersion(CorePluginName, Version),
	}

	var updates = make(PluginDependencies, 0)
	for _, name := range plugins {
		pv := GetInstalledPluginVersion(name)
		r, err := semver.ParseRange(">=" + pv) // Try to get newer versions.
		if err == nil {
			updates = append(updates, &PluginDependency{
				Name:  name,
				Range: r,
			})
		}
	}

	selected, err := GetAllPluginPackages().Resolve(microVersion, updates)
	if err != nil {
		TermMessage(err)
		return
	}
	selected.install()
}
