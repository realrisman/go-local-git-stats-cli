package main

import (
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestExpandHome(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("user.Current() failed: %v", err)
	}
	home := usr.HomeDir

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"bare tilde", "~", home},
		{"tilde slash", "~/Codes/psy-care", home + "/Codes/psy-care"},
		{"absolute path unchanged", "/Users/you/Codes", "/Users/you/Codes"},
		{"relative path unchanged", "Codes/psy-care", "Codes/psy-care"},
		{"empty string unchanged", "", ""},
		{"tilde user not expanded", "~other/Codes", "~other/Codes"},
		{"tilde mid-path unchanged", "/opt/~/Codes", "/opt/~/Codes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expandHome(tt.in); got != tt.want {
				t.Errorf("expandHome(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSliceContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	if !sliceContains(slice, "b") {
		t.Error("sliceContains should find an existing element")
	}
	if sliceContains(slice, "z") {
		t.Error("sliceContains should not find a missing element")
	}
	if sliceContains(nil, "a") {
		t.Error("sliceContains on nil slice should be false")
	}
}

func TestJoinSlices(t *testing.T) {
	existing := []string{"a", "b"}
	got := joinSlices([]string{"b", "c", "d"}, existing)
	want := []string{"a", "b", "c", "d"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("joinSlices = %v, want %v (duplicates should be skipped)", got, want)
	}
}

func TestFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "repos")

	// parsing a freshly created (empty) file yields no lines
	if got := parseFileLinesToSlice(path); len(got) != 0 {
		t.Errorf("parseFileLinesToSlice on empty file = %v, want empty", got)
	}

	repos := []string{"/one", "/two"}
	dumpStringSliceToFile(repos, path)
	if got := parseFileLinesToSlice(path); !reflect.DeepEqual(got, repos) {
		t.Errorf("round trip = %v, want %v", got, repos)
	}

	// adding overlapping + new entries de-duplicates against existing content
	addNewSliceElementsToFile(path, []string{"/two", "/three"})
	want := []string{"/one", "/two", "/three"}
	if got := parseFileLinesToSlice(path); !reflect.DeepEqual(got, want) {
		t.Errorf("addNewSliceElementsToFile = %v, want %v", got, want)
	}
}

func TestRecursiveScanFolder(t *testing.T) {
	root := t.TempDir()
	// Two real repos (one nested), plus repos under vendor/ and node_modules/
	// which scanGitFolders must skip.
	for _, dir := range []string{
		"repoA/.git",
		"group/repoB/.git",
		"vendor/skipme/.git",
		"node_modules/skipme/.git",
	} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	got := recursiveScanFolder(root)
	sort.Strings(got)
	want := []string{
		filepath.Join(root, "group/repoB"),
		filepath.Join(root, "repoA"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("recursiveScanFolder = %v, want %v (vendor/node_modules should be skipped)", got, want)
	}
}
