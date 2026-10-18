package execenv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func readClaudeSettings(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func claudeMdExcludesOf(t *testing.T, settings map[string]any) []string {
	t.Helper()
	raw, _ := settings["claudeMdExcludes"].([]any)
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		out = append(out, v.(string))
	}
	return out
}

func TestExcludeClaudeMdCreatesSettings(t *testing.T) {
	t.Parallel()
	env := &Environment{RootDir: t.TempDir()}
	md := filepath.Join(t.TempDir(), "CLAUDE.md")

	if err := ExcludeClaudeMd(env, md); err != nil {
		t.Fatal(err)
	}
	if env.ClaudeSettingsPath != filepath.Join(env.RootDir, claudeRuntimeSkillSettingsFile) {
		t.Fatalf("ClaudeSettingsPath = %q", env.ClaudeSettingsPath)
	}
	if got := claudeMdExcludesOf(t, readClaudeSettings(t, env.ClaudeSettingsPath)); !slices.Equal(got, []string{md}) {
		t.Fatalf("claudeMdExcludes = %v", got)
	}
}

func TestExcludeClaudeMdKeepsSkillPolicy(t *testing.T) {
	t.Parallel()
	env := &Environment{RootDir: t.TempDir()}
	path, err := prepareClaudeSkillSettings(env.RootDir, []RuntimeSkillRefForEnv{{Key: "user/review", Name: "review"}}, nil)
	if err != nil || path == "" {
		t.Fatalf("prepare skill settings: path=%q err=%v", path, err)
	}
	env.ClaudeSettingsPath = path
	md := filepath.Join(t.TempDir(), "CLAUDE.md")

	for range 2 {
		if err := ExcludeClaudeMd(env, md); err != nil {
			t.Fatal(err)
		}
	}
	settings := readClaudeSettings(t, path)
	if got := claudeMdExcludesOf(t, settings); !slices.Equal(got, []string{md}) {
		t.Fatalf("claudeMdExcludes = %v, want exactly one entry", got)
	}
	if _, ok := settings["skillOverrides"]; !ok {
		t.Fatalf("skillOverrides dropped: %v", settings)
	}
	if _, ok := settings["permissions"]; !ok {
		t.Fatalf("permissions dropped: %v", settings)
	}
}
