//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type HeaderYaml struct {
	FileType string `yaml:"filetype"`
	Detect   struct {
		FNameRgx     string `yaml:"filename"`
		HeaderRgx    string `yaml:"header"`
		SignatureRgx string `yaml:"signature"`
	} `yaml:"detect"`
}

type Header struct {
	FileType     string
	FNameRgx     string
	HeaderRgx    string
	SignatureRgx string
}

func main() {
	if len(os.Args) > 1 {
		os.Chdir(os.Args[1])
	}
	files, _ := ioutil.ReadDir(".")
	for _, f := range files {
		fname := f.Name()
		if strings.HasSuffix(fname, ".yaml") {
			convert(fname[:len(fname)-5])
		}
	}
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
	f.WriteString(c.Detect.SignatureRgx + "\n")
	f.Close()
}

func decode(name string) Header {
	start := time.Now()
	data, _ := ioutil.ReadFile(name + ".hdr")
	strs := bytes.Split(data, []byte{'\n'})
	var hdr Header
	hdr.FileType = string(strs[0])
	hdr.FNameRgx = string(strs[1])
	hdr.HeaderRgx = string(strs[2])
	hdr.SignatureRgx = string(strs[3])
	fmt.Printf("took %v\n", time.Since(start))

	return hdr
}
