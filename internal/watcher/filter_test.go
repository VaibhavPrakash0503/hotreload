package watcher

import (
	"testing"
)

func TestShouldIgnoreDir(t *testing.T) {
	f := NewFilter()

	ignored := []string{
		".git", ".github", ".svn",
		"node_modules", "vendor", "venv", ".venv",
		"dist", "build", "bin", "out",
		".vscode", ".idea",
		".cache", ".next", ".nuxt", "__pycache__",
	}
	for _, dir := range ignored {
		if !f.ShouldIgnoreDir(dir) {
			t.Errorf("expected %q to be ignored", dir)
		}
	}

	allowed := []string{"internal", "cmd", "pkg", "handlers", "myapp"}
	for _, dir := range allowed {
		if f.ShouldIgnoreDir(dir) {
			t.Errorf("expected %q NOT to be ignored", dir)
		}
	}
}

func TestShouldIgnoreExt(t *testing.T) {
	f := NewFilter()

	ignored := []string{
		"app.log", "temp.tmp", "swap.swp",
		"backup.bak", "notes.DS_Store",
		"program.exe", "lib.dll", "module.so",
		"model.pkl", "obj.o", "archive.a",
		"release.zip", "tarball.tar", "compressed.gz",
	}
	for _, file := range ignored {
		if !f.ShouldIgnoreExt(file) {
			t.Errorf("expected %q to be ignored by extension", file)
		}
	}

	allowed := []string{
		"main.go", "index.js", "app.py",
		"README.md", "config.yaml", "Dockerfile",
	}
	for _, file := range allowed {
		if f.ShouldIgnoreExt(file) {
			t.Errorf("expected %q NOT to be ignored by extension", file)
		}
	}
}

// TestShouldIgnoreDirNestedPath verifies the filter uses basename only,
// so a full path like "a/b/.git/config" is checked against ".git".
func TestShouldIgnoreDirNestedPath(t *testing.T) {
	f := NewFilter()

	if !f.ShouldIgnoreDir("some/deep/path/.git") {
		t.Error("expected full path ending in .git to be ignored")
	}
	if f.ShouldIgnoreDir("some/deep/path/internal") {
		t.Error("expected full path ending in 'internal' NOT to be ignored")
	}
}
