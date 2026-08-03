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

// WriteIfNotExists writes content to path only when the path does not already
// exist. Parent directories are created as needed.
func WriteIfNotExists(path, content string) error {
	if _, err := os.Lstat(path); err == nil {
		return nil // already exists — leave it untouched
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("lstat %s: %w", path, err)
	}
	return WriteFile(path, content)
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
			return fmt.Errorf("%s exists as a regular file or directory; run 'ktrai align' to migrate it automatically", link)
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

// MigrateSkillsDir migrates each top-level skill subdirectory from srcDir into
// agentSkillsDir, using MoveToAgent for each entry. It is a no-op when srcDir
// does not exist or is already a symlink (the caller's wireToolSymlinks will
// re-point it). It returns the list of skill names that were actually moved so
// the caller can detect collisions before invoking this function.
//
// Entries inside srcDir that are symlinks whose resolved target is already
// inside agentSkillsDir are skipped — they are already managed.
func MigrateSkillsDir(srcDir, agentSkillsDir string) ([]string, error) {
	info, err := os.Lstat(srcDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("lstat %s: %w", srcDir, err)
	}
	// Already a symlink — wireToolSymlinks will update it; nothing to migrate.
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, nil
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s exists but is not a directory", srcDir)
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", srcDir, err)
	}

	agentAbs, err := filepath.Abs(agentSkillsDir)
	if err != nil {
		return nil, err
	}

	var migrated []string
	for _, e := range entries {
		entryPath := filepath.Join(srcDir, e.Name())

		// If it's a symlink that already resolves into agentSkillsDir, skip.
		if e.Type()&os.ModeSymlink != 0 {
			target, rerr := os.Readlink(entryPath)
			if rerr == nil {
				if !filepath.IsAbs(target) {
					target = filepath.Join(filepath.Dir(entryPath), target)
				}
				targetAbs, aerr := filepath.Abs(target)
				if aerr == nil {
					rel, rerr2 := filepath.Rel(agentAbs, targetAbs)
					if rerr2 == nil && (rel == "." || (len(rel) >= 2 && rel[:2] != "..")) {
						continue // already points into .agent/skills/
					}
				}
			}
			// Symlink to somewhere else — skip; we cannot safely move it.
			continue
		}

		if err := MoveToAgent(entryPath, agentSkillsDir); err != nil {
			return migrated, fmt.Errorf("migrating skill %s: %w", e.Name(), err)
		}
		migrated = append(migrated, e.Name())
	}
	return migrated, nil
}

// CollapseSkillSymlinks removes individual per-skill symlinks from srcDir and
// removes srcDir itself so that EnsureSymlink can replace the whole directory
// with a single directory-level symlink pointing to agentSkillsDir.
//
// A symlink entry is safe to remove when either:
//   - it already points into agentSkillsDir (created by a previous MoveToAgent
//     call), or
//   - the skill name exists in agentSkillsDir (the skill was migrated there
//     from another source, making this symlink stale).
//
// If any entry is a real file or directory, or a symlink whose skill is not
// present in agentSkillsDir, the function returns without removing anything.
func CollapseSkillSymlinks(srcDir, agentSkillsDir string) error {
	info, err := os.Lstat(srcDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("lstat %s: %w", srcDir, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil // already a symlink or not a directory — nothing to do
	}

	agentAbs, err := filepath.Abs(agentSkillsDir)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("reading %s: %w", srcDir, err)
	}

	// Verify every entry can be safely removed before touching anything.
	for _, e := range entries {
		if e.Type()&os.ModeSymlink == 0 {
			return nil // real file or directory present — cannot collapse safely
		}
		target, rerr := os.Readlink(filepath.Join(srcDir, e.Name()))
		if rerr != nil {
			return nil // unreadable symlink — leave as is
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(srcDir, target)
		}
		targetAbs, aerr := filepath.Abs(target)
		if aerr != nil {
			return nil
		}
		rel, rerr := filepath.Rel(agentAbs, targetAbs)
		alreadyInAgent := rerr == nil && rel != "" && (len(rel) < 2 || rel[:2] != "..")
		if alreadyInAgent {
			continue // symlink already points into agentSkillsDir — safe to remove
		}
		// Symlink points elsewhere. Safe to remove only if the skill has been
		// migrated into agentSkillsDir (making this symlink stale).
		if _, serr := os.Lstat(filepath.Join(agentSkillsDir, e.Name())); serr != nil {
			return nil // skill not in agentSkillsDir — cannot collapse safely
		}
	}

	// All entries are removable — remove symlinks then the directory.
	for _, e := range entries {
		if rerr := os.Remove(filepath.Join(srcDir, e.Name())); rerr != nil {
			return fmt.Errorf("removing symlink %s: %w", filepath.Join(srcDir, e.Name()), rerr)
		}
	}
	return os.Remove(srcDir)
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
