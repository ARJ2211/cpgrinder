package texlite

import (
	"regexp"
	"strings"
)

var (
	reText = regexp.MustCompile(`\\text\{([^{}]*)\}`)
	reFrac = regexp.MustCompile(`\\frac\{([^{}]+)\}\{([^{}]+)\}`)
	reCmd  = regexp.MustCompile(`\\([A-Za-z]+)`)
)

func HumanizeMathInMarkdown(md string) string {
	lines := strings.Split(md, "\n")
	inFence := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		lines[i] = replaceInlineMath(line)
	}

	return strings.Join(lines, "\n")
}

func replaceInlineMath(s string) string {
	var out strings.Builder
	var expr strings.Builder

	inMath := false
	delim := "" // "$" or "$$"

	i := 0
	for i < len(s) {
		if s[i] == '$' && !isEscaped(s, i) {
			// detect $ vs $$
			d := "$"
			if i+1 < len(s) && s[i+1] == '$' {
				d = "$$"
			}

			if !inMath {
				inMath = true
				delim = d
				expr.Reset()
				i += len(d)
				continue
			}

			if inMath && d == delim {
				// close
				out.WriteString(humanizeLatex(expr.String()))
				inMath = false
				delim = ""
				expr.Reset()
				i += len(d)
				continue
			}

			// mismatched delimiter inside math, treat as literal
			expr.WriteString(d)
			i += len(d)
			continue
		}

		if inMath {
			expr.WriteByte(s[i])
		} else {
			out.WriteByte(s[i])
		}
		i++
	}

	// if unclosed, keep original text
	if inMath {
		out.WriteString(delim)
		out.WriteString(expr.String())
	}

	return out.String()
}

func isEscaped(s string, idx int) bool {
	// count preceding backslashes; odd => escaped
	n := 0
	for j := idx - 1; j >= 0 && s[j] == '\\'; j-- {
		n++
	}
	return n%2 == 1
}

func humanizeLatex(expr string) string {
	e := strings.TrimSpace(expr)
	if e == "" {
		return ""
	}

	// Codeforces sometimes puts macro definitions at the top; just drop them.
	if strings.Contains(e, `\def\`) {
		return ""
	}

	// TeX line breaks
	e = strings.ReplaceAll(e, `\\`, "\n")

	// unwrap common constructs repeatedly
	for iter := 0; iter < 6; iter++ {
		next := e
		next = reText.ReplaceAllString(next, "$1")
		next = reFrac.ReplaceAllString(next, "($1/$2)")
		if next == e {
			break
		}
		e = next
	}

	// common symbol replacements
	repl := map[string]string{
		`\\le`: "≤", `\le`: "≤",
		`\\ge`: "≥", `\ge`: "≥",
		`\\lt`: "<", `\lt`: "<",
		`\\gt`: ">", `\gt`: ">",
		`\\in`: "∈", `\in`: "∈",

		`\\dots`: "...", `\dots`: "...",
		`\\ldots`: "...", `\ldots`: "...",
		`\\cdots`: "...", `\cdots`: "...",

		`\\cdot`: "·", `\cdot`: "·",
		`\\times`: "×", `\times`: "×",

		`\\rightarrow`: "->", `\rightarrow`: "->",
		`\\to`: "->", `\to`: "->",

		`\\left`: "", `\left`: "",
		`\\right`: "", `\right`: "",
	}
	for k, v := range repl {
		e = strings.ReplaceAll(e, k, v)
	}

	// bracket/brace escapes often show as \[1,2\] or \{...\}
	e = strings.ReplaceAll(e, `\[`, "[")
	e = strings.ReplaceAll(e, `\]`, "]")
	e = strings.ReplaceAll(e, `\{`, "{")
	e = strings.ReplaceAll(e, `\}`, "}")

	// common escapes
	e = strings.ReplaceAll(e, `\_`, "_")

	// strip remaining commands like \RED, \BLUE, \text (if any are left)
	e = reCmd.ReplaceAllString(e, "$1")

	// remove grouping braces after we handled frac/text
	e = strings.ReplaceAll(e, "{", "")
	e = strings.ReplaceAll(e, "}", "")

	// normalize whitespace a bit
	e = strings.ReplaceAll(e, "\t", " ")
	for strings.Contains(e, "  ") {
		e = strings.ReplaceAll(e, "  ", " ")
	}
	e = strings.TrimSpace(e)

	// prevent markdown italics for variables like a_i
	e = strings.ReplaceAll(e, "_", `\_`)

	return e
}
