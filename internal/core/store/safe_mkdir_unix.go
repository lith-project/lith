//go:build !windows

package store

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func safeMkdirAll(path string, mode uint32) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	fd, err := unix.Open(string(filepath.Separator), unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer func() { _ = unix.Close(fd) }()
	for _, component := range strings.Split(strings.TrimPrefix(abs, string(filepath.Separator)), string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}
		next, openErr := unix.Openat(fd, component, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if errors.Is(openErr, unix.ENOENT) {
			if mkdirErr := unix.Mkdirat(fd, component, mode); mkdirErr != nil && !errors.Is(mkdirErr, unix.EEXIST) {
				return fmt.Errorf("mkdir %s: %w", component, mkdirErr)
			}
			next, openErr = unix.Openat(fd, component, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		}
		if openErr != nil {
			if errors.Is(openErr, unix.ELOOP) {
				return ErrDerivedStateInsideVault
			}
			return fmt.Errorf("open %s: %w", component, openErr)
		}
		if err := unix.Close(fd); err != nil {
			return err
		}
		fd = next
	}
	return nil
}
