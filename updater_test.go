package tuikit_test

import (
	"testing"

	tuikit "github.com/moneycaringcoder/tuikit-go"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input   string
		major   int
		minor   int
		patch   int
		wantErr bool
	}{
		{"v1.2.3", 1, 2, 3, false},
		{"0.4.0", 0, 4, 0, false},
		{"v0.10.1", 0, 10, 1, false},
		{"v1.0.0", 1, 0, 0, false},
		{"bad", 0, 0, 0, true},
		{"v1.2", 0, 0, 0, true},
		{"", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := tuikit.ParseVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v.Major != tt.major || v.Minor != tt.minor || v.Patch != tt.patch {
				t.Errorf("got %d.%d.%d, want %d.%d.%d", v.Major, v.Minor, v.Patch, tt.major, tt.minor, tt.patch)
			}
		})
	}
}

func TestVersionNewerThan(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"v0.4.0", "v0.3.0", true},
		{"v0.3.0", "v0.4.0", false},
		{"v0.4.0", "v0.4.0", false},
		{"v1.0.0", "v0.99.0", true},
		{"v0.4.1", "v0.4.0", true},
		{"v2.0.0", "v1.9.9", true},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			a, _ := tuikit.ParseVersion(tt.a)
			b, _ := tuikit.ParseVersion(tt.b)
			if got := a.NewerThan(b); got != tt.want {
				t.Errorf("(%s).NewerThan(%s) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
