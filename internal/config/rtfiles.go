package config

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	RTColorscheme  = 0
	RTSyntax       = 1
	RTHelp         = 2
	RTPlugin       = 3
	RTSyntaxHeader = 4
)

var (
	NumTypes = 5 // How many filetypes are there
)

type RTFiletype int

// RuntimeFile allows the program to read runtime data like colorschemes or syntax files
type RuntimeFile interface {
	// Name returns a name of the file without paths or extensions
	Name() string
	// Data returns the content of the file.
	Data() ([]byte, error)
}

// allFiles contains all available files, mapped by filetype
var allFiles [][]RuntimeFile
var realFiles [][]RuntimeFile

func init() {
	allFiles = make([][]RuntimeFile, NumTypes)
	realFiles = make([][]RuntimeFile, NumTypes)
}

// NewRTFiletype creates a new RTFiletype
func NewRTFiletype() int {
	NumTypes++
	allFiles = append(allFiles, []RuntimeFile{})
	realFiles = append(realFiles, []RuntimeFile{})
	return NumTypes - 1
}

// some file on filesystem
type realFile string

// some asset file
type assetFile string

// some file on filesystem but with a different name
type namedFile struct {
	realFile
	name string
}

// a file with the data stored in memory
type memoryFile struct {
	name string
	data []byte
}

func (mf memoryFile) Name() string {
	return mf.name
}
func (mf memoryFile) Data() ([]byte, error) {
	return mf.data, nil
}

func (rf realFile) Name() string {
	fn := filepath.Base(string(rf))
	return fn[:len(fn)-len(filepath.Ext(fn))]
}

func (rf realFile) Data() ([]byte, error) {
	return ioutil.ReadFile(string(rf))
}

func (af assetFile) Name() string {
	fn := path.Base(string(af))
	return fn[:len(fn)-len(path.Ext(fn))]
}

func (af assetFile) Data() ([]byte, error) {
	return Asset(string(af))
}

func (nf namedFile) Name() string {
	return nf.name
}

// AddRuntimeFile registers a file for the given filetype
func AddRuntimeFile(fileType RTFiletype, file RuntimeFile) {
	allFiles[fileType] = append(allFiles[fileType], file)
}

// AddRealRuntimeFile registers a file for the given filetype
func AddRealRuntimeFile(fileType RTFiletype, file RuntimeFile) {
	allFiles[fileType] = append(allFiles[fileType], file)
	realFiles[fileType] = append(realFiles[fileType], file)
}

// AddRuntimeFilesFromDirectory registers each file from the given directory for
// the filetype which matches the file-pattern
func AddRuntimeFilesFromDirectory(fileType RTFiletype, directory, pattern string) {
	files, _ := ioutil.ReadDir(directory)
	for _, f := range files {
		if ok, _ := filepath.Match(pattern, f.Name()); !f.IsDir() && ok {
			fullPath := filepath.Join(directory, f.Name())
			AddRealRuntimeFile(fileType, realFile(fullPath))
		}
	}
}

// AddRuntimeFilesFromAssets registers each file from the given asset-directory for
// the filetype which matches the file-pattern
func AddRuntimeFilesFromAssets(fileType RTFiletype, directory, pattern string) {
	files, err := AssetDir(directory)
	if err != nil {
		return
	}
	for _, f := range files {
		if ok, _ := path.Match(pattern, f); ok {
			AddRuntimeFile(fileType, assetFile(path.Join(directory, f)))
		}
	}
}

// FindRuntimeFile finds a runtime file of the given filetype and name
// will return nil if no file was found
func FindRuntimeFile(fileType RTFiletype, name string) RuntimeFile {
	for _, f := range ListRuntimeFiles(fileType) {
		if f.Name() == name {
			return f
		}
	}
	return nil
}

// ListRuntimeFiles lists all known runtime files for the given filetype
func ListRuntimeFiles(fileType RTFiletype) []RuntimeFile {
	return allFiles[fileType]
}

// ListRealRuntimeFiles lists all real runtime files (on disk) for a filetype
// these runtime files will be ones defined by the user and loaded from the config directory
func ListRealRuntimeFiles(fileType RTFiletype) []RuntimeFile {
	return realFiles[fileType]
}

// InitRuntimeFiles initializes all assets file and the config directory
func InitRuntimeFiles() {
	add := func(fileType RTFiletype, dir, pattern string) {
		AddRuntimeFilesFromDirectory(fileType, filepath.Join(ConfigDir, dir), pattern)
		AddRuntimeFilesFromAssets(fileType, path.Join("runtime", dir), pattern)
	}

	add(RTColorscheme, "colorschemes", "*.micro")
	add(RTSyntax, "syntax", "*.yaml")
	add(RTSyntaxHeader, "syntax", "*.hdr")
	add(RTHelp, "help", "*.md")

	initlua := filepath.Join(ConfigDir, "init.lua")
	if _, err := os.Stat(initlua); !os.IsNotExist(err) {
		p := new(Plugin)
		p.Name = "initlua"
		p.DirName = "initlua"
		p.Srcs = append(p.Srcs, realFile(initlua))
		Plugins = append(Plugins, p)
	}

	// Search ConfigDir for plugin-scripts
	plugdir := filepath.Join(ConfigDir, "plug")
	files, _ := ioutil.ReadDir(plugdir)

	isID := regexp.MustCompile(`^[_A-Za-z0-9]+$`).MatchString

	for _, d := range files {
		if d.IsDir() {
			srcs, _ := ioutil.ReadDir(filepath.Join(plugdir, d.Name()))
			p := new(Plugin)
			p.Name = d.Name()
			p.DirName = d.Name()
			for _, f := range srcs {
				if strings.HasSuffix(f.Name(), ".lua") {
					p.Srcs = append(p.Srcs, realFile(filepath.Join(plugdir, d.Name(), f.Name())))
				} else if strings.HasSuffix(f.Name(), ".json") {
					data, err := ioutil.ReadFile(filepath.Join(plugdir, d.Name(), f.Name()))
					if err != nil {
						continue
					}
					p.Info, err = NewPluginInfo(data)
					if err != nil {
						continue
					}
					p.Name = p.Info.Name
				}
			}

			if !isID(p.Name) || len(p.Srcs) <= 0 {
				log.Println(p.Name, "is not a plugin")
				continue
			}
			Plugins = append(Plugins, p)
		}
	}

	plugdir = filepath.Join("runtime", "plugins")
	if files, err := AssetDir(plugdir); err == nil {
		for _, d := range files {
			if srcs, err := AssetDir(filepath.Join(plugdir, d)); err == nil {
				p := new(Plugin)
				p.Name = d
				p.DirName = d
				p.Default = true
				for _, f := range srcs {
					if strings.HasSuffix(f, ".lua") {
						p.Srcs = append(p.Srcs, assetFile(filepath.Join(plugdir, d, f)))
					} else if strings.HasSuffix(f, ".json") {
						data, err := Asset(filepath.Join(plugdir, d, f))
						if err != nil {
							continue
						}
						p.Info, err = NewPluginInfo(data)
						if err != nil {
							continue
						}
						p.Name = p.Info.Name
					}
				}
				if !isID(p.Name) || len(p.Srcs) <= 0 {
					log.Println(p.Name, "is not a plugin")
					continue
				}
				Plugins = append(Plugins, p)
			}
		}
	}
}

// PluginReadRuntimeFile allows plugin scripts to read the content of a runtime file
func PluginReadRuntimeFile(fileType RTFiletype, name string) string {
	if file := FindRuntimeFile(fileType, name); file != nil {
		if data, err := file.Data(); err == nil {
			return string(data)
		}
	}
	return ""
}

// PluginListRuntimeFiles allows plugins to lists all runtime files of the given type
func PluginListRuntimeFiles(fileType RTFiletype) []string {
	files := ListRuntimeFiles(fileType)
	result := make([]string, len(files))
	for i, f := range files {
		result[i] = f.Name()
	}
	return result
}

// PluginAddRuntimeFile adds a file to the runtime files for a plugin
func PluginAddRuntimeFile(plugin string, filetype RTFiletype, filePath string) error {
	pl := FindPlugin(plugin)
	if pl == nil {
		return errors.New("Plugin " + plugin + " does not exist")
	}
	pldir := pl.DirName
	fullpath := filepath.Join(ConfigDir, "plug", pldir, filePath)
	if _, err := os.Stat(fullpath); err == nil {
		AddRealRuntimeFile(filetype, realFile(fullpath))
	} else {
		fullpath = path.Join("runtime", "plugins", pldir, filePath)
		AddRuntimeFile(filetype, assetFile(fullpath))
	}
	return nil
}

// PluginAddRuntimeFilesFromDirectory adds files from a directory to the runtime files for a plugin
func PluginAddRuntimeFilesFromDirectory(plugin string, filetype RTFiletype, directory, pattern string) error {
	pl := FindPlugin(plugin)
	if pl == nil {
		return errors.New("Plugin " + plugin + " does not exist")
	}
	pldir := pl.DirName
	fullpath := filepath.Join(ConfigDir, "plug", pldir, directory)
	if _, err := os.Stat(fullpath); err == nil {
		AddRuntimeFilesFromDirectory(filetype, fullpath, pattern)
	} else {
		fullpath = path.Join("runtime", "plugins", pldir, directory)
		AddRuntimeFilesFromAssets(filetype, fullpath, pattern)
	}
	return nil
}

// PluginAddRuntimeFileFromMemory adds a file to the runtime files for a plugin from a given string
func PluginAddRuntimeFileFromMemory(filetype RTFiletype, filename, data string) {
	AddRealRuntimeFile(filetype, memoryFile{filename, []byte(data)})
}
