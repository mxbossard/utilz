package trust

import (
	//"fmt"
	"testing"
	"os"
	"path/filepath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mby.fr/utils/test"
)

func assertSignatureOk(t *testing.T, actual string, err error, msg string) {
        assert.NoError(t, err, msg + " signature should not produce an error")
        assert.NotEmpty(t, actual, msg + " signature should not be empty")
}

func assertSameSignature(t *testing.T, expected, actual string, err error, msg string) {
	assertSignatureOk(t, actual, err, msg)
        assert.Equal(t, expected, actual, msg + " signature should stay the same")
}

func assertSignatureDiffer(t *testing.T, expected, actual string, err error, msg string) {
	assertSignatureOk(t, actual, err, msg)
        assert.NotEqual(t, expected, actual, msg + " signature should be different")
}

func TestHashEmptyDir(t *testing.T) {
	path, err := test.MkRandTempDir()
	require.NoError(t, err, "should not error")
	defer os.RemoveAll(path)
	assert.DirExists(t, path, "Temp dir should exists")

	dir1 := filepath.Join(path, "dir1")
	os.Mkdir(dir1, 0755)

	s1, err := SignDirContent(dir1)
	assertSignatureOk(t, s1, err, "empty dir1")

	s2, err := SignDirContent(dir1)
	assertSameSignature(t, s1, s2, err, "empty dir1")

}

func TestHashDir(t *testing.T) {
	path, err := test.MkRandTempDir()
	require.NoError(t, err, "should not error")
	defer os.RemoveAll(path)
	assert.DirExists(t, path, "Temp dir should exists")

	// Empty dir1
	dir1 := filepath.Join(path, "dir1")
	os.Mkdir(dir1, 0755)

	s1, err := SignDirContent(dir1)
	assertSignatureOk(t, s1, err, "empty dir1")

	s2, err := SignDirContent(dir1)
	assertSameSignature(t, s1, s2, err, "empty dir1")

	// Add empty file1
	file1 := filepath.Join(dir1, "file1")
	os.WriteFile(file1, []byte(""), 0644)
	s3, err := SignDirContent(dir1)
	assertSignatureDiffer(t, s1, s3, err, "adding empty file1")

	s4, err := SignDirContent(dir1)
	assertSameSignature(t, s3, s4, err, "adding empty file1")

	// Add not empty file2
	file2 := filepath.Join(dir1, "file2")
	os.WriteFile(file2, []byte("foo"), 0644)
	s5, err := SignDirContent(dir1)
	assertSignatureDiffer(t, s4, s5, err, "adding file2")

	s6, err := SignDirContent(dir1)
	assertSameSignature(t, s5, s6, err, "adding file2")

	// Replace file2 content
	os.WriteFile(file2, []byte("bar"), 0644)
	s7, err := SignDirContent(dir1)
	assertSignatureDiffer(t, s6, s7, err, "modifying file2")

	s8, err := SignDirContent(dir1)
	assertSameSignature(t, s7, s8, err, "modifying file2")

	// Add sub dir2
	dir2 := filepath.Join(dir1, "dir2")
	os.Mkdir(dir2, 0755)
	s9, err := SignDirContent(dir1)
	//assertSignatureDiffer(t, s8, s9, err, "adding sub dir2")
	assertSameSignature(t, s8, s9, err, "adding sub dir2")

	s10, err := SignDirContent(dir1)
	assertSameSignature(t, s7, s10, err, "adding sub dir2")

	// Add not empty file3 in sub dir2
	file3 := filepath.Join(dir2, "file3")
	os.WriteFile(file3, []byte("baz"), 0644)
	s11, err := SignDirContent(dir1)
	assertSignatureDiffer(t, s10, s11, err, "adding filie3")

	s12, err := SignDirContent(dir1)
	assertSameSignature(t, s11, s12, err, "adding file3")

}

