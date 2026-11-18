package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllowTokenWarningsCLI(t *testing.T) {
	bin := buildBinary(t)
	words := func(n int) string { return strings.Repeat("word ", n) }
	tests := []struct {
		name       string
		body       string
		files      map[string]string
		extraFlags []string
		multi      bool
		otherSkill bool // add an unrelated line-count warning in another skill
		wantCode   int  // with --allow-token-warnings, without --strict
		want       string
	}{
		{
			name: "body warning", body: words(5_500),
			wantCode: 0, want: "spec recommends < 5000",
		},
		{
			name: "per-file and aggregate warnings from CI", files: map[string]string{
				"references/quick-start.md": words(11_000),
				"references/a.md":           words(8_000),
				"references/b.md":           words(8_000),
				"references/c.md":           words(8_000),
			},
			wantCode: 0, want: "**Result: 2 warnings**",
		},
		{
			name: "aggregate warning across skills", multi: true, files: map[string]string{
				"references/a.md": words(9_000),
				"references/b.md": words(9_000),
				"references/c.md": words(9_000),
			},
			wantCode: 0, want: "total reference files:",
		},
		{
			name: "flat layout warning", files: map[string]string{"guide.md": words(11_000)},
			extraFlags: []string{"--allow-flat-layouts"}, wantCode: 0, want: "consider splitting",
		},
		{
			name: "line warning remains blocking", body: words(5_500) + strings.Repeat("\n", 501),
			wantCode: 2, want: "spec recommends < 500)",
		},
		{
			name: "unrelated error remains blocking", body: words(5_500) + "\n[Missing](missing.md)\n",
			wantCode: 1, want: "::error file=",
		},
		{
			name: "per-file hard limit remains blocking", files: map[string]string{"references/huge.md": words(26_000)},
			wantCode: 1, want: "meaningfully degrade agent performance",
		},
		{
			name: "aggregate hard limit remains blocking", multi: true, files: map[string]string{
				"references/a.md": words(18_000),
				"references/b.md": words(18_000),
				"references/c.md": words(18_000),
			},
			wantCode: 1, want: "context window; reduce content",
		},
		{
			name: "other skill warning remains blocking", body: words(5_500), multi: true, otherSkill: true,
			wantCode: 2, want: "spec recommends < 500)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "token-skill")
			writeSkill := func(path, body string, files map[string]string) {
				t.Helper()
				for name, content := range files {
					file := filepath.Join(path, filepath.FromSlash(name))
					if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
						t.Fatal(err)
					}
					body += "\n[Guide](" + name + ")\n"
				}
				if err := os.MkdirAll(path, 0o755); err != nil {
					t.Fatal(err)
				}
				content := "---\nname: " + filepath.Base(path) + "\ndescription: Validate token budgets.\nmetadata:\n  labelText: Validate token budgets\n---\n# Token budgets\n\n" + body
				if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			writeSkill(dir, tt.body, tt.files)
			target := dir
			if tt.multi {
				target = root
				body := "A small skill."
				if tt.otherSkill {
					body = strings.Repeat("line\n", 501)
				}
				writeSkill(filepath.Join(root, "other-skill"), body, nil)
			}
			for _, command := range [][]string{{"check", "--only=structure"}, {"validate", "structure"}} {
				for _, strict := range []bool{false, true} {
					args := append([]string{}, command...)
					args = append(args, "--emit-annotations", "-o", "markdown")
					args = append(args, tt.extraFlags...)
					if strict {
						args = append(args, "--strict")
					}
					var baseline string
					for _, allow := range []bool{false, true} {
						runArgs := append([]string{}, args...)
						if allow {
							runArgs = append(runArgs, "--allow-token-warnings")
						}
						runArgs = append(runArgs, target)
						cmd := exec.Command(bin, runArgs...)
						out, _ := cmd.CombinedOutput()
						wantCode := tt.wantCode
						if !allow && wantCode == 0 {
							wantCode = 2
						}
						if strict && wantCode == 2 {
							wantCode = 1
						}
						if got := cmd.ProcessState.ExitCode(); got != wantCode {
							t.Errorf("%v: exit code = %d, want %d\n%s", runArgs, got, wantCode, out)
						}
						for _, want := range []string{tt.want, "::warning", "**Warning:**", "token-skill", "| File | Tokens |"} {
							if !strings.Contains(string(out), want) {
								t.Errorf("%v: output missing %q:\n%s", runArgs, want, out)
							}
						}
						if !allow {
							baseline = string(out)
						} else if string(out) != baseline {
							t.Errorf("%v: allowing token warnings changed the report or annotations", runArgs)
						}
					}
				}
			}
		})
	}
}
