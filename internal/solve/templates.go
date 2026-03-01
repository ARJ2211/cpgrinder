package solve

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed templates/**
var templatesFS embed.FS

/*
TemplateForLanguage returns (

	embedded template path,
	output file name,
	ok boolean

).
*/
func TemplateForLanguage(lang string) (templPath string, outName string, ok bool) {
	l := strings.ToLower(strings.TrimSpace(lang))
	if l == "" {
		return "", "", false
	}

	switch l {
	case "python3":
		// NOTE: embedded paths always use forward slashes, so path.Join is correct here
		return path.Join("templates", l, "main.py.tmpl"), "main.py", true
	default:
		return "", "", false
	}
}

/*
EnsureStarterFile makes sure the starter solution file exists inside dir.
dir = <workspaceDir>/<source>/<sourceID>
*/
func EnsureStarterFile(dir string, lang string) (solutionPath string, created bool, err error) {
	l := strings.ToLower(strings.TrimSpace(lang))
	if l == "" {
		return "", false, errors.New("no language provided")
	}
	if strings.TrimSpace(dir) == "" {
		return "", false, errors.New("no dir provided")
	}

	templPath, outName, ok := TemplateForLanguage(l)
	if !ok {
		return "", false, fmt.Errorf("no template found for %s", l)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", false, err
	}

	solutionPath = filepath.Join(dir, outName)

	if _, err := os.Stat(solutionPath); err == nil {
		return solutionPath, false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, err
	}

	templateBytes, err := fs.ReadFile(templatesFS, templPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read embedded template %s: %w", templPath, err)
	}

	tmpPath := solutionPath + ".tmp"
	if err := os.WriteFile(tmpPath, templateBytes, 0o644); err != nil {
		return "", false, err
	}
	if err := os.Rename(tmpPath, solutionPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", false, err
	}

	return solutionPath, true, nil
}
