package vfsutil_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/shurcooL/httpfs/vfsutil"
	"golang.org/x/tools/godoc/vfs/httpfs"
	"golang.org/x/tools/godoc/vfs/mapfs"
)

func ExampleWalk() {
	var fs http.FileSystem = httpfs.New(mapfs.New(map[string]string{
		"zzz-last-file.txt":   "It should be visited last.",
		"a-file.txt":          "It has stuff.",
		"another-file.txt":    "Also stuff.",
		"folderA/entry-A.txt": "Alpha.",
		"folderA/entry-B.txt": "Beta.",
	}))

	walkFn := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Printf("can't stat file %s: %v\n", path, err)
			return nil
		}
		fmt.Println(path)
		return nil
	}

	err := vfsutil.Walk(fs, "/", walkFn)
	if err != nil {
		panic(err)
	}

	// Output:
	// /
	// /a-file.txt
	// /another-file.txt
	// /folderA
	// /folderA/entry-A.txt
	// /folderA/entry-B.txt
	// /zzz-last-file.txt
}

func ExampleWalkFiles() {
	var fs http.FileSystem = httpfs.New(mapfs.New(map[string]string{
		"zzz-last-file.txt":   "It should be visited last.",
		"a-file.txt":          "It has stuff.",
		"another-file.txt":    "Also stuff.",
		"folderA/entry-A.txt": "Alpha.",
		"folderA/entry-B.txt": "Beta.",
	}))

	walkFn := func(path string, fi os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			log.Printf("can't stat file %s: %v\n", path, err)
			return nil
		}
		fmt.Println(path)
		if !fi.IsDir() {
			b, err := ioutil.ReadAll(r)
			if err != nil {
				log.Printf("can't read file %s: %v\n", path, err)
				return nil
			}
			fmt.Printf("%q\n", b)
		}
		return nil
	}

	err := vfsutil.WalkFiles(fs, "/", walkFn)
	if err != nil {
		panic(err)
	}

	// Output:
	// /
	// /a-file.txt
	// "It has stuff."
	// /another-file.txt
	// "Also stuff."
	// /folderA
	// /folderA/entry-A.txt
	// "Alpha."
	// /folderA/entry-B.txt
	// "Beta."
	// /zzz-last-file.txt
	// "It should be visited last."
}
