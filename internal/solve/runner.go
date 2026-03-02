// internal/solve/runner.go
package solve

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type RunResult struct {
	Stdout     string
	Stderr     string
	Stdin      string
	ExitCode   int
	DurationMS time.Duration
	TimedOut   bool
	Cmd        []string
}

type RunOptions struct {
	Dir     string
	Argv    []string // e.g. ["python3", "main.py"]
	Stdin   []byte
	Timeout time.Duration
	Env     map[string]string // optional overrides
}

func RunCommand(ctx context.Context, opts RunOptions) (RunResult, error) {
	var res RunResult

	if opts.Dir == "" {
		return res, errors.New("run: Dir is required")
	}
	if len(opts.Argv) == 0 {
		return res, errors.New("run: Argv is required")
	}
	if opts.Timeout <= 0 {
		return res, errors.New("run: Timeout must be > 0")
	}

	res.Cmd = append([]string(nil), opts.Argv...)
	res.Stdin = string(opts.Stdin)

	ctxRun, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctxRun, opts.Argv[0], opts.Argv[1:]...)
	cmd.Dir = opts.Dir
	cmd.Env = mergedEnv(opts.Env)
	cmd.Stdin = bytes.NewReader(opts.Stdin)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	start := time.Now()
	err := cmd.Run()
	res.DurationMS = time.Since(start)

	res.Stdout = outBuf.String()
	res.Stderr = errBuf.String()

	if ctxRun.Err() == context.DeadlineExceeded {
		res.TimedOut = true
		res.ExitCode = -1
		return res, nil
	}

	if err == nil {
		res.ExitCode = 0
		return res, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		res.ExitCode = exitErr.ExitCode()
		return res, nil
	}

	// If we get here, the process likely didn't start (bad binary, permission, etc.)
	res.ExitCode = -1
	return res, err
}

func mergedEnv(overrides map[string]string) []string {
	if len(overrides) == 0 {
		return os.Environ()
	}

	envMap := make(map[string]string, 64)
	for _, kv := range os.Environ() {
		k, v, ok := splitEnv(kv)
		if ok {
			envMap[k] = v
		}
	}
	for k, v := range overrides {
		envMap[k] = v
	}

	out := make([]string, 0, len(envMap))
	for k, v := range envMap {
		out = append(out, k+"="+v)
	}
	return out
}

func splitEnv(kv string) (key, val string, ok bool) {
	for i := 0; i < len(kv); i++ {
		if kv[i] == '=' {
			return kv[:i], kv[i+1:], true
		}
	}
	return "", "", false
}

func RequireBinary(name string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("missing required binary %q in PATH", name)
	}
	return nil
}

func RunPython3(ctx context.Context, dir string, stdin []byte, timeout time.Duration) (RunResult, error) {
	if err := RequireBinary("python3"); err != nil {
		return RunResult{}, err
	}

	return RunCommand(ctx, RunOptions{
		Dir:     dir,
		Argv:    []string{"python3", "main.py"},
		Stdin:   stdin,
		Timeout: timeout,
	})
}
