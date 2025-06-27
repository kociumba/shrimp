package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
)

// only call after setting c.Active
func ActivateProfile(name string, force bool, dry bool) error {
	ReadConfig()
	warnings := 0

	if _, ok := c.Profiles[name]; !ok {
		return fmt.Errorf("profile %q does not exist", name)
	}

	for _, f := range c.Profiles[c.Active].Files {
		disabled := f + "." + c.Active + ".disabled"

		if _, err := os.Stat(disabled); errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("a managed config file is missing: %q", disabled)
		}

		if _, err := os.Stat(f); err == nil {
			if !force {
				slog.Warn(fmt.Sprintf("target %q already exists, use -f to overwrite", f))
				warnings++
				continue
			}

			if !dry {
				if err := os.RemoveAll(f); err != nil {
					return fmt.Errorf("failed to remove existing %q: %w", f, err)
				}
			}
		}

		if !dry {
			err := os.Rename(disabled, f)
			if err != nil {
				return err
			}
		}
	}

	if !dry { // skip in dry runs, to emit all errors from both activate and deactivate
		if warnings != 0 {
			return fmt.Errorf("there were warnings emitted while activating %q", name)
		}
	}

	return nil
}

func DeactivateAll(force bool, dry bool) error {
	ReadConfig()
	warnings := 0

	for n, p := range c.Profiles {
		if n == c.Active {
			continue
		}

		for _, f := range p.Files {
			// f should always be cleaned by the config so it doesn't have trailing /
			// if it does then this logic breaks for directories
			disabled := f + "." + n + ".disabled"

			if _, err := os.Stat(f); errors.Is(err, fs.ErrNotExist) {
				if _, err := os.Stat(disabled); errors.Is(err, fs.ErrNotExist) {
					slog.Warn(fmt.Sprintf("path does not exist: %q, consider removing it from the profile", f))
					warnings++
				}
				continue
			}

			if _, err := os.Stat(disabled); err == nil && !force {
				slog.Warn(fmt.Sprintf("target %q already exists, use -f to overwrite", disabled))
				warnings++
				continue
			}

			if !dry {
				err := os.Rename(f, disabled)
				if err != nil {
					return err
				}
			}
		}
	}

	if !dry {
		if warnings != 0 {
			return errors.New("there were warnings emitted while deactivating all profiles")
		}
	}

	return nil
}
