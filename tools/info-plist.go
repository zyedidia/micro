//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	if runtime.GOOS != "darwin" {
		return
	}
	if len(os.Args) == 3 {
		if os.Args[1] == "darwin" && runtime.GOOS == "darwin" {
			rawInfoPlistString := `<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version="1.0">
	<dict>
		<key>CFBundleIdentifier</key>
		<string>io.github.micro-editor</string>
		<key>CFBundleName</key>
		<string>micro</string>
		<key>CFBundleInfoDictionaryVersion</key>
		<string>6.0</string>
		<key>CFBundlePackageType</key>
		<string>APPL</string>
		<key>CFBundleShortVersionString</key>
		<string>` + os.Args[2] + `</string>
	</dict>
	</plist>
	`
			infoPlistData := []byte(rawInfoPlistString)

			err := ioutil.WriteFile("/tmp/micro-info.plist", infoPlistData, 0644)
			check(err)
			fmt.Println("-linkmode external -extldflags -Wl,-sectcreate,__TEXT,__info_plist,/tmp/micro-info.plist")
		}
	} else {
		panic("missing arguments")
	}
}
