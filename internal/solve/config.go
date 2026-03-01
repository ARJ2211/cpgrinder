package solve

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const RunConfigName string = "run.json"

type RunConfig struct {
	Language string `json:"language"`
}

/*
Reads the config file in the directory
*/
func (r RunConfig) ReadRunConfig(dir string) (RunConfig, bool, error) {
	targetPath := filepath.Join(dir, RunConfigName)

	cfg := RunConfig{}
	data, err := os.ReadFile(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return RunConfig{}, false, nil
		}
		return RunConfig{}, true, err
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return RunConfig{}, true, err
	}

	if strings.TrimSpace(cfg.Language) == "" {
		return RunConfig{}, true, errors.New("language not set")
	}

	return cfg, true, nil
}

/*
Writes into a file the run config
*/
func (r RunConfig) WriteRunConfig(dir string) error {
	if strings.TrimSpace(dir) == "" {
		return errors.New("empty directory received")
	}
	if strings.TrimSpace(r.Language) == "" {
		return errors.New("language not set")
	}

	finalPath := filepath.Join(dir, RunConfigName)

	tmp, err := os.CreateTemp(dir, "run.json.*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		cleanup()
		return err
	}
	data = append(data, '\n')

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return err
	}

	if err := tmp.Sync(); err != nil {
		cleanup()
		return err
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
}
