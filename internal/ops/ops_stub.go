//go:build !linux

package ops

import "errors"

func InstallBaseTools() (string, error) {
	return "", errors.New("ops are only supported on linux")
}

func UninstallBaseTools() (string, error) {
	return "", errors.New("ops are only supported on linux")
}

func BaseTools() []string {
	return []string{}
}
