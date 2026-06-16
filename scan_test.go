package main

import (
	"os/user"
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
