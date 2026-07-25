// Package scaffold provides idempotent filesystem primitives for
// creating the .agent/ directory layout and tool-specific symlinks.
package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir creates dir (and all parents) if it does not exist. Idempotent.
func EnsureDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}
	return nil
}

// WriteFile writes content to path, creating parent directories as needed.
// If path already exists it is overwritten. Idempotent by content.
func WriteFile(path, content string) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// EnsureSymlink creates a symlink at link pointing to target.
// If link already exists and is already a symlink pointing to target, it is left
// unchanged (idempotent). If link exists but points elsewhere, it is replaced.
// If link exists as a regular file or directory, an error is returned — the
// caller should migrate the file first via MoveToAgent.
func EnsureSymlink(link, target string) error {
	if err := EnsureDir(filepath.Dir(link)); err != nil {
		return err
	}

	existing, err := os.Lstat(link)
	if err == nil {
		// path exists
		if existing.Mode()&os.ModeSymlink != 0 {
			current, rerr := os.Readlink(link)
			if rerr == nil && current == target {
				return nil // already correct
			}
			// wrong target — replace
			if rerr := os.Remove(link); rerr != nil {
				return fmt.Errorf("removing stale symlink %s: %w", link, rerr)
			}
		} else {
			return fmt.Errorf("%s exists as a regular file or directory; migrate it first with 'ktrai restructure'", link)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("lstat %s: %w", link, err)
	}

	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("creating symlink %s -> %s: %w", link, target, err)
	}
	return nil
}

// IsSymlinkInto reports whether path is a symlink whose target is inside agentDir.
func IsSymlinkInto(path, agentDir string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}
	target, err := os.Readlink(path)
	if err != nil {
		return false, err
	}
	// Resolve relative symlink relative to the symlink's directory.
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	agentAbs, err := filepath.Abs(agentDir)
	if err != nil {
		return false, err
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(agentAbs, targetAbs)
	if err != nil {
		return false, err
	}
	// If rel starts with ".." it is outside agentDir.
	return len(rel) >= 2 && rel[:2] != "..", nil
}

// MoveToAgent moves src to destDir/basename(src) and creates a symlink
// at src pointing to the new location (relative path).
// If src is already a symlink into agentDir, it is left unchanged.
func MoveToAgent(src, destDir string) error {
	agentRoot := filepath.Dir(destDir)
	ok, err := IsSymlinkInto(src, agentRoot)
	if err != nil {
		return err
	}
	if ok {
		return nil // already managed
	}

	info, err := os.Lstat(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // nothing to move
		}
		return fmt.Errorf("lstat %s: %w", src, err)
	}

	dest := filepath.Join(destDir, filepath.Base(src))
	if err := EnsureDir(destDir); err != nil {
		return err
	}

	// If dest already exists (e.g. from a previous partial run), remove it first.
	if _, serr := os.Lstat(dest); serr == nil {
		if rerr := os.RemoveAll(dest); rerr != nil {
			return fmt.Errorf("clearing existing %s: %w", dest, rerr)
		}
	}

	if info.IsDir() {
		if err := copyDir(src, dest); err != nil {
			return fmt.Errorf("copying dir %s to %s: %w", src, dest, err)
		}
		if err := os.RemoveAll(src); err != nil {
			return fmt.Errorf("removing original %s: %w", src, err)
		}
	} else {
		if err := os.Rename(src, dest); err != nil {
			return fmt.Errorf("moving %s to %s: %w", src, dest, err)
		}
	}

	// Compute relative symlink target from src's directory to dest.
	srcDir := filepath.Dir(src)
	rel, err := filepath.Rel(srcDir, dest)
	if err != nil {
		return fmt.Errorf("computing relative path: %w", err)
	}
	return EnsureSymlink(src, rel)
}

// copyDir recursively copies src directory to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
