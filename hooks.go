package main

import (
	"os"
	"os/exec"
)

func ExecutePre(path string) error {
	ReadConfig(path)

	var p = c.Profiles[c.Active]
	if len(p.PreActivate) > 0 {
		cmd := exec.Command(p.PreActivate[0], p.PreActivate[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func ExecutePost(path string) error {
	ReadConfig(path)

	var p = c.Profiles[c.Active]
	if len(p.PostActivate) > 0 {
		cmd := exec.Command(p.PostActivate[0], p.PostActivate[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
