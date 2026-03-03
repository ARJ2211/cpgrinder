package solve

import "strings"

type CompareResult struct {
	OK            bool
	FirstDiffLine int
	ExpectedLine  string
	GotLine       string
}

func NormalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func CompareNormalized(expected, got string) CompareResult {
	e := NormalizeOutput(expected)
	g := NormalizeOutput(got)
	if e == g {
		return CompareResult{OK: true}
	}

	el := strings.Split(strings.TrimSuffix(e, "\n"), "\n")
	gl := strings.Split(strings.TrimSuffix(g, "\n"), "\n")

	max := len(el)
	if len(gl) > max {
		max = len(gl)
	}

	for i := 0; i < max; i++ {
		var a, b string
		if i < len(el) {
			a = el[i]
		}
		if i < len(gl) {
			b = gl[i]
		}
		if a != b {
			return CompareResult{
				OK:            false,
				FirstDiffLine: i + 1,
				ExpectedLine:  a,
				GotLine:       b,
			}
		}
	}

	return CompareResult{OK: false}
}
