package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/blang/semver"
)

func getTag(match ...string) (string, *semver.PRVersion) {
	args := append([]string{
		"describe", "--tags",
	}, match...)
	if tag, err := exec.Command("git", args...).Output(); err != nil {
		return "", nil
	} else {
		tagParts := strings.Split(string(tag), "-")
		if len(tagParts) == 3 {
			if ahead, err := semver.NewPRVersion(tagParts[1]); err == nil {
				return tagParts[0], &ahead
			}
		}

		return tagParts[0], nil
	}
}

func main() {
	// Find the last vX.X.X Tag and get how many builds we are ahead of it.
	versionStr, ahead := getTag("--match", "v*")
	version, err := semver.ParseTolerant(versionStr)
	if err != nil {
		// no version tag found so just return what ever we can find.
		fmt.Println("0.0.0-unknown")
		return
	}
	// Get the tag of the current revision.
	tag, _ := getTag("--exact-match")
	if tag == versionStr {
		// Seems that we are going to build a release.
		// So the version number should already be correct.
		fmt.Println(version.String())
		return
	}

	// If we don't have any tag assume "dev"
	if tag == "" {
		tag = "dev"
	}
	// Get the most likely next version:
	version.Patch = version.Patch + 1

	if pr, err := semver.NewPRVersion(tag); err == nil {
		// append the tag as pre-release name
		version.Pre = append(version.Pre, pr)
	}

	if ahead != nil {
		// if we know how many commits we are ahead of the last release, append that too.
		version.Pre = append(version.Pre, *ahead)
	}

	fmt.Println(version.String())
}
