package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	configDirNix = ".rum-goggles"
	configDirWin = "RumGoggles"

	imageDir = "images"

	logFile = "rumgoggles.log"
	sqlFile = "rumgoggles.db"
)

func Database() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", pkgErr("error getting config directory", err)
	}

	path := filepath.Join(dir, sqlFile)

	f, err := os.OpenFile(path, os.O_CREATE, 0644)
	if err != nil {
		return "", pkgErr("error opening database file", err)
	}
	defer f.Close()

	return path, nil
}

func ImageDir() (string, error) {
	cfgDir, err := configDir()
	if err != nil {
		return "", pkgErr("error getting config directory", err)
	}

	dir := filepath.Join(cfgDir, imageDir)

	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return "", fmt.Errorf("error making directory: %v", err)
	}

	return dir, nil
}

// TODO: implement log rotation
// Rotate log file every week?
// Keep most recent 4 logs?
func Log() (*os.File, error) {
	dir, err := configDir()
	if err != nil {
		return nil, pkgErr("error getting config directory", err)
	}

	path := filepath.Join(dir, logFile)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, pkgErr("error opening log file", err)
	}

	return f, nil
}

func configDir() (string, error) {
	var dir string
	var err error

	switch runtime.GOOS {
	case "windows":
		dir, err = userDirWin()
	default:
		dir, err = userDir()
	}
	if err != nil {
		return "", fmt.Errorf("error getting user directory: %v", err)
	}

	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return "", fmt.Errorf("error making directory: %v", err)
	}

	return dir, nil
}

func userDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}

	return filepath.Join(dir, configDirNix), nil
}

func userDirWin() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("error getting cache directory: %v", err)
	}

	return filepath.Join(dir, configDirWin), nil
}
