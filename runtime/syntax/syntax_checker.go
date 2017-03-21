package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/zyedidia/micro/cmd/micro/highlight"
)

func main() {
	files, _ := ioutil.ReadDir(".")

	hadErr := false
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".yaml") {
			input, _ := ioutil.ReadFile(f.Name())
			_, err := highlight.ParseDef(input)
			if err != nil {
				hadErr = true
				fmt.Printf("%s:\n", f.Name())
				fmt.Println(err)
				continue
			}
		}
	}
	if !hadErr {
		fmt.Println("No issues!")
	}
}
