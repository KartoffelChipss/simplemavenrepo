package util

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// CheckDirAccess verifies that dir is a directory that can be read and written.
func CheckDirAccess(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	f, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", dir, err)
	}
	_, err = f.ReadDir(1)
	f.Close()
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("cannot read %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".write-test-*")
	if err != nil {
		return fmt.Errorf("cannot write to %s: %w", dir, err)
	}
	name := tmp.Name()
	tmp.Close()
	return os.Remove(name)
}
