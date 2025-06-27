package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
)

type ValidationIssue struct {
	Type       string // "error", "warning"
	Operation  string // "activate", "deactivate"
	Profile    string
	File       string
	Message    string
	Suggestion string
}

type ValidationResult struct {
	Issues      []ValidationIssue
	HasErrors   bool
	HasWarnings bool
}

func (vr *ValidationResult) AddError(operation, profile, file, message, suggestion string) {
	vr.Issues = append(vr.Issues, ValidationIssue{
		Type:       "error",
		Operation:  operation,
		Profile:    profile,
		File:       file,
		Message:    message,
		Suggestion: suggestion,
	})
	vr.HasErrors = true
}

func (vr *ValidationResult) AddWarning(operation, profile, file, message, suggestion string) {
	vr.Issues = append(vr.Issues, ValidationIssue{
		Type:       "warning",
		Operation:  operation,
		Profile:    profile,
		File:       file,
		Message:    message,
		Suggestion: suggestion,
	})
	vr.HasWarnings = true
}

func (vr *ValidationResult) PrintIssues() {
	for _, issue := range vr.Issues {
		prefix := fmt.Sprintf("[%s:%s:%s]", issue.Type, issue.Operation, issue.Profile)
		if issue.File != "" {
			fmt.Printf("%s %s (file: %q)\n", prefix, issue.Message, issue.File)
		} else {
			fmt.Printf("%s %s\n", prefix, issue.Message)
		}
		if issue.Suggestion != "" {
			fmt.Printf("    -> Suggestion: %s\n", issue.Suggestion)
		}
	}
}

func isFileManaged(filePath string, excludeProfile string) (bool, string) {
	for profileName, profile := range c.Profiles {
		if profileName == excludeProfile {
			continue
		}
		if slices.Contains(profile.Files, filePath) {
			return true, profileName
		}
	}
	return false, ""
}

func ValidateProfileActivation(name string, force bool) *ValidationResult {
	ReadConfig(cli.Globals.Config)
	result := &ValidationResult{}

	if _, ok := c.Profiles[name]; !ok {
		result.AddError("activate", name, "",
			fmt.Sprintf("profile %q does not exist", name),
			"check available profiles with \"list\"")
		return result
	}

	for _, f := range c.Profiles[name].Files {
		disabled := f + "." + name + ".disabled"

		if _, err := os.Stat(disabled); errors.Is(err, fs.ErrNotExist) {
			result.AddError("activate", name, f,
				fmt.Sprintf("managed config file is missing: %q", disabled),
				"check if any the file still exists or was moved")
			continue
		}

		if _, err := os.Stat(f); err == nil {
			isManaged, _ := isFileManaged(f, name)

			if isManaged {
				continue
			} else if !force {
				result.AddWarning("activate", name, f,
					fmt.Sprintf("target %q already exists and is not managed by any other profile", f),
					"use -f flag to overwrite unmanaged file")
			}
		}
	}

	return result
}

func ValidateProfileDeactivation(force bool) *ValidationResult {
	ReadConfig(cli.Globals.Config)
	result := &ValidationResult{}

	for n, p := range c.Profiles {
		if n == c.Active {
			continue
		}

		for _, f := range p.Files {
			disabled := f + "." + n + ".disabled"

			if _, err := os.Stat(f); errors.Is(err, fs.ErrNotExist) {
				if _, err := os.Stat(disabled); errors.Is(err, fs.ErrNotExist) {
					result.AddWarning("deactivate", n, f,
						fmt.Sprintf("path does not exist: %q", f),
						"consider removing it from the profile configuration")
				}
				continue
			}

			if _, err := os.Stat(disabled); err == nil {
				continue
			}

			isManaged, managingProfile := isFileManaged(f, n)

			if isManaged && managingProfile == c.Active {
				continue
			}

			// this condition is just unnescessary why the fuck did it end up here
			// if !isManaged && !force {
			// 	result.AddWarning("deactivate", n, f,
			// 		fmt.Sprintf("file %q is not managed by any other profile, deactivating may cause loss", f),
			// 		"use -f flag to force deactivation")
			// }
		}
	}

	return result
}

func ValidateProfileSwitch(targetProfile string, force bool) *ValidationResult {
	deactivateResult := ValidateProfileDeactivation(force)
	activateResult := ValidateProfileActivation(targetProfile, force)

	combined := &ValidationResult{
		HasErrors:   deactivateResult.HasErrors || activateResult.HasErrors,
		HasWarnings: deactivateResult.HasWarnings || activateResult.HasWarnings,
	}
	combined.Issues = append(combined.Issues, deactivateResult.Issues...)
	combined.Issues = append(combined.Issues, activateResult.Issues...)

	return combined
}

func ActivateProfile(name string, force bool, dry bool) error {
	ReadConfig(cli.Globals.Config)

	if _, ok := c.Profiles[name]; !ok {
		return fmt.Errorf("profile %q does not exist", name)
	}

	for _, f := range c.Profiles[name].Files {
		disabled := f + "." + name + ".disabled"

		if _, err := os.Stat(f); err == nil && force {
			if !dry {
				if err := os.RemoveAll(f); err != nil {
					return fmt.Errorf("failed to remove existing %q: %w", f, err)
				}
			}
		}

		if !dry {
			err := os.Rename(disabled, f)
			if err != nil {
				return fmt.Errorf("failed to activate %q: %w", f, err)
			}
		}
	}

	return nil
}

func DeactivateAll(force bool, dry bool) error {
	ReadConfig(cli.Globals.Config)

	for n, p := range c.Profiles {
		if n == c.Active {
			continue
		}

		for _, f := range p.Files {
			disabled := f + "." + n + ".disabled"

			if _, err := os.Stat(f); errors.Is(err, fs.ErrNotExist) {
				continue
			}

			if _, err := os.Stat(disabled); err == nil {
				continue
			}

			isManaged, managingProfile := isFileManaged(f, n)
			if isManaged && managingProfile == n {
				continue
			}

			if !dry {
				err := os.Rename(f, disabled)
				if err != nil {
					return fmt.Errorf("failed to deactivate %q: %w", f, err)
				}
			}
		}
	}

	return nil
}

func SwitchToProfile(targetProfile string, force bool, dry bool) error {
	validation := ValidateProfileSwitch(targetProfile, force)

	if len(validation.Issues) > 0 {
		fmt.Printf("Validation found %d issues for profile switch to %q:\n",
			len(validation.Issues), targetProfile)
		validation.PrintIssues()
	}

	if validation.HasErrors {
		return errors.New("cannot proceed with profile switch due to validation errors")
	}

	if validation.HasWarnings && !force {
		return errors.New("cannot proceed with profile switch due to warnings (use -f to force)")
	}

	if err := DeactivateAll(force, dry); err != nil {
		return fmt.Errorf("failed to deactivate profiles: %w", err)
	}

	if err := ActivateProfile(targetProfile, force, dry); err != nil {
		return fmt.Errorf("failed to activate profile %q: %w", targetProfile, err)
	}

	if !dry {
		fmt.Printf("Successfully switched to profile %q", targetProfile)
	} else {
		fmt.Printf("Dry run: would switch to profile %q with no issues detected", targetProfile)
	}

	return nil
}
