package main

import (
	"fmt"
	"os"
)

var lockfiles = []string{
	"bun.lockb",
	"pnpm-lock.yaml",
	"package-lock.json",
}

func FileExists(filepath string) bool {
	if _, err := os.Stat(filepath); err == nil {
		return true
	}
	return false
}

func GetLockfile() (string, error) {
	cwd, err := os.Getwd()
	if err == nil {
		for _, lockfile := range lockfiles {
			filepath := cwd + "/" + lockfile
			if FileExists(filepath) {
				return filepath, nil
			}
		}
	}
	return "", fmt.Errorf(bold(red("[!]")), red("no lockfile found in %s"), blue(cwd))
}

func WriteNvmrc(nodeVersion string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	f, err := os.Create(cwd + "/.nvmrc")
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(nodeVersion)
	return err
}
