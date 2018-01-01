package optionprovider

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

// GoCode is an OptionProvider which provides options to the autocompletion system.
func GoCode(buffer []byte, offset int) (options []Option, err error) {
	cmd := exec.Command("gocode", "-f=json", "autocomplete", strconv.Itoa(offset))
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	_, err = stdin.Write(buffer)
	if err != nil {
		return
	}
	err = stdin.Close()
	if err != nil {
		return
	}
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return
	}

	// Unmarshal the JSON, it's an awkward format (mixed array)
	// [1, [ { "class": "", "name": "", "type": "" } ]]
	results := []interface{}{}
	err = json.Unmarshal(stdoutStderr, &results)
	if err != nil {
		err = fmt.Errorf("gocode: failed to unmarshal output '%v': %v", string(stdoutStderr), err)
	}

	// Skip the number.
	if len(results) > 0 {
		if firstElement, isArray := results[1].([]interface{}); isArray {
			results = firstElement
		}
	}

	// Convert the array of maps into something useful.
	for _, r := range results {
		m, mok := r.(map[string]interface{})
		if mok {
			options = append(options, mapToGcr(m))
		}
	}

	return
}

func mapToGcr(m map[string]interface{}) Option {
	gcr := GoCodeResult{}
	if cv, ok := m["class"]; ok {
		if c, cok := cv.(string); cok {
			gcr.Class = c
		}
	}
	if nv, ok := m["name"]; ok {
		if n, nok := nv.(string); nok {
			gcr.Name = n
		}
	}
	if tv, ok := m["type"]; ok {
		if t, tok := tv.(string); tok {
			gcr.Type = t
		}
	}
	if pv, ok := m["package"]; ok {
		if p, pok := pv.(string); pok {
			gcr.Package = p
		}
	}
	return Option(gcr)
}

// GoCodeResult is the JSON output from gocode.
type GoCodeResult struct {
	Class   string `json:"class"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Package string `json:"package"`
}

// Text is the string that will be inserted.
func (gcr GoCodeResult) Text() string {
	return gcr.Name
}

// Hint describes the text, e.g. if it's a method.
func (gcr GoCodeResult) Hint() string {
	return gcr.Type
}
