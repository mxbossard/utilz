package fz

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/mxbossard/utilz/filez"
)

func Exists(path string) (bool, error) {
	return filez.Exists(path)
}

func ExistsOrPanic(path string) bool {
	ok, err := Exists(path)
	if err != nil {
		panic(err)
	}
	return ok
}

func IsSymLink(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	isSymlink := fileInfo.Mode()&os.ModeSymlink != 0
	return isSymlink, nil
}

func IsSymLinkOrPanic(path string) bool {
	ok, err := IsSymLink(path)
	if err != nil {
		panic(err)
	}
	return ok
}

func CopySymLink(src, dst string, preserve bool) error {
	if preserve {
		exists, err := Exists(dst)
		if err != nil {
			return err
		}
		if exists {
			// Preserve existing file
			return nil
		}
	}
	link, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(link, dst)
}

func CopySymLinkOrPanic(src, dst string, preserve bool) {
	err := CopySymLink(src, dst, preserve)
	if err != nil {
		panic(err)
	}
}

// Copy the file from src to dst.
// If preserve is true do not override dst file if it already exists.
// Copy permissions and ownership of file
func CopyFile(src, dst string, preserve bool) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	var out *os.File
	exists := false
	if preserve {
		exists, err = filez.Exists(dst)
		if err != nil {
			return err
		}
		if exists {
			// Preserve existing file
			return nil
		}
	}

	out, err = os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	err = CopyPermsAndOwnership(src, dst)

	return
}

func CopyFileOrPanic(src, dst string, preserve bool) {
	err := CopyFile(src, dst, preserve)
	if err != nil {
		panic(err)
	}
}

func CopyPermsAndOwnership(src, dst string) (err error) {
	fileInfo, err := os.Stat(src)
	if err != nil {
		return
	}
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("unable to get syscall.Stat_t for file: [%s]", src)
	}

	err = os.Chmod(dst, fileInfo.Mode())
	if err != nil {
		return
	}
	err = os.Lchown(dst, int(stat.Uid), int(stat.Gid))
	return
}

func CopyPermsAndOwnershipOrPanic(src, dst string) {
	err := CopyPermsAndOwnership(src, dst)
	if err != nil {
		panic(err)
	}
}

// Copy the dir from src to dst.
// Src dir MUST exists
// Dst dir may not exists
// If preserve is true do not override dst files if they already exists.
// Copy permissions and ownership of dirs & files
// Copy symlinks
func CopyDir(srcDir, dst string, preserve bool) error {
	dstExists, err := Exists(dst)
	if err != nil {
		return err
	}
	if !dstExists || !preserve {
		err = filez.MkdirAll(dst, filez.DefaultDirPerms)
		if err != nil {
			return err
		}
		CopyPermsAndOwnership(srcDir, dst)
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		fileInfo, err := os.Stat(srcPath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CopyDir(srcPath, dstPath, preserve); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(srcPath, dstPath, preserve); err != nil {
				return err
			}
		default:
			if err := CopyFile(srcPath, dstPath, preserve); err != nil {
				return err
			}
		}
	}
	return nil
}

func CopyDirOrPanic(srcDir, dst string, preserve bool) {
	err := CopyDir(srcDir, dst, preserve)
	if err != nil {
		panic(err)
	}
}
