package main

import (
	"errors"
	"io"
	"strings"
	"testing"
)

type nopCloseReader struct {
	reader io.Reader
}
func (ncr nopCloseReader) Read(p []byte) (n int, err error) {
	return ncr.reader.Read(p)
}
func (ncr nopCloseReader) Close() error {
	return nil
}
func newStringReadCloser(in string) nopCloseReader {
	return nopCloseReader{strings.NewReader(in)}
}

type stubFilesystem struct{
	openFunks map[string]func() (io.ReadCloser, error)
	statErrors map[string]error
}
func (fs stubFilesystem) Open(name string) (io.ReadCloser, error) {
	return fs.openFunks[name]()
}
func (fs stubFilesystem) Stat(name string) (interface{}, error) {
	return nil, fs.statErrors[name]
}

func TestLocateBindingsOnYamlFilePassReturnsYamlAndTrue(t *testing.T) {
	fs := stubFilesystem{}
	fs.statErrors = make(map[string]error)
	res, found := locateBindingsFile(fs, "config")
	assertTrue(t, found)
	assertEqual(t, "config/bindings.yaml", res.Path)
}

func TestLocateBindingsOnYamlFileFailWithJsonPassReturnsJsonAndTrue(t *testing.T) {
	fs := stubFilesystem{}
	fs.statErrors = map[string]error{
		"config/bindings.yaml": errors.New("Anything"),
	}
	res, found := locateBindingsFile(fs, "config")
	assertTrue(t, found)
	assertEqual(t, "config/bindings.json", res.Path)
}

func TestLocateBindingsOnNoFilePassReturnsNilAndFalse(t *testing.T) {
	fs := stubFilesystem{}
	fs.statErrors = map[string]error{
		"config/bindings.yaml": errors.New("Something"),
		"config/bindings.json": errors.New("Something Else"),
	}
	_, found := locateBindingsFile(fs, "config")
	assertEqual(t, false, found)
}

func TestBindingsFileLoadOnOpenErrorReturnsEmptyMap(t *testing.T) {
	fs := stubFilesystem{}
	fs.openFunks = map[string]func() (io.ReadCloser, error){
		"config/bindings.yaml": func() (io.ReadCloser, error) {
			return nil, errors.New("Failed on Open")
		},
	}
	res := bindingsFile{fs, "config/bindings.yaml"}.Load()
	assertEqual(t, 0, len(res))
}

func TestBindingsFileLoadOnMalformedFileReturnsEmptyMap(t *testing.T) {
	fs := stubFilesystem{}
	fs.openFunks = map[string]func() (io.ReadCloser, error) {
		"config/bindings.yaml": func() (io.ReadCloser, error) {
			return newStringReadCloser("{[Malformed JSON]}"), nil			
		},
	}
	res := bindingsFile{fs, "config/bindings.yaml"}.Load()
	assertEqual(t, 0, len(res))
}

func TestBindingsFileLoadOnValidYAMLReturnsMap(t *testing.T) {
	fs := stubFilesystem{}
	fs.openFunks = map[string]func() (io.ReadCloser, error) {
		"config/bindings.yaml": func() (io.ReadCloser, error) {
			return newStringReadCloser("CtrlN: CursorDown\nCtrlP: CursorUp"), nil
		},
	}
	res := bindingsFile{fs, "config/bindings.yaml"}.Load()
	assertEqual(t, 2, len(res))
	assertEqual(t, "CursorDown", res["CtrlN"])
	assertEqual(t, "CursorUp", res["CtrlP"])
}

func TestBindingsFileLoadOnValidJSONReturnsMap(t *testing.T) {
	fs := stubFilesystem{}
	fs.openFunks = map[string]func() (io.ReadCloser, error) {
		"config/bindings.yaml": func() (io.ReadCloser, error) {
			return newStringReadCloser(`{"Up": "CursorUp", "ShiftUp": "SelectUp"}`), nil
		},
	}
	res := bindingsFile{fs, "config/bindings.yaml"}.Load()
	assertEqual(t, 2, len(res))
	assertEqual(t, "CursorUp", res["Up"])
	assertEqual(t, "SelectUp", res["ShiftUp"])
}
