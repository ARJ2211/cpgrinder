package solve

import (
	"path/filepath"
	"strings"
	"time"
)

type LanguageID string

const (
	LangPython3    LanguageID = "python3"
	LangJavaScript LanguageID = "javascript"
	LangCPP        LanguageID = "cpp"
	LangGo         LanguageID = "go"
	LangJava       LanguageID = "java"
)

type LanguageSpec struct {
	ID             LanguageID
	DisplayName    string
	SourceFile     string        // file to edit
	TemplatePath   string        // embed path, forward slashes
	RunCmd         []string      // argv to execute (relative OK)
	DefaultTimeout time.Duration // per run

	// compiled languages
	IsCompiled      bool
	BuildDir        string   // default ".build" for compiled
	CompileArgv     []string // argv to compile
	ArtifactRelPath string   // optional: ".build/main"
}

// stable order (useful for picker later)
var languageOrder = []LanguageID{
	LangPython3,
	LangJavaScript,
	LangCPP,
	LangGo,
	LangJava,
}

// registry
var specMap = map[LanguageID]LanguageSpec{
	LangPython3:    python3Spec(),
	LangJavaScript: jsSpec(),
	LangCPP:        cppSpec(),
	LangGo:         goSpec(),
	LangJava:       javaSpec(),
}

func NormalizeLanguageID(s string) LanguageID {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return LanguageID(s)
}

func DefaultLanguage() LanguageID { return LangPython3 }

func ListLanguageIDs() []LanguageID {
	out := make([]LanguageID, len(languageOrder))
	copy(out, languageOrder)
	return out
}

func ListLanguageSpecs() []LanguageSpec {
	out := make([]LanguageSpec, 0, len(languageOrder))
	for _, id := range languageOrder {
		if spec, ok := specMap[id]; ok {
			out = append(out, applyDefaults(spec))
		}
	}
	return out
}

func GetLanguageSpec(id LanguageID) (LanguageSpec, bool) {
	id = NormalizeLanguageID(string(id))

	spec, ok := specMap[id]
	if !ok {
		return LanguageSpec{}, false
	}
	return applyDefaults(spec), true
}

func applyDefaults(spec LanguageSpec) LanguageSpec {
	if spec.DefaultTimeout == 0 {
		spec.DefaultTimeout = 2 * time.Second
	}
	if spec.BuildDir == "" && spec.IsCompiled {
		spec.BuildDir = ".build"
	}
	return spec
}

func (s LanguageSpec) SourcePath(problemDir string) string {
	return filepath.Join(problemDir, s.SourceFile)
}

func (s LanguageSpec) ArtifactPath(problemDir string) string {
	if strings.TrimSpace(s.ArtifactRelPath) == "" {
		return ""
	}
	return filepath.Join(problemDir, s.ArtifactRelPath)
}

// ------------------ SPECS ------------------

func python3Spec() LanguageSpec {
	return LanguageSpec{
		ID:             LangPython3,
		DisplayName:    "Python 3",
		SourceFile:     "main.py",
		TemplatePath:   "templates/python3/main.py.tmpl",
		RunCmd:         []string{"python3", "main.py"},
		DefaultTimeout: 2 * time.Second,
		IsCompiled:     false,
	}
}

func jsSpec() LanguageSpec {
	return LanguageSpec{
		ID:             LangJavaScript,
		DisplayName:    "JavaScript (Node)",
		SourceFile:     "main.js",
		TemplatePath:   "templates/javascript/main.js.tmpl",
		RunCmd:         []string{"node", "main.js"},
		DefaultTimeout: 2 * time.Second,
		IsCompiled:     false,
	}
}

func cppSpec() LanguageSpec {
	return LanguageSpec{
		ID:           LangCPP,
		DisplayName:  "C++",
		SourceFile:   "main.cpp",
		TemplatePath: "templates/cpp/main.cpp.tmpl",

		IsCompiled:      true,
		BuildDir:        ".build",
		CompileArgv:     []string{"g++", "main.cpp", "-O2", "-std=c++17", "-o", ".build/main"},
		RunCmd:          []string{"./.build/main"},
		DefaultTimeout:  2 * time.Second,
		ArtifactRelPath: ".build/main",
	}
}

func goSpec() LanguageSpec {
	return LanguageSpec{
		ID:           LangGo,
		DisplayName:  "Go",
		SourceFile:   "main.go",
		TemplatePath: "templates/go/main.go.tmpl",

		IsCompiled:      true,
		BuildDir:        ".build",
		CompileArgv:     []string{"go", "build", "-o", ".build/main", "main.go"},
		RunCmd:          []string{"./.build/main"},
		DefaultTimeout:  2 * time.Second,
		ArtifactRelPath: ".build/main",
	}
}

func javaSpec() LanguageSpec {
	return LanguageSpec{
		ID:           LangJava,
		DisplayName:  "Java",
		SourceFile:   "Main.java",
		TemplatePath: "templates/java/Main.java.tmpl",

		IsCompiled: true,
		BuildDir:   ".build",

		// compile to Java 8 bytecode so even older java runtimes can run it
		CompileArgv: []string{"javac", "-source", "8", "-target", "8", "-d", ".build", "Main.java"},

		RunCmd:          []string{"java", "-cp", ".build", "Main"},
		DefaultTimeout:  2 * time.Second,
		ArtifactRelPath: "",
	}
}
