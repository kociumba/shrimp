package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Active   string
	Profiles map[string]Profile
}

type Profile struct {
	LastActive time.Time
	Files      []string
}

var (
	// global config
	c Config
)

func getConfigPath() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".config", "shrimp", "shrimp.toml"), nil
}

func (c *Config) DefaultPath() string {
	p, err := getConfigPath()
	if err != nil {
		fmt.Println(fmt.Errorf("failed to get default config path: %w", err))
	}
	return p
}

func getConfigDir(force bool) (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("could not get home dir")
	}
	cfg := filepath.Join(h, ".config", "shrimp")
	_, err = os.Stat(cfg)
	if err == nil {
		// if !force {
		// 	slog.Warn(fmt.Sprintf("%s already exists, please check if it's an old shrimp configuration or another tool, to create a new config there anyway use -f to force.", cfg))
		// 	os.Exit(69)
		// }

	} else if errors.Is(err, fs.ErrNotExist) {
		err = os.MkdirAll(cfg, os.ModePerm)
		if err != nil {
			return "", errors.New("could not create config dir")
		}
	} else {
		return "", errors.New("error getting info on the config directory")
	}

	// slog.Info(cfg)

	return cfg, nil
}

func SaveConfig(path string, force bool) error {
	_, err := getConfigDir(force)
	if err != nil {
		return err
	}

	if path == "" {
		path = c.DefaultPath()
	} else {
		if !strings.Contains(filepath.Base(path), ".") { // could be better but idc
			path = filepath.Join(path, "shrimp.toml")
		}
	}

	b, err := toml.Marshal(c)
	if err != nil {
		return errors.Join(errors.New("failed to marshal config"), err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		return errors.Join(errors.New("failed to write config to disk"), err)
	}
	return nil
}

func ReadConfig(path string) error {
	if path == "" {
		path = c.DefaultPath()
	} else {
		if !strings.Contains(filepath.Base(path), ".") { // could be better but idc
			path = filepath.Join(path, "shrimp.toml")
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Join(errors.New("failed to read config file"), err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return errors.Join(errors.New("failed to parse config"), err)
	}
	return nil
}

func AddProfile(name string, force bool, path string) error {
	ReadConfig(path) // read ignoring errors

	if c.Profiles == nil {
		c.Profiles = make(map[string]Profile)
	}

	if _, exists := c.Profiles[name]; exists {
		return fmt.Errorf("profile %q already exists", name)
	}

	p := Profile{
		LastActive: time.Time{},
	}
	if len(c.Profiles) == 0 {
		c.Active = name
		p.LastActive = time.Now()
	}
	c.Profiles[name] = p
	return SaveConfig(path, force)
}

func RemoveProfile(name string, force bool, path string) error {
	ReadConfig(path)

	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile %q, does not exist", name)
	}

	delete(c.Profiles, name)
	if c.Active == name {
		c.Active = ""
	}
	return SaveConfig(path, force)
}

func SetActiveProfile(name string, path string) error {
	ReadConfig(path)

	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile %q, does not exist", name)
	}

	c.Active = name

	p := c.Profiles[name]
	p.LastActive = time.Now()
	c.Profiles[name] = p

	return SaveConfig(path, false)
}

func AddFileToActiveProfile(path string, pathCFG string) error {
	ReadConfig(path)

	path, err := ExpandPath(path)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	profile := c.Profiles[c.Active]

	if slices.Contains(profile.Files, path) {
		return fmt.Errorf("path: %q, is already managed in profile: %q", path, c.Active)
	}

	profile.Files = append(profile.Files, path)
	c.Profiles[c.Active] = profile

	fmt.Printf("path: %q, added to profile: %q", path, c.Active)

	return SaveConfig(pathCFG, false)
}

func RemoveFileFromActiveProfile(path string, pathCFG string) error {
	ReadConfig(path)

	path, err := ExpandPath(path)
	if err != nil {
		return err
	}

	profile := c.Profiles[c.Active]
	removed := false
	filtered := []string{}
	for _, file := range profile.Files {
		if file == path {
			removed = true
			continue
		}
		filtered = append(filtered, file)
	}

	if !removed {
		return fmt.Errorf("path not found in profile: %s", path)
	}

	profile.Files = filtered
	c.Profiles[c.Active] = profile

	fmt.Printf("path: %q, removed from profile: %q", path, c.Active)

	return SaveConfig(pathCFG, false)
}
