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

type MissingBinaryError struct {
	Name string
}

func (e MissingBinaryError) Error() string {
	return fmt.Sprintf("missing required binary '%s'", e.Name)
}

type RunnerDeps struct {
	LookPath func(string) (string, error)
}

type Runner struct {
	deps RunnerDeps
}

func NewRunner(deps RunnerDeps) Runner {
	if deps.LookPath == nil {
		deps.LookPath = exec.LookPath
	}
	return Runner{deps: deps}
}

func DefaultRunner() Runner {
	return NewRunner(RunnerDeps{})
}

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

func (r Runner) RequireBinary(name string) error {
	_, err := r.deps.LookPath(name)
	if err != nil {
		return MissingBinaryError{Name: name}
	}
	return nil
}

// Backward-compatible wrapper.
func RequireBinary(name string) error {
	return DefaultRunner().RequireBinary(name)
}

func (r Runner) requireIfOnPath(prog string) error {
	if strings.ContainsAny(prog, `/\\`) || strings.HasPrefix(prog, ".") {
		return nil
	}
	return r.RequireBinary(prog)
}

func (r Runner) compileIfNeeded(ctx context.Context, spec LanguageSpec, dir string) (RunResult, bool, error) {
	spec = applyDefaults(spec)
	if !spec.IsCompiled {
		return RunResult{}, true, nil
	}

	if err := os.MkdirAll(filepath.Join(dir, spec.BuildDir), 0o755); err != nil {
		return RunResult{}, false, err
	}
	if len(spec.CompileArgv) == 0 {
		return RunResult{}, false, fmt.Errorf("compile argv missing for %s", spec.ID)
	}
	if err := r.requireIfOnPath(spec.CompileArgv[0]); err != nil {
		return RunResult{}, false, err
	}

	compileRes, err := RunCommand(ctx, RunOptions{
		Dir:     dir,
		Argv:    spec.CompileArgv,
		Stdin:   nil,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return compileRes, false, err
	}

	if compileRes.TimedOut || compileRes.ExitCode != 0 {
		if strings.TrimSpace(compileRes.Stderr) != "" {
			compileRes.Stderr = "compile failed:\n" + compileRes.Stderr
		} else {
			compileRes.Stderr = "compile failed"
		}
		return compileRes, false, nil
	}

	return compileRes, true, nil
}

func (r Runner) runOnly(ctx context.Context, spec LanguageSpec, dir string, stdin []byte, timeoutOverride time.Duration) (RunResult, error) {
	spec = applyDefaults(spec)

	if len(spec.RunCmd) == 0 {
		return RunResult{}, fmt.Errorf("run argv missing for %s", spec.ID)
	}
	if err := r.requireIfOnPath(spec.RunCmd[0]); err != nil {
		return RunResult{}, err
	}

	runTimeout := spec.DefaultTimeout
	if timeoutOverride > 0 {
		runTimeout = timeoutOverride
	}

	return RunCommand(ctx, RunOptions{
		Dir:     dir,
		Argv:    spec.RunCmd,
		Stdin:   stdin,
		Timeout: runTimeout,
	})
}

/*
Runs any registered language spec.
If spec.IsCompiled, it compiles every time for now (no caching yet).
*/
func (r Runner) RunWithSpec(ctx context.Context, spec LanguageSpec, dir string, stdin []byte, timeoutOverride time.Duration) (RunResult, error) {
	compileRes, ok, err := r.compileIfNeeded(ctx, spec, dir)
	if err != nil {
		return RunResult{}, err
	}
	if !ok {
		return compileRes, nil
	}
	return r.runOnly(ctx, spec, dir, stdin, timeoutOverride)
}

// Backward-compatible wrapper.
func RunWithSpec(ctx context.Context, spec LanguageSpec, dir string, stdin []byte, timeoutOverride time.Duration) (RunResult, error) {
	return DefaultRunner().RunWithSpec(ctx, spec, dir, stdin, timeoutOverride)
}

// ====================== WRAPPERS =========================

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
