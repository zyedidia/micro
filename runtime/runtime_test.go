package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssetDir(t *testing.T) {
	t.Parallel()
	// Test AssetDir
	entries, err := AssetDir("syntax")
	assert.NoError(t, err)
	assert.Contains(t, entries, "go.yaml")
	assert.True(t, len(entries) > 5)
}
