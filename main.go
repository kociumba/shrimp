package main

import (
	"errors"
	"fmt"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Create   CreateCmd   `cmd:"" help:"create."`
	Remove   RemoveCmd   `cmd:"" help:"Remove a profile."`
	Activate ActivateCmd `cmd:"" help:"Activate a profile."`
	List     ListCmd     `cmd:"" help:"List all profiles."`
	File     FileCmd     `cmd:"" help:"Manage a profile"`
}

type CreateCmd struct {
	Name  string `arg:"" required:"" help:"Name of a profile to create."`
	Force bool   `short:"f" help:"Force creation of a profile."`
}

func (a *CreateCmd) Run() error {
	err := AddProfile(a.Name, a.Force)
	if err != nil {
		return err
	}
	fmt.Printf("Profile %q, created", a.Name)
	return nil
}

type RemoveCmd struct {
	Name  string `arg:"" required:"" help:"Name of a profile to remove."`
	Force bool   `short:"f" help:"Force delete a profile."`
}

func (r *RemoveCmd) Run() error {
	err := RemoveProfile(r.Name, r.Force)
	if err != nil {
		return err
	}
	fmt.Printf("Profile %q, deleted", r.Name)
	return nil
}

type ActivateCmd struct {
	Name  string `arg:"" required:"" help:"Name of a profile to activate."`
	Force bool   `short:"f" help:"Force activate a profile, overriting files USE CREAFULLY!."`
	Dry   bool   `short:"d" help:"Perform a dry-run of activating a profile to test if any issues arise"`
}

func (ac *ActivateCmd) Run() error {
	save := ""
	if ac.Dry {
		ReadConfig()
		save = c.Active
	}
	err := SetActiveProfile(ac.Name)
	if err != nil {
		return err
	}
	err = DeactivateAll(ac.Force, ac.Dry)
	if err != nil {
		return err
	}
	err = ActivateProfile(ac.Name, ac.Force, ac.Dry)
	if err != nil {
		return err
	}
	if ac.Dry {
		c.Active = save
		err := SaveConfig(false)
		if err != nil {
			return err
		}
	}

	if !ac.Dry {
		fmt.Printf("Activated profile: %q", ac.Name)
	}
	return nil
}

type ListCmd struct{}

func (l *ListCmd) Run() error {
	err := ReadConfig()
	if err != nil {
		return err
	}
	fmt.Println("Profiles:")
	for key := range c.Profiles {
		if key == c.Active {
			fmt.Printf("  - %q (active)\n", key)
		} else {
			fmt.Printf("  - %q\n", key)
		}
	}
	return nil
}

type FileCmd struct {
	Add    FileAddCmd    `cmd:"" help:"Add a file to the active profile."`
	Remove FileRemoveCmd `cmd:"" help:"Remove a file from the active profile."`
	List   FileListCmd   `cmd:"" help:"List files in the active profile."`
}

type FileAddCmd struct {
	Path string `arg:"" required:"" help:"Path to file to add." type:"path"`
}

func (f *FileAddCmd) Run() error {
	err := IsProfileActive()
	if err != nil {
		return err
	}
	return AddFileToActiveProfile(f.Path)
}

type FileRemoveCmd struct {
	Path string `arg:"" type:"path" required:"" help:"Path to file to remove."`
}

func (f *FileRemoveCmd) Run() error {
	err := IsProfileActive()
	if err != nil {
		return err
	}
	return RemoveFileFromActiveProfile(f.Path)
}

type FileListCmd struct{}

func (f *FileListCmd) Run() error {
	err := ReadConfig()
	if err != nil {
		return err
	}

	err = IsProfileActive()
	if err != nil {
		return err
	}

	if len(c.Profiles[c.Active].Files) <= 0 {
		fmt.Printf("profile %q is empty", c.Active)
		return nil
	}
	fmt.Printf("Paths in profile: %q\n", c.Active)
	for _, file := range c.Profiles[c.Active].Files {
		fmt.Printf("  - %q\n", file)
	}
	return nil
}

func IsProfileActive() error {
	ReadConfig()
	if c.Active == "" {
		return errors.New("no active profile, use \"shrimp activate <profile>\" to set the active profile")
	}
	return nil
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("shrimp"),
		kong.Description("Shrimp is a CLI tool to manage multiple configs with ease."),
		kong.ShortUsageOnError(),
	)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
