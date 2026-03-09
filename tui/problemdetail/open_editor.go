package problemdetail

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ARJ2211/cpgrinder/internal/solve"
	"github.com/ARJ2211/cpgrinder/internal/store"
)

type editorDoneMsg struct {
	err error
}

func openEditorCmd(db *store.Store, p store.ProblemID) tea.Cmd {
	if db == nil {
		return func() tea.Msg { return editorDoneMsg{err: errors.New("missing db store")} }
	}
	if strings.TrimSpace(p.Id) == "" {
		return func() tea.Msg { return editorDoneMsg{err: errors.New("no problem loaded")} }
	}

	wsRoot := db.WorkspacePath()

	dir, err := solve.EnsureProblemDir(wsRoot, p)
	if err != nil {
		return func() tea.Msg { return editorDoneMsg{err: err} }
	}

	cfg, exists, err := (solve.RunConfig{}).ReadRunConfig(dir)
	if err != nil {
		return func() tea.Msg { return editorDoneMsg{err: err} }
	}
	if !exists {
		cfg = solve.RunConfig{Language: "python3"}
		if err := cfg.WriteRunConfig(dir); err != nil {
			return func() tea.Msg { return editorDoneMsg{err: err} }
		}
	}

	lang := solve.NormalizeLanguageID(cfg.Language)
	spec, ok := solve.GetLanguageSpec(lang)
	if !ok {
		return func() tea.Msg { return editorDoneMsg{err: fmt.Errorf("unknown language %q", cfg.Language)} }
	}

	if _, _, err := solve.EnsureStarterFile(dir, cfg.Language); err != nil {
		return func() tea.Msg { return editorDoneMsg{err: err} }
	}

	filePath := filepath.Join(dir, spec.SourceFile)

	editorCmd, err := buildEditorCmd(filePath)
	if err != nil {
		return func() tea.Msg { return editorDoneMsg{err: err} }
	}

	return tea.ExecProcess(editorCmd, func(err error) tea.Msg {
		return editorDoneMsg{err: err}
	})
}

func buildEditorCmd(filePath string) (*exec.Cmd, error) {
	editor := strings.TrimSpace(os.Getenv("CPGRINDER_EDITOR"))
	if editor == "" {
		editor = strings.TrimSpace(os.Getenv("EDITOR"))
	}

	var argv []string
	if editor != "" {
		argv = strings.Fields(editor)
	} else {
		argv = defaultEditorArgv()
		if len(argv) == 0 {
			return nil, errors.New("no editor found; set $EDITOR (or $CPGRINDER_EDITOR)")
		}
	}

	bin, err := exec.LookPath(argv[0])
	if err != nil {
		return nil, fmt.Errorf("editor not found in PATH: %s", argv[0])
	}

	args := append([]string{}, argv[1:]...)

	// If use VS Code without waiting, add -w so cpgrinder waits until file closes.
	// if filepath.Base(argv[0]) == "code" && !hasWaitFlag(args) {
	// 	args = append([]string{"-w"}, args...)
	// }

	args = append(args, filePath)

	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

func defaultEditorArgv() []string {
	candidates := [][]string{
		// {"code", "-w"},
		{"code"},
		{"nvim"},
		{"vim"},
		{"nano"},
	}

	for _, c := range candidates {
		if _, err := exec.LookPath(c[0]); err == nil {
			return c
		}
	}
	return nil
}

func hasWaitFlag(args []string) bool {
	for _, a := range args {
		if a == "-w" || a == "--wait" {
			return true
		}
	}
	return false
}
