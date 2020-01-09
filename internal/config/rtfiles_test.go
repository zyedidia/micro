package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	InitRuntimeFiles()
}

func TestAddFile(t *testing.T) {
	AddRuntimeFile(RTPlugin, memoryFile{"foo.lua", []byte("hello world\n")})
	AddRuntimeFile(RTSyntax, memoryFile{"bar", []byte("some syntax file\n")})

	f1 := FindRuntimeFile(RTPlugin, "foo.lua")
	assert.NotNil(t, f1)
	assert.Equal(t, "foo.lua", f1.Name())
	data, err := f1.Data()
	assert.Nil(t, err)
	assert.Equal(t, []byte("hello world\n"), data)

	f2 := FindRuntimeFile(RTSyntax, "bar")
	assert.NotNil(t, f2)
	assert.Equal(t, "bar", f2.Name())
	data, err = f2.Data()
	assert.Nil(t, err)
	assert.Equal(t, []byte("some syntax file\n"), data)
}

func TestFindFile(t *testing.T) {
	f := FindRuntimeFile(RTSyntax, "go")
	assert.NotNil(t, f)
	assert.Equal(t, "go", f.Name())
	data, err := f.Data()
	assert.Nil(t, err)
	assert.Equal(t, []byte("filetype: go"), data[:12])

	e := FindRuntimeFile(RTSyntax, "foobar")
	assert.Nil(t, e)
}
