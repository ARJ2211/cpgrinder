package solve

import (
	"context"
	"errors"
	"time"

	"github.com/ARJ2211/cpgrinder/internal/store"
)

type SampleCaseResult struct {
	Name     string
	Input    string
	Expected string
	Got      string

	Run  RunResult
	Diff CompareResult
	OK   bool
}

type SamplesResult struct {
	Compile   *RunResult
	Cases     []SampleCaseResult
	AllPassed bool
}

func RunSamples(ctx context.Context, spec LanguageSpec, dir string, samples []store.Sample, timeoutOverride time.Duration) (SamplesResult, error) {
	if len(samples) == 0 {
		return SamplesResult{}, errors.New("no samples available")
	}

	r := DefaultRunner()
	spec = applyDefaults(spec)

	if spec.IsCompiled {
		compileRes, ok, err := r.compileIfNeeded(ctx, spec, dir)
		if err != nil {
			return SamplesResult{}, err
		}
		if !ok {
			return SamplesResult{Compile: &compileRes, AllPassed: false}, nil
		}
	}

	out := SamplesResult{Cases: make([]SampleCaseResult, 0, len(samples))}
	all := true

	for _, s := range samples {
		res, err := r.runOnly(ctx, spec, dir, []byte(s.In), timeoutOverride)
		if err != nil {
			return SamplesResult{}, err
		}

		caseRes := SampleCaseResult{
			Name:     s.Name,
			Input:    s.In,
			Expected: s.Out,
			Got:      res.Stdout,
			Run:      res,
		}

		if res.TimedOut || res.ExitCode != 0 {
			caseRes.OK = false
			all = false
			out.Cases = append(out.Cases, caseRes)
			continue
		}

		diff := CompareNormalized(s.Out, res.Stdout)
		caseRes.Diff = diff
		caseRes.OK = diff.OK
		if !caseRes.OK {
			all = false
		}

		out.Cases = append(out.Cases, caseRes)
	}

	out.AllPassed = all
	return out, nil
}
