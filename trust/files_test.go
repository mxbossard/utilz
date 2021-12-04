package trust

import (
	//"fmt"
	"testing"
	"os"
	"path/filepath"
	"github.com/stretchr/testify/assert"

	"mby.fr/utils/test"
)

func TestHashEmptyDir(t *testing.T) {
	path := test.MkRandTempDir()
	defer os.RemoveAll(path)
	assert.DirExists(t, path, "Temp dir should exists")

	dir1 := filepath.Join(path, "dir1")
	os.Mkdir(dir1, 0755)

	s1, err := SignDirAll(dir1)
	assert.NoError(t, err, "should not produce an error")
	assert.NotEmpty(t, s1, "s1 should not be empty")

	s2, err := SignDirAll(dir1)
	assert.NoError(t, err, "should not produce an error")
	assert.NotEmpty(t, s2, "s2 should not be empty")
	assert.Equal(t, s1, s2, "signatures should be identical")

}

func TestHashDir(t *testing.T) {
	path := test.MkRandTempDir()
	defer os.RemoveAll(path)
	assert.DirExists(t, path, "Temp dir should exists")

	// Empty dir1
	dir1 := filepath.Join(path, "dir1")
	os.Mkdir(dir1, 0755)

	s1, err := SignDirAll(dir1)
	assert.NoError(t, err, "should not produce an error")
	assert.NotEmpty(t, s1, "s1 should not be empty")

	s2, err := SignDirAll(dir1)
	assert.NoError(t, err, "should not produce an error")
	assert.NotEmpty(t, s2, "s2 should not be empty")
	assert.Equal(t, s1, s2, "signatures should be identical")

	// Add empty file1
	file1 := filepath.Join(dir1, "file1")
	os.WriteFile(file1, []byte(""), 0644)
	s3, err := SignDirAll(dir1)
	assert.NoError(t, err, "should not produce an error")
	assert.NotEmpty(t, s3, "s3 should not be empty")
	assert.Equal(t, s3, s1, "signatures should not be identical")

	s4, err := SignDirAll(dir1)
	assert.NoError(t, err, "should not produce an error")
	assert.NotEmpty(t, s4, "s4 should not be empty")
	assert.Equal(t, s3, s4, "signatures should be identical")

	// Add not empty file2
	file2 := filepath.Join(dir1, "file2")
	os.WriteFile(file2, []byte("foo"), 0644)
	s5, err := SignDirAll(dir1)
	assert.NoError(t, err, "should not produce an error")
	assert.NotEmpty(t, s5, "s5 should not be empty")
	assert.Equal(t, s5, s3, "signatures should not be identical")

	s6, err := SignDirAll(dir1)
	assert.NoError(t, err, "should not produce an error")
	assert.NotEmpty(t, s6, "s6 should not be empty")
	assert.Equal(t, s5, s6, "signatures should be identical")

	// Replace file2 content
	os.WriteFile(file2, []byte("bar"), 0644)
	s7, err := SignDirAll(dir1)
	assert.NoError(t, err, "should not produce an error")
	assert.NotEmpty(t, s7, "s7 should not be empty")
	assert.Equal(t, s7, s5, "signatures should not be identical")

	s8, err := SignDirAll(dir1)
	assert.NoError(t, err, "should not produce an error")
	assert.NotEmpty(t, s8, "s8 should not be empty")
	assert.Equal(t, s7, s8, "signatures should be identical")
}

