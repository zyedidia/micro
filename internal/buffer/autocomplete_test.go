package buffer

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/zyedidia/micro/v2/internal/config"
)

func setupTestDir() (string, error) {
	dir, err := ioutil.TempDir("", "testdir")
	if err != nil {
		return "", err
	}

	subdirs := []string{"subdir1", "subdir2"}
	files := []string{"file1.txt", "file2.txt"}

	for _, subdir := range subdirs {
		if err := os.Mkdir(dir+string(os.PathSeparator)+subdir, 0755); err != nil {
			return "", err
		}
	}

	for _, file := range files {
		if _, err := os.Create(dir + string(os.PathSeparator) + file); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func init() {
	config.InitRuntimeFiles(false)
	config.InitGlobalSettings()
	config.GlobalSettings["backup"] = false
	config.GlobalSettings["fastdirty"] = true
}

func teardownTestDir(dir string) {
	os.RemoveAll(dir)
}

func TestFileComplete(t *testing.T) {
	dir, err := setupTestDir()
	if err != nil {
		t.Fatalf("Failed to set up test directory: %v", err)
	}
	defer teardownTestDir(dir)

	tests := []struct {
		name         string
		input        string
		expectedComp []string
		expectedSugg []string
		expectedErr  bool
	}{
		{
			name:         "Complete subdir",
			input:        dir + string(os.PathSeparator) + "sub",
			expectedComp: []string{"dir1" + string(os.PathSeparator), "dir2" + string(os.PathSeparator)},
			expectedSugg: []string{"subdir1" + string(os.PathSeparator), "subdir2" + string(os.PathSeparator)},
		},
		{
			name:         "Complete file",
			input:        dir + string(os.PathSeparator) + "f",
			expectedComp: []string{"ile1.txt", "ile2.txt"},
			expectedSugg: []string{"file1.txt", "file2.txt"},
		},
		{
			name:         "No match",
			input:        dir + string(os.PathSeparator) + "nomatch",
			expectedComp: []string{},
			expectedSugg: []string{},
		},
		{
			name:         "Complete subdir inside quotes",
			input:        "\"" + dir + string(os.PathSeparator) + "sub",
			expectedComp: []string{"dir1" + string(os.PathSeparator), "dir2" + string(os.PathSeparator)},
			expectedSugg: []string{"subdir1" + string(os.PathSeparator), "subdir2" + string(os.PathSeparator)},
		},
		{
			name:         "Complete inside function and quotes",
			input:        "funtion(\"" + dir + string(os.PathSeparator) + "sub",
			expectedComp: []string{"dir1" + string(os.PathSeparator), "dir2" + string(os.PathSeparator)},
			expectedSugg: []string{"subdir1" + string(os.PathSeparator), "subdir2" + string(os.PathSeparator)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBufferFromString(tt.input, "", BTDefault)
			c := b.GetActiveCursor()
			c.End()

			completions, suggestions := FileComplete(b)
			if len(completions) != len(tt.expectedComp) {
				t.Errorf("expected %d completions, got %d", len(tt.expectedComp), len(completions))
			}
			for i, exp := range tt.expectedComp {
				if completions[i] != exp {
					t.Errorf("expected completion %s, got %s", exp, completions[i])
				}
			}
			if len(suggestions) != len(tt.expectedSugg) {
				t.Errorf("expected %d suggestions, got %d", len(tt.expectedSugg), len(suggestions))
			}
			for i, exp := range tt.expectedSugg {
				if suggestions[i] != exp {
					t.Errorf("expected suggestion %s, got %s", exp, suggestions[i])
				}
			}
		})
	}
}
