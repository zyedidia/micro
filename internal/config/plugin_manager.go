package config

import (
	"bytes"
	"encoding/json"
	"errors"
)

var (
	ErrMissingName = errors.New("Missing or empty name field")
	ErrMissingDesc = errors.New("Missing or empty description field")
	ErrMissingSite = errors.New("Missing or empty website field")
)

// PluginInfo contains all the needed info about a plugin
// The info is just strings and are not used beyond that (except
// the Site and Install fields should be valid URLs). This means
// that the requirements for example can be formatted however the
// plugin maker decides, the fields will only be parsed by humans
// Name: name of plugin
// Desc: description of plugin
// Site: home website of plugin
// Install: install link for plugin (can be link to repo or zip file)
// Vstr: version
// Require: list of dependencies and requirements
type PluginInfo struct {
	Name string `json:"Name"`
	Desc string `json:"Description"`
	Site string `json:"Website"`
}

// NewPluginInfo parses a JSON input into a valid PluginInfo struct
// Returns an error if there are any missing fields or any invalid fields
// There are no optional fields in a plugin info json file
func NewPluginInfo(data []byte) (*PluginInfo, error) {
	var info []PluginInfo

	dec := json.NewDecoder(bytes.NewReader(data))
	// dec.DisallowUnknownFields() // Force errors

	if err := dec.Decode(&info); err != nil {
		return nil, err
	}

	return &info[0], nil
}
