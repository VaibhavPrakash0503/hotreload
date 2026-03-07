package watcher

import (
	"path/filepath"
)

type Filter struct {
	ignoreDirs map[string]bool
	ignoreExts map[string]bool
}

func NewFilter() *Filter {
	return &Filter{
		ignoreDirs: map[string]bool{
			// version control
			".git":    true,
			".github": true,
			".svn":    true,

			// dependencies
			"node_modules": true,
			"vendor":       true,
			"venv":         true,
			".venv":        true,

			// build artifacts
			"dist":  true,
			"build": true,
			"bin":   true,
			"out":   true,

			// IDE and system files
			".vscode": true,
			".idea":   true,

			// cache
			".cache":      true,
			".next":       true,
			".nuxt":       true,
			"__pycache__": true,
		},
		ignoreExts: map[string]bool{
			".log":      true,
			".tmp":      true,
			".temp":     true,
			".swp":      true,
			".bak":      true,
			".DS_Store": true,
			".exe":      true,
			".dll":      true,
			".so":       true,
			".pkl":      true,
			".o":        true,
			".a":        true,
			".zip":      true,
			".tar":      true,
			".gz":       true,
		},
	}
}

func (f *Filter) ShouldIgnoreDir(path string) bool {
	base := filepath.Base(path)
	return f.ignoreDirs[base]
}

func (f *Filter) ShouldIgnoreExt(path string) bool {
	ext := filepath.Ext(path)
	return f.ignoreExts[ext]
}
