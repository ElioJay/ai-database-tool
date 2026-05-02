package config

import (
	"os"
	"path/filepath"
	"runtime"
)

type Mode int

const (
	ModeInstalled Mode = iota
	ModePortable
)

type ConfigDir struct {
	Path string
	Mode Mode
}

func Resolve() ConfigDir {
	exePath, err := os.Executable()
	if err != nil {
		return installedDir()
	}
	return resolveFromDir(filepath.Dir(exePath))
}

func resolveFromDir(exeDir string) ConfigDir {
	portablePath := filepath.Join(exeDir, ".aidbt")
	if info, err := os.Stat(portablePath); err == nil && info.IsDir() {
		return ConfigDir{Path: portablePath, Mode: ModePortable}
	}
	return installedDir()
}

func installedDir() ConfigDir {
	return ConfigDir{Path: installedPath(), Mode: ModeInstalled}
}

func installedPath() string {
	if runtime.GOOS == "windows" {
		if p := os.Getenv("USERPROFILE"); p != "" {
			return filepath.Join(p, ".aidbt")
		}
	}
	if h := os.Getenv("HOME"); h != "" {
		return filepath.Join(h, ".aidbt")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aidbt")
}

func (m Mode) String() string {
	if m == ModePortable {
		return "portable"
	}
	return "installed"
}
