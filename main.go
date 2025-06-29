package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/kong"
)

type Globals struct {
	Config string `type:"path" short:"c" help:"a path to use for a non global config file"`
}

type CLI struct {
	Globals

	Create   CreateCmd   `cmd:"" aliases:"c" help:"Create a profile"`
	Remove   RemoveCmd   `cmd:"" aliases:"r" help:"Remove a profile"`
	Activate ActivateCmd `cmd:"" aliases:"a" help:"Activate a profile"`
	Reaload  ReloadCmd   `cmd:"" aliases:"re" help:"reload the current active profile"`
	Clone    CloneCmd    `cmd:"" aliases:"cl" help:"clone a profile, to save time when creating new ones"`
	List     ListCmd     `cmd:"" aliases:"l" help:"List all profiles"`
	File     FileCmd     `cmd:"" aliases:"f" help:"Manage a profile"`
	Hook     HookCmd     `cmd:"" aliases:"h" help:"allows for editing pre and post profile activation hooks"`
}

type CreateCmd struct {
	Name  string `arg:"" required:"" help:"Name of a profile to create"`
	Force bool   `short:"f" help:"Force creation of a profile"`
}

func (a *CreateCmd) Run(globals *Globals) error {
	err := AddProfile(a.Name, a.Force, globals.Config)
	if err != nil {
		return err
	}
	fmt.Printf("Profile %q, created", a.Name)
	return nil
}

type RemoveCmd struct {
	Name  string `arg:"" required:"" help:"Name of a profile to remove"`
	Force bool   `short:"f" help:"Force delete a profile"`
}

func (r *RemoveCmd) Run(globals *Globals) error {
	err := RemoveProfile(r.Name, r.Force, globals.Config)
	if err != nil {
		return err
	}
	fmt.Printf("Profile %q, deleted", r.Name)
	return nil
}

type ActivateCmd struct {
	Name  string `arg:"" required:"" help:"Name of a profile to activate"`
	Force bool   `short:"f" help:"Force activate a profile, overriting files USE CREAFULLY!"`
	Dry   bool   `short:"d" help:"Perform a dry-run of activating a profile to test if any issues arise"`
}

func (ac *ActivateCmd) Run(globals *Globals) error {
	s := time.Now()
	ReadConfig(globals.Config)
	if ac.Name == c.Active {
		return fmt.Errorf("profile: %q is already active", ac.Name)
	}

	save := c.Active
	err := SetActiveProfile(ac.Name, globals.Config)
	if err != nil {
		return err
	}

	err = ExecutePre(globals.Config, ac.Dry)
	if err != nil {
		if err := SetActiveProfile(save, globals.Config); err != nil {
			return err
		}
		return err
	}

	err = SwitchToProfile(ac.Name, ac.Force, ac.Dry)
	if err != nil {
		if err := SetActiveProfile(save, globals.Config); err != nil {
			return err
		}
		return err
	}

	err = ExecutePost(globals.Config, ac.Dry)
	if err != nil {
		if err := SetActiveProfile(save, globals.Config); err != nil {
			return err
		}
		return err
	}

	if !ac.Dry {
		fmt.Printf("Successfully switched to profile %q, in %v\n", ac.Name, time.Since(s))
	} else {
		fmt.Printf("Dry run: would switch to profile %q with no issues detected\n", ac.Name)
	}

	return nil
}

type ReloadCmd struct {
	Dry bool `short:"d" help:"Perform a dry-run of activating a profile to test if any issues arise"`
}

func (rc *ReloadCmd) Run(globals *Globals) error {
	ReadConfig(globals.Config)
	save := c.Active

	err := ExecutePre(globals.Config, rc.Dry)
	if err != nil {
		if err := SetActiveProfile(save, globals.Config); err != nil {
			return err
		}
		return err
	}

	err = ExecutePost(globals.Config, rc.Dry)
	if err != nil {
		if err := SetActiveProfile(save, globals.Config); err != nil {
			return err
		}
		return err
	}

	return nil
}

type CloneCmd struct {
	NewName string `arg:"" required:"" help:"name of the clone profile"`
	Name    string `arg:"" optional:"" help:"name of the profile that will be cloned, if left empty clones the active profile"`
}

func (c *CloneCmd) Run(globals *Globals) error {
	err := CloneProfile(c.NewName, c.Name, globals.Config)
	if err != nil {
		return err
	}

	return nil
}

type ListCmd struct{}

func (l *ListCmd) Run(globals *Globals) error {
	err := ReadConfig(globals.Config)
	if err != nil {
		return err
	}
	fmt.Println("Profiles:")
	for key, p := range c.Profiles {
		if key == c.Active {
			fmt.Printf("  - %q (active)\n", key)
		} else {
			fmt.Printf("  - %q (last active: %s)\n", key, p.LastActive.Local().Format(time.DateTime))
		}
	}
	return nil
}

type FileCmd struct {
	Add    FileAddCmd    `cmd:"" aliases:"a" help:"Add a file to the active profile"`
	Remove FileRemoveCmd `cmd:"" aliases:"r" help:"Remove a file from the active profile"`
	List   FileListCmd   `cmd:"" aliases:"l" help:"List files in the active profile"`
}

type FileAddCmd struct {
	Path string `arg:"" required:"" help:"Path to file to add" type:"path"`
}

func (f *FileAddCmd) Run(globals *Globals) error {
	err := IsProfileActive(globals.Config)
	if err != nil {
		return err
	}
	return AddFileToActiveProfile(f.Path, globals.Config)
}

type FileRemoveCmd struct {
	Path string `arg:"" type:"path" required:"" help:"Path to file to remove"`
}

func (f *FileRemoveCmd) Run(globals *Globals) error {
	err := IsProfileActive(globals.Config)
	if err != nil {
		return err
	}
	return RemoveFileFromActiveProfile(f.Path, globals.Config)
}

type FileListCmd struct{}

func (f *FileListCmd) Run(globals *Globals) error {
	err := ReadConfig(globals.Config)
	if err != nil {
		return err
	}

	err = IsProfileActive(globals.Config)
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

type HookCmd struct {
	Pre  PreCmd  `cmd:"" help:"edit pre profile activation hooks"`
	Post PostCmd `cmd:"" help:"edit post profile activation hooks"`
}

type PreCmd struct {
	Cmd []string `arg:"" help:"commands to execute" default:""`
}

func (p *PreCmd) Run(globals *Globals) error {
	err := SetPreHook(globals.Config, p.Cmd)
	if err != nil {
		return err
	}

	return nil
}

type PostCmd struct {
	Cmd []string `arg:"" help:"commands to execute" default:""`
}

func (p *PostCmd) Run(globals *Globals) error {
	err := SetPostHook(globals.Config, p.Cmd)
	if err != nil {
		return err
	}

	return nil
}

func IsProfileActive(path string) error {
	ReadConfig(path)
	if c.Active == "" {
		return errors.New("no active profile, use \"shrimp activate <profile>\" to set the active profile")
	}
	return nil
}

var cli CLI

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("shrimp"),
		kong.Description("🦐 shrimp is a CLI tool to manage multiple configs with ease."),
		kong.UsageOnError(),
	)

	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}
