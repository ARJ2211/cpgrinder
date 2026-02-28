package texlite

import (
	"regexp"
	"strings"
)

var (
	// Convert (\\){2,}\leq -> \leq  (works for \\leq, \\\\leq, etc)
	reMultiSlashCmd = regexp.MustCompile(`(\\){2,}([A-Za-z])`)

	// \(...\) and \[...\] outside $...$
	reParenMath = regexp.MustCompile(`\\\((.*?)\\\)`)
	reBrackMath = regexp.MustCompile(`\\\[(.*?)\\\]`)

	reText = regexp.MustCompile(`\\text\{([^{}]*)\}`)
	reFrac = regexp.MustCompile(`\\frac\{([^{}]+)\}\{([^{}]+)\}`)
	reSub  = regexp.MustCompile(`_\{([^{}]+)\}`)
	reSup  = regexp.MustCompile(`\^\{([^{}]+)\}`)

	// Any remaining \COMMAND
	reCmd = regexp.MustCompile(`\\([A-Za-z]+)`)
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

		// Handle \(...\) and \[...\] outside $...$
		line = reParenMath.ReplaceAllStringFunc(line, func(m string) string {
			inner := m[len(`\(`) : len(m)-len(`\)`)]
			return humanizeLatex(inner)
		})
		line = reBrackMath.ReplaceAllStringFunc(line, func(m string) string {
			inner := m[len(`\[`) : len(m)-len(`\]`)]
			return humanizeLatex(inner)
		})

		// Handle $...$ and $$...$$
		lines[i] = replaceDollarMath(line)
	}

	return strings.Join(lines, "\n")
}

func replaceDollarMath(s string) string {
	var out strings.Builder
	var expr strings.Builder

	inMath := false
	delim := "" // "$" or "$$"

	i := 0
	for i < len(s) {
		if s[i] == '$' && !isEscaped(s, i) {
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
				out.WriteString(humanizeLatex(expr.String()))
				inMath = false
				delim = ""
				expr.Reset()
				i += len(d)
				continue
			}

			// mismatched delimiter inside math -> keep literal
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

	if inMath {
		out.WriteString(delim)
		out.WriteString(expr.String())
	}

	return out.String()
}

func isEscaped(s string, idx int) bool {
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

	// 1) normalize \\leq, \\\\dots, etc -> \leq, \dots
	// replacement must be ONE backslash + the letter
	e = reMultiSlashCmd.ReplaceAllString(e, "\\$2")

	// 2) drop \def macro lines (they're noise in terminal)
	if strings.Contains(e, `\def\`) {
		var kept []string
		for _, ln := range strings.Split(e, "\n") {
			if strings.Contains(ln, `\def\`) {
				continue
			}
			kept = append(kept, ln)
		}
		e = strings.TrimSpace(strings.Join(kept, " "))
		if e == "" {
			return ""
		}
	}

	// 3) TeX linebreak \\ -> space
	e = strings.ReplaceAll(e, `\\`, " ")

	// 4) unwrap common constructs a few times
	for iter := 0; iter < 8; iter++ {
		next := e
		next = reText.ReplaceAllString(next, "$1")
		next = reFrac.ReplaceAllString(next, "($1/$2)")
		next = reSub.ReplaceAllString(next, "_$1")
		next = reSup.ReplaceAllString(next, "^$1")
		if next == e {
			break
		}
		e = next
	}

	// 5) replacements (include \leq/\dots/\times etc)
	repl := map[string]string{
		`\leqslant`: "≤", `\geqslant`: "≥",
		`\leq`: "≤", `\geq`: "≥",
		`\le`: "≤", `\ge`: "≥",
		`\lt`: "<", `\gt`: ">",

		`\neq`: "≠", `\ne`: "≠",
		`\approx`: "≈",
		`\pm`:     "±",

		`\in`:   "∈",
		`\dots`: "...", `\ldots`: "...", `\cdots`: "...",
		`\cdot`: "·", `\times`: "×",

		`\rightarrow`: "->", `\to`: "->",

		`\left`: "", `\right`: "",

		// spacing commands
		`\,`: " ", `\;`: " ", `\:`: " ",
	}
	for k, v := range repl {
		e = strings.ReplaceAll(e, k, v)
	}

	// common escapes
	e = strings.ReplaceAll(e, `\_`, "_")
	e = strings.ReplaceAll(e, `\[`, "[")
	e = strings.ReplaceAll(e, `\]`, "]")
	e = strings.ReplaceAll(e, `\{`, "{")
	e = strings.ReplaceAll(e, `\}`, "}")

	// 6) strip leftover commands (\RED -> RED, \alpha -> alpha)
	e = reCmd.ReplaceAllString(e, "$1")

	// normalize whitespace
	e = strings.ReplaceAll(e, "\t", " ")
	for strings.Contains(e, "  ") {
		e = strings.ReplaceAll(e, "  ", " ")
	}
	e = strings.TrimSpace(e)

	// 7) stop markdown italics in the final rendered markdown
	// Glamour will show the underscore without italicizing.
	e = strings.ReplaceAll(e, "_", `\_`)

	return e
}
