package structure

import (
	"strings"
	"testing"

	"github.com/agent-ecosystem/skill-validator/types"
)

func TestTokenLimitCodes(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		files   map[string]string
		opts    Options
		message string
		level   types.Level
		code    string
	}{
		{
			name: "body soft limit", body: strings.Repeat("word ", 5_500),
			message: "spec recommends < 5000", level: types.Warning, code: types.CodeTokenLimit,
		},
		{
			name: "line count is not a token limit", body: strings.Repeat("line\n", 501),
			message: "spec recommends < 500)", level: types.Warning,
		},
		{
			name: "reference soft limit", files: map[string]string{"references/a.md": strings.Repeat("word ", 11_000)},
			message: "consider splitting", level: types.Warning, code: types.CodeTokenLimit,
		},
		{
			name: "reference hard limit", files: map[string]string{"references/a.md": strings.Repeat("word ", 26_000)},
			message: "meaningfully degrade", level: types.Error, code: types.CodeTokenLimit,
		},
		{
			name: "flat layout soft limit", files: map[string]string{"guide.md": strings.Repeat("word ", 11_000)},
			opts:    Options{AllowFlatLayouts: true},
			message: "consider splitting", level: types.Warning, code: types.CodeTokenLimit,
		},
		{
			name: "flat layout hard limit", files: map[string]string{"guide.md": strings.Repeat("word ", 26_000)},
			opts:    Options{AllowFlatLayouts: true},
			message: "meaningfully degrade", level: types.Error, code: types.CodeTokenLimit,
		},
		{
			name: "reference aggregate soft limit", files: map[string]string{
				"references/a.md": strings.Repeat("word ", 14_000),
				"references/b.md": strings.Repeat("word ", 14_000),
			},
			message: "total reference files:", level: types.Warning, code: types.CodeTokenLimit,
		},
		{
			name: "reference aggregate hard limit", files: map[string]string{
				"references/a.md": strings.Repeat("word ", 18_000),
				"references/b.md": strings.Repeat("word ", 18_000),
				"references/c.md": strings.Repeat("word ", 18_000),
			},
			message: "total reference files:", level: types.Error, code: types.CodeTokenLimit,
		},
		{
			name: "other aggregate soft limit", files: map[string]string{"notes.md": strings.Repeat("word ", 30_000)},
			message: "non-standard files total", level: types.Warning, code: types.CodeTokenLimit,
		},
		{
			name: "other aggregate hard limit", files: map[string]string{"notes.md": strings.Repeat("word ", 101_000)},
			message: "non-standard files total", level: types.Error, code: types.CodeTokenLimit,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for path, content := range tt.files {
				writeFile(t, dir, path, content)
			}
			results, _, _ := CheckTokens(dir, tt.body, tt.opts)
			for _, result := range results {
				if strings.Contains(result.Message, tt.message) {
					if result.Level != tt.level || result.Code != tt.code {
						t.Errorf("got level=%v code=%q, want level=%v code=%q", result.Level, result.Code, tt.level, tt.code)
					}
					return
				}
			}
			t.Fatalf("no finding containing %q: %+v", tt.message, results)
		})
	}
}

func TestTruncatedTokenCountIsNotTokenLimit(t *testing.T) {
	results := truncationNotes(types.ResultContext{Category: "Tokens"}, []types.TokenCount{
		{File: "references/large.md", Truncated: true},
	})
	if len(results) != 1 || results[0].Level != types.Warning || results[0].Code != "" {
		t.Fatalf("truncation warning must not be classified as a token limit: %+v", results)
	}
}
