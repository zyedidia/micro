package encoding

import (
	"fmt"
	"io"
	"strings"
)

var registry []Entry

// Add adds an entry to the registry
func Add(entry Entry) {
	registry = append(registry, entry)
}

// Entry is an entry in the registry
type Entry struct {
	Extensions []string
	Settings   []string
	Encoding   Encoding
}

// Matches return true if the encoding matches
func (e Entry) Matches(extension string, settings map[string]interface{}) bool {
	matches := false
	for _, e := range e.Extensions {
		if e == extension {
			matches = true
			break
		}
	}
	if !matches {
		return false
	}
	for _, s := range e.Settings {
		if _, ok := settings[s]; !ok {
			return false
		}
	}
	return true
}

// Encoding is a type of encoding
type Encoding interface {
	Encode(writer io.WriteCloser, settings map[string]interface{}) (io.WriteCloser, error)
	Decode(reader io.Reader, settings map[string]interface{}) (io.Reader, error)
}

// Encoder builds an encoder for a file name
func Encoder(writer io.WriteCloser, name string, settings map[string]interface{}) (io.WriteCloser, error) {
	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		return writer, nil
	}
	var chain []Encoding
search:
	for _, part := range parts[1:] {
		for _, entry := range registry {
			if !entry.Matches(part, settings) {
				if len(chain) > 0 {
					return writer, fmt.Errorf("%s format is unsupported", part)
				}
				return writer, nil
			}
			chain = append(chain, entry.Encoding)
			continue search
		}
	}
	for _, encoding := range chain {
		var err error
		writer, err = encoding.Encode(writer, settings)
		if err != nil {
			return writer, err
		}
	}
	return writer, nil
}

// Decoder builds an dencoder for a file name
func Decoder(reader io.Reader, name string, settings map[string]interface{}) (io.Reader, error) {
	parts := strings.Split(name, ".")
	length := len(parts)
	if length < 2 {
		return reader, nil
	}
	var chain []Encoding
search:
	for i := range parts[1:] {
		part := parts[length-1-i]
		for _, entry := range registry {
			if !entry.Matches(part, settings) {
				if len(chain) > 0 {
					return reader, fmt.Errorf("%s format is unsupported", part)
				}
				return reader, nil
			}
			chain = append(chain, entry.Encoding)
			continue search
		}
	}
	for _, encoding := range chain {
		var err error
		reader, err = encoding.Decode(reader, settings)
		if err != nil {
			return reader, err
		}
	}
	return reader, nil
}
