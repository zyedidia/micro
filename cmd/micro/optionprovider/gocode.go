package optionprovider

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

// GoCode is an OptionProvider which provides options to the autocompletion system.
func GoCode(buffer []byte, startOffset, currentOffset int) (options []Option, startOffsetDelta int, err error) {
	cmd := exec.Command("gocode", "-f=json", "autocomplete", strconv.Itoa(currentOffset))
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
			options = append(options, mapToOption(m))
		}
	}

	return
}

func mapToOption(m map[string]interface{}) Option {
	// Available values are "class", "name", "type" and "package"
	o := Option{}
	if nv, ok := m["name"]; ok {
		if n, nok := nv.(string); nok {
			// text
			o.T = n
		}
	}
	if tv, ok := m["type"]; ok {
		if t, tok := tv.(string); tok {
			// hint
			o.H = t
		}
	}
	return o
}
