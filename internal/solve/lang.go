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

/*
Language specifications that we will need
to run the different codes.
*/
type LanguageSpec struct {
	ID             LanguageID
	DisplayName    string
	SourceFile     string        // The file that needs to be opened
	TemplatePath   string        // The path to the template for that language (embed path, use forward slashes)
	RunCmd         []string      // CMD to run the file for that language
	DefaultTimeout time.Duration // In case of infinite loops etc.

	// ------------ OPTIONAL -------------
	IsCompiled      bool
	BuildDir        string // default for now .build
	CompileArgv     []string
	ArtifactRelPath string // for cpp later (ex: ".build/main")
}

// stable order (useful for a language picker later)
var languageOrder = []LanguageID{
	LangPython3,
	// add more later in the order you want them displayed
}

// registry (initialized once)
var specMap = map[LanguageID]LanguageSpec{
	LangPython3: python3Spec(),
	// add more later: LangJavaScript: jsSpec(), etc.
}

/*
Easier to parse configs
*/
func NormalizeLanguageID(s string) LanguageID {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return LanguageID(s)
}

/*
Get the default language
*/
func DefaultLanguage() LanguageID {
	return LangPython3
}

/*
List languages in a stable order (useful for UI later)
*/
func ListLanguageIDs() []LanguageID {
	out := make([]LanguageID, len(languageOrder))
	copy(out, languageOrder)
	return out
}

/*
List specs in stable order (useful for UI later)
*/
func ListLanguageSpecs() []LanguageSpec {
	out := make([]LanguageSpec, 0, len(languageOrder))
	for _, id := range languageOrder {
		if spec, ok := specMap[id]; ok {
			out = append(out, applyDefaults(spec))
		}
	}
	return out
}

/*
This will return the registered spec for the id
*/
func GetLanguageSpec(id LanguageID) (LanguageSpec, bool) {
	id = NormalizeLanguageID(string(id))

	spec, ok := specMap[id]
	if !ok {
		return LanguageSpec{}, false
	}

	return applyDefaults(spec), true
}

func applyDefaults(spec LanguageSpec) LanguageSpec {
	// timeouts
	if spec.DefaultTimeout == 0 {
		spec.DefaultTimeout = 2 * time.Second
	}
	// build dir default (for compiled langs later)
	if spec.BuildDir == "" && spec.IsCompiled {
		spec.BuildDir = ".build"
	}
	return spec
}

/*
Tiny helper to return the full path to the solution file
*/
func (s LanguageSpec) SourcePath(problemDir string) string {
	return filepath.Join(problemDir, s.SourceFile)
}

/*
Helper function to return the compiled path if there is any
*/
func (s LanguageSpec) ArtifactPath(problemDir string) string {
	if strings.TrimSpace(s.ArtifactRelPath) == "" {
		return ""
	}
	return filepath.Join(problemDir, s.ArtifactRelPath)
}

// ------------------ MAP BUILDER FOR THE SPECS -----------------

// PYTHON 3 SPEC
func python3Spec() LanguageSpec {
	return LanguageSpec{
		ID:             LangPython3,
		DisplayName:    "Python 3",
		SourceFile:     "main.py",
		TemplatePath:   "templates/python3/main.py.tmpl",
		RunCmd:         []string{"python3", "main.py"},
		DefaultTimeout: 2 * time.Second,

		IsCompiled:      false,
		BuildDir:        "",
		CompileArgv:     nil,
		ArtifactRelPath: "",
	}
}
