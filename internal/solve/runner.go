package solve

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	Argv    []string
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

/*
Runs any registered language spec.
If spec.IsCompiled, it compiles every time for now (no caching yet).
*/
func RunWithSpec(ctx context.Context, spec LanguageSpec, dir string, stdin []byte, timeoutOverride time.Duration) (RunResult, error) {
	spec = applyDefaults(spec)

	runTimeout := spec.DefaultTimeout
	if timeoutOverride > 0 {
		runTimeout = timeoutOverride
	}

	if spec.IsCompiled {
		if err := os.MkdirAll(filepath.Join(dir, spec.BuildDir), 0o755); err != nil {
			return RunResult{}, err
		}
		if len(spec.CompileArgv) == 0 {
			return RunResult{}, fmt.Errorf("compile argv missing for %s", spec.ID)
		}

		if err := requireIfOnPath(spec.CompileArgv[0]); err != nil {
			return RunResult{}, err
		}

		compileRes, err := RunCommand(ctx, RunOptions{
			Dir:     dir,
			Argv:    spec.CompileArgv,
			Stdin:   nil,
			Timeout: 10 * time.Second,
		})
		if err != nil {
			return compileRes, err
		}
		if compileRes.TimedOut || compileRes.ExitCode != 0 {
			// Treat as "ran but failed" so UI can show compiler stderr.
			// (You can label this as CE later in compare.go)
			if strings.TrimSpace(compileRes.Stderr) != "" {
				compileRes.Stderr = "compile failed:\n" + compileRes.Stderr
			} else {
				compileRes.Stderr = "compile failed"
			}
			return compileRes, nil
		}
	}

	if len(spec.RunCmd) == 0 {
		return RunResult{}, fmt.Errorf("run argv missing for %s", spec.ID)
	}
	if err := requireIfOnPath(spec.RunCmd[0]); err != nil {
		return RunResult{}, err
	}

	return RunCommand(ctx, RunOptions{
		Dir:     dir,
		Argv:    spec.RunCmd,
		Stdin:   stdin,
		Timeout: runTimeout,
	})
}

func requireIfOnPath(prog string) error {
	// If it looks like a path (contains a slash or starts with "."), don't LookPath.
	// Examples: "./.build/main", ".build/main"
	if strings.ContainsAny(prog, `/\`) || strings.HasPrefix(prog, ".") {
		return nil
	}
	return RequireBinary(prog)
}

// Convenience wrappers

func RunPython3(ctx context.Context, dir string, stdin []byte, timeout time.Duration) (RunResult, error) {
	spec, ok := GetLanguageSpec(LangPython3)
	if !ok {
		return RunResult{}, fmt.Errorf("unknown language: %s", LangPython3)
	}
	return RunWithSpec(ctx, spec, dir, stdin, timeout)
}

func RunJavaScript(ctx context.Context, dir string, stdin []byte, timeout time.Duration) (RunResult, error) {
	spec, ok := GetLanguageSpec(LangJavaScript)
	if !ok {
		return RunResult{}, fmt.Errorf("unknown language: %s", LangJavaScript)
	}
	return RunWithSpec(ctx, spec, dir, stdin, timeout)
}

func RunCPP(ctx context.Context, dir string, stdin []byte, timeout time.Duration) (RunResult, error) {
	spec, ok := GetLanguageSpec(LangCPP)
	if !ok {
		return RunResult{}, fmt.Errorf("unknown language: %s", LangCPP)
	}
	return RunWithSpec(ctx, spec, dir, stdin, timeout)
}

func RunGo(ctx context.Context, dir string, stdin []byte, timeout time.Duration) (RunResult, error) {
	spec, ok := GetLanguageSpec(LangGo)
	if !ok {
		return RunResult{}, fmt.Errorf("unknown language: %s", LangGo)
	}
	return RunWithSpec(ctx, spec, dir, stdin, timeout)
}

func RunJava(ctx context.Context, dir string, stdin []byte, timeout time.Duration) (RunResult, error) {
	spec, ok := GetLanguageSpec(LangJava)
	if !ok {
		return RunResult{}, fmt.Errorf("unknown language: %s", LangJava)
	}
	return RunWithSpec(ctx, spec, dir, stdin, timeout)
}
