package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetDir(t *testing.T) {
	t.Parallel()
	// Change working directory to runtime if needed
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	if !strings.Contains(cwd, "runtime") {
		err := os.Chdir("./runtime")
		require.NoError(t, err)
	}
	// Test AssetDir
	entries, err := AssetDir("syntax")
	assert.NoError(t, err)
	assert.Contains(t, entries, "go.yaml")
	assert.True(t, len(entries) > 5)
}
