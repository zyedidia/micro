package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/zyedidia/highlight"
)

func main() {
	files, _ := ioutil.ReadDir(".")

	hadErr := false
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".yaml") {
			input, _ := ioutil.ReadFile(f.Name())
			//fmt.Println("Checking file -> ", f.Name())
			file, err := highlight.ParseFile(input)
			if err != nil {
				hadErr = true
				fmt.Printf("Could not parse file -> %s:\n", f.Name())
				fmt.Println(err)
				continue
			}
			_, err1 := highlight.ParseDef(file, nil)
			if err1 != nil {
				hadErr = true
				fmt.Printf("Could not parse input file using highlight.ParseDef(%s):\n", f.Name())
				fmt.Println(err1)
				continue
			}
		}
	}
	if !hadErr {
		fmt.Println("No issues found!")
	}
}
