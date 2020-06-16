// +build ignore

// This script generates the embedded runtime filesystem, and also creates
// syntax header metadata which makes loading syntax files at runtime faster
// Invoke as go run runtime_generate.go syntaxDir runtimeDir

package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/shurcooL/vfsgen"
)

type HeaderYaml struct {
	FileType string `yaml:"filetype"`
	Detect   struct {
		FNameRgx  string `yaml:"filename"`
		HeaderRgx string `yaml:"header"`
	} `yaml:"detect"`
}

type Header struct {
	FileType  string
	FNameRgx  string
	HeaderRgx string
}

func convert(name string) {
	filename := name + ".yaml"
	var hdr HeaderYaml
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &hdr)
	if err != nil {
		panic(err)
	}
	encode(name, hdr)
}

func encode(name string, c HeaderYaml) {
	f, _ := os.Create(name + ".hdr")
	f.WriteString(c.FileType + "\n")
	f.WriteString(c.Detect.FNameRgx + "\n")
	f.WriteString(c.Detect.HeaderRgx + "\n")
	f.Close()
}

func decode(name string) Header {
	data, _ := ioutil.ReadFile(name + ".hdr")
	strs := bytes.Split(data, []byte{'\n'})
	var hdr Header
	hdr.FileType = string(strs[0])
	hdr.FNameRgx = string(strs[1])
	hdr.HeaderRgx = string(strs[2])

	return hdr
}

func main() {
	orig, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't get cwd")
		return
	}
	if len(os.Args) < 2 {
		log.Fatalln("Not enough arguments")
	}

	syntaxDir := os.Args[1]
	assetDir := os.Args[2]

	os.Chdir(syntaxDir)
	files, _ := ioutil.ReadDir(".")

	// first remove all existing header files (clean the directory)
	for _, f := range files {
		fname := f.Name()
		if strings.HasSuffix(fname, ".hdr") {
			os.Remove(fname)
		}
	}

	// now create a header file for each yaml
	for _, f := range files {
		fname := f.Name()
		if strings.HasSuffix(fname, ".yaml") {
			convert(fname[:len(fname)-5])
		}
	}

	// create the assets_vfsdata.go file for embedding in the binary
	os.Chdir(orig)

	var assets http.FileSystem = http.Dir(assetDir)
	err = vfsgen.Generate(assets, vfsgen.Options{
		PackageName:  "config",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
