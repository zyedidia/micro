//go:build ignore

package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	if runtime.GOOS != "darwin" {
		return
	}
	if len(os.Args) != 3 {
		panic("missing arguments")
	}
	if os.Args[1] != "darwin" {
		return
	}
	rawInfoPlist := `<?xml version="1.0" encoding="UTF-8"?>
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

	err := os.WriteFile("/tmp/micro-info.plist", []byte(rawInfoPlist), 0666)
	if err != nil {
		panic(err)
	}
	fmt.Println("-linkmode external -extldflags -Wl,-sectcreate,__TEXT,__info_plist,/tmp/micro-info.plist")
}
