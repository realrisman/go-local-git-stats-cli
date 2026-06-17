package main

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// commitTo creates a git repo at dir with a single commit authored by email.
func commitTo(t *testing.T, dir, email string) {
	t.Helper()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("PlainInit: %v", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := wt.Add("file.txt"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	_, err = wt.Commit("initial", &git.CommitOptions{
		Author: &object.Signature{Name: "Tester", Email: email, When: time.Now()},
	})
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
}

func sumCommits(m map[int]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}

func TestFillCommits(t *testing.T) {
	const email = "me@example.com"
	repo := t.TempDir()
	commitTo(t, repo, email)

	commits := make(map[int]int)

	// A matching author email is counted.
	commits = fillCommits(email, repo, commits)
	if got := sumCommits(commits); got != 1 {
		t.Errorf("fillCommits counted %d matching commits, want 1", got)
	}

	// A non-matching author email is ignored.
	commits = fillCommits("someone-else@example.com", repo, commits)
	if got := sumCommits(commits); got != 1 {
		t.Errorf("fillCommits added a non-matching commit, total = %d, want 1", got)
	}
}

func TestProcessRepositories(t *testing.T) {
	const email = "me@example.com"
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	commitTo(t, repo, email)

	// A dot file listing the repo path, just like getDotFilePath would point to.
	dotFile := filepath.Join(t.TempDir(), "dotfile")
	if err := os.WriteFile(dotFile, []byte(repo+"\n"), 0o644); err != nil {
		t.Fatalf("write dotfile: %v", err)
	}

	commits := processRepositories(email, dotFile)
	if got := sumCommits(commits); got != 1 {
		t.Errorf("processRepositories counted %d commits, want 1", got)
	}
}

func TestStats(t *testing.T) {
	const email = "me@example.com"
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	commitTo(t, repo, email)

	dotFile := filepath.Join(t.TempDir(), "dotfile")
	if err := os.WriteFile(dotFile, []byte(repo+"\n"), 0o644); err != nil {
		t.Fatalf("write dotfile: %v", err)
	}

	// Redirect the dot file lookup at our temp file so stats() never touches $HOME.
	orig := getDotFilePath
	getDotFilePath = func() string { return dotFile }
	defer func() { getDotFilePath = orig }()

	out := captureStdout(t, func() { stats(email) })
	if out == "" {
		t.Error("stats produced no output")
	}
}

func TestGetBeginningOfDay(t *testing.T) {
	in := time.Date(2026, time.June, 16, 17, 48, 35, 123, time.UTC)
	got := getBeginningOfDay(in)
	want := time.Date(2026, time.June, 16, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("getBeginningOfDay(%v) = %v, want %v", in, got, want)
	}
	if got.Location() != in.Location() {
		t.Errorf("location not preserved: got %v, want %v", got.Location(), in.Location())
	}
}

func TestCountDaysSinceDate(t *testing.T) {
	today := getBeginningOfDay(time.Now())

	if got := countDaysSinceDate(today); got != 0 {
		t.Errorf("countDaysSinceDate(today) = %d, want 0", got)
	}

	yesterday := today.Add(-24 * time.Hour)
	if got := countDaysSinceDate(yesterday); got != 1 {
		t.Errorf("countDaysSinceDate(yesterday) = %d, want 1", got)
	}

	old := today.Add(-200 * 24 * time.Hour)
	if got := countDaysSinceDate(old); got != outOfRange {
		t.Errorf("countDaysSinceDate(200 days ago) = %d, want %d", got, outOfRange)
	}
}

func TestCalcOffset(t *testing.T) {
	got := calcOffset()
	// Offset maps weekday -> remaining days in the week; always in [1,7].
	if got < 1 || got > 7 {
		t.Errorf("calcOffset() = %d, want value in [1,7]", got)
	}
	want := map[time.Weekday]int{
		time.Sunday: 7, time.Monday: 6, time.Tuesday: 5, time.Wednesday: 4,
		time.Thursday: 3, time.Friday: 2, time.Saturday: 1,
	}[time.Now().Weekday()]
	if got != want {
		t.Errorf("calcOffset() = %d, want %d for %v", got, want, time.Now().Weekday())
	}
}

func TestSortMapIntoSlice(t *testing.T) {
	m := map[int]int{3: 1, 1: 1, 2: 1, 0: 1}
	got := sortMapIntoSlice(m)
	want := []int{0, 1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sortMapIntoSlice = %v, want %v", got, want)
	}
}

func TestBuildCols(t *testing.T) {
	// keys 0..6 form a full week (week 0); the column is committed at dayInWeek==6.
	commits := map[int]int{0: 5, 1: 0, 2: 1, 3: 0, 4: 2, 5: 0, 6: 3}
	keys := []int{0, 1, 2, 3, 4, 5, 6}
	cols := buildCols(keys, commits)

	col, ok := cols[0]
	if !ok {
		t.Fatalf("buildCols missing week 0, got %v", cols)
	}
	want := column{5, 0, 1, 0, 2, 0, 3}
	if !reflect.DeepEqual(col, want) {
		t.Errorf("buildCols week 0 = %v, want %v", col, want)
	}
}

// captureStdout redirects os.Stdout for the duration of fn and returns what was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = orig
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	return string(out)
}

func TestPrintCellAndDayCol(t *testing.T) {
	// Each commit-count bucket produces non-empty output containing the value.
	for _, val := range []int{0, 3, 7, 12} {
		if out := captureStdout(t, func() { printCell(val, false) }); out == "" {
			t.Errorf("printCell(%d) produced no output", val)
		}
	}
	if out := captureStdout(t, func() { printCell(2, true) }); out == "" {
		t.Error("printCell(today) produced no output")
	}

	if out := captureStdout(t, func() { printDayCol(1) }); out != " Mon " {
		t.Errorf("printDayCol(1) = %q, want %q", out, " Mon ")
	}
	if out := captureStdout(t, func() { printDayCol(0) }); out != "     " {
		t.Errorf("printDayCol(0) = %q, want blank", out)
	}
}

func TestPrintCommitStats(t *testing.T) {
	// Exercises sortMapIntoSlice -> buildCols -> printCells (and the print helpers).
	commits := make(map[int]int)
	for i := 0; i <= daysInLastSixMonths; i++ {
		commits[i] = i % 11
	}
	out := captureStdout(t, func() { printCommitStats(commits) })
	if out == "" {
		t.Error("printCommitStats produced no output")
	}
}
