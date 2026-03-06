package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type options struct {
	write bool
}

func main() {
	var opts options
	flag.BoolVar(&opts.write, "write", false, "")
	flag.Parse()

	files, err := gitTrackedFiles()
	if err != nil {
		exitErr(err)
	}

	var changed int
	for _, path := range files {
		if !strings.HasSuffix(path, ".go") {
			continue
		}
		updated, err := stripGoFile(path, opts.write)
		if err != nil {
			exitErr(fmt.Errorf("%s: %w", path, err))
		}
		if updated {
			changed++
		}
	}

	if !opts.write {
		fmt.Printf("%d file(s) would change\n", changed)
		return
	}
	fmt.Printf("%d file(s) changed\n", changed)
}

func gitTrackedFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	lines := strings.Split(out.String(), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		files = append(files, filepath.FromSlash(line))
	}
	return files, nil
}

func stripGoFile(path string, write bool) (bool, error) {
	orig, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, orig, parser.ParseComments)
	if err != nil {
		return false, err
	}

	keep := keepCommentGroups(f)
	f.Comments = keep

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return false, err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return false, err
	}

	if bytes.Equal(orig, formatted) {
		return false, nil
	}

	if !write {
		return true, nil
	}

	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func keepCommentGroups(f *ast.File) []*ast.CommentGroup {
	pkgPos := f.Package
	var keep []*ast.CommentGroup
	for _, cg := range f.Comments {
		if shouldKeepDirectiveGroup(cg) {
			keep = append(keep, cg)
			continue
		}
		if cg.End() <= pkgPos && looksLikeLicense(cg.Text()) {
			keep = append(keep, cg)
			continue
		}
	}
	return keep
}

func shouldKeepDirectiveGroup(cg *ast.CommentGroup) bool {
	for _, c := range cg.List {
		t := strings.TrimSpace(c.Text)
		if strings.HasPrefix(t, "//go:") {
			return true
		}
		if strings.HasPrefix(t, "// +build") || strings.HasPrefix(t, "//+build") {
			return true
		}
		if strings.HasPrefix(t, "//line ") {
			return true
		}
		if strings.HasPrefix(t, "// Code generated ") && strings.Contains(t, "DO NOT EDIT") {
			return true
		}
	}
	return false
}

func looksLikeLicense(text string) bool {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "spdx-") {
		return true
	}
	if strings.Contains(lower, "copyright") {
		return true
	}
	if strings.Contains(lower, "license") {
		return true
	}
	if strings.Contains(lower, "apache") || strings.Contains(lower, "mit") || strings.Contains(lower, "mozilla") {
		return true
	}
	if strings.Contains(lower, "gnu general public license") || strings.Contains(lower, "gpl") {
		return true
	}
	return false
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
