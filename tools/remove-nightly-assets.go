//go:build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/micro-editor/json5"
)

func main() {
	resp, err := http.Get("https://api.github.com/repos/zyedidia/micro/releases")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var data interface{}

	err = json5.Unmarshal(body, &data)

	for _, val := range data.([]interface{}) {
		m := val.(map[string]interface{})
		releaseName := m["name"].(string)
		assets := m["assets"].([]interface{})
		for _, asset := range assets {
			assetInfo := asset.(map[string]interface{})
			url := assetInfo["url"].(string)
			if strings.Contains(strings.ToLower(releaseName), "nightly") {
				cmd := exec.Command("hub", "api", "-X", "DELETE", url)
				cmd.Run()
			}
		}
	}
}
