package cmd

import (
	"testing"

	"github.com/agent-ecosystem/skill-validator/types"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name     string
		opts     exitOpts
		errors   int
		warnings int
		want     int
	}{
		{"clean", exitOpts{}, 0, 0, ExitClean},
		{"errors only", exitOpts{}, 3, 0, ExitError},
		{"warnings only", exitOpts{}, 0, 2, ExitWarning},
		{"errors and warnings", exitOpts{}, 1, 5, ExitError},
		{"strict clean", exitOpts{strict: true}, 0, 0, ExitClean},
		{"strict errors", exitOpts{strict: true}, 1, 0, ExitError},
		{"strict warnings", exitOpts{strict: true}, 0, 4, ExitError},
		{"strict errors and warnings", exitOpts{strict: true}, 2, 3, ExitError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.resolve(tt.errors, tt.warnings)
			if got != tt.want {
				t.Errorf("resolve(%d, %d) with strict=%v = %d, want %d",
					tt.errors, tt.warnings, tt.opts.strict, got, tt.want)
			}
		})
	}
}

func TestAllowTokenWarnings(t *testing.T) {
	tests := []struct {
		name    string
		results []types.Result
		want    int
	}{
		{"token warning", []types.Result{{Level: types.Warning, Code: types.CodeTokenLimit}}, ExitClean},
		{"token error", []types.Result{{Level: types.Error, Code: types.CodeTokenLimit}}, ExitError},
		{"other warning", []types.Result{{Level: types.Warning, Category: "Frontmatter"}}, ExitWarning},
		{"non-budget token warning", []types.Result{{Level: types.Warning, Category: "Tokens"}}, ExitWarning},
		{"mixed warnings", []types.Result{
			{Level: types.Warning, Code: types.CodeTokenLimit},
			{Level: types.Warning, Category: "Structure"},
		}, ExitWarning},
		{"token warning with error", []types.Result{
			{Level: types.Warning, Code: types.CodeTokenLimit},
			{Level: types.Error, Category: "Frontmatter"},
		}, ExitError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &types.Report{Results: tt.results}
			r.Tally()
			originalWarnings := r.Warnings
			for _, strict := range []bool{false, true} {
				opts := exitOpts{strict: strict, allowTokenWarnings: true}
				want := tt.want
				if strict && want == ExitWarning {
					want = ExitError
				}
				if got := opts.resolve(r.Errors, opts.warningCount(r)); got != want {
					t.Errorf("strict=%v: got exit code %d, want %d", strict, got, want)
				}
				if r.Warnings != originalWarnings {
					t.Error("exit policy changed the report warning count")
				}
			}
			if got := (exitOpts{}).warningCount(r); got != originalWarnings {
				t.Errorf("default policy: got %d warnings, want %d", got, originalWarnings)
			}
		})
	}
}
