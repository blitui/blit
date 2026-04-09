package fuzzy_test

import (
	"testing"

	"github.com/moneycaringcoder/tuikit-go/internal/fuzzy"
)

func TestScore_EmptyPattern(t *testing.T) {
	m := fuzzy.Score("", "anything")
	if m.Score != 1 {
		t.Errorf("empty pattern score = %v, want 1", m.Score)
	}
}

func TestScore_NoMatch(t *testing.T) {
	m := fuzzy.Score("xyz", "abcdef")
	if m.Score != 0 {
		t.Errorf("non-matching pattern score = %v, want 0", m.Score)
	}
}

func TestScore_ExactMatch(t *testing.T) {
	m := fuzzy.Score("abc", "abc")
	if m.Score == 0 {
		t.Error("exact match should have positive score")
	}
}

func TestScore_PrefixHigherThanMiddle(t *testing.T) {
	prefix := fuzzy.Score("go", "golang")
	middle := fuzzy.Score("go", "ergo")
	if prefix.Score <= middle.Score {
		t.Errorf("prefix score (%v) should be > middle score (%v)", prefix.Score, middle.Score)
	}
}

func TestScore_ConsecutiveBonus(t *testing.T) {
	consec := fuzzy.Score("abc", "abcdef")
	spread := fuzzy.Score("abc", "axbycz")
	if consec.Score <= spread.Score {
		t.Errorf("consecutive score (%v) should be > spread score (%v)", consec.Score, spread.Score)
	}
}

func TestScore_WordBoundaryBonus(t *testing.T) {
	bound := fuzzy.Score("gb", "go-build")
	middle := fuzzy.Score("gb", "rgb")
	if bound.Score <= middle.Score {
		t.Errorf("word boundary score (%v) should be > middle score (%v)", bound.Score, middle.Score)
	}
}

func TestScore_CamelCaseBoundary(t *testing.T) {
	camel := fuzzy.Score("gb", "GetBlob")
	plain := fuzzy.Score("gb", "rgb")
	if camel.Score <= plain.Score {
		t.Errorf("camelCase boundary score (%v) should be > plain (%v)", camel.Score, plain.Score)
	}
}

func TestScore_SmartCase_Insensitive(t *testing.T) {
	lower := fuzzy.Score("go", "Golang")
	if lower.Score == 0 {
		t.Error("lowercase pattern should match case-insensitively")
	}
}

func TestScore_SmartCase_Sensitive(t *testing.T) {
	upper := fuzzy.Score("Go", "golang")
	if upper.Score != 0 {
		t.Errorf("uppercase pattern should match case-sensitively, got score %v", upper.Score)
	}
}

func TestScore_SmartCase_SensitiveMatch(t *testing.T) {
	m := fuzzy.Score("Go", "Golang")
	if m.Score == 0 {
		t.Error("uppercase pattern Go should match Golang")
	}
}

func TestScore_Positions(t *testing.T) {
	m := fuzzy.Score("ac", "abcd")
	if len(m.Positions) != 2 {
		t.Fatalf("positions len = %d, want 2", len(m.Positions))
	}
	if m.Positions[0] != 0 || m.Positions[1] != 2 {
		t.Errorf("positions = %v, want [0 2]", m.Positions)
	}
}

func TestScore_ScoreRange(t *testing.T) {
	cases := []struct{ pat, target string }{
		{"f", "file.go"},
		{"go", "golang"},
		{"git", "git-commit"},
		{"ts", "TypeScript"},
		{"mk", "make_build"},
		{"ab", "a-b-c"},
		{"abc", "a_b_c"},
		{"run", "runTests"},
	}
	for _, c := range cases {
		m := fuzzy.Score(c.pat, c.target)
		if m.Score < 0 || m.Score > 1 {
			t.Errorf("Score(%q, %q) = %v, out of [0,1]", c.pat, c.target, m.Score)
		}
	}
}

func TestScore_Ranking(t *testing.T) {
	a := fuzzy.Score("git", "git-commit")
	b := fuzzy.Score("git", "greet")
	if a.Score <= b.Score {
		t.Errorf("git vs git-commit(%v) should outscore greet(%v)", a.Score, b.Score)
	}
}
