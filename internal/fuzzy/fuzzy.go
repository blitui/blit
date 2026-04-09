// Package fuzzy implements an fzf-inspired fuzzy scorer for command pickers.
//
// Scoring strategy:
//   - Sequential letter match (prefix run): highest bonus
//   - Word boundary match (after space/hyphen/underscore/slash/dot): medium bonus
//   - Consecutive matched characters: per-run bonus
//   - Smart case: case-insensitive unless pattern contains an uppercase letter
package fuzzy

import "strings"

const (
	scoreBase        = 1
	scoreConsecutive = 3
	scoreWordBound   = 6
	scorePrefix      = 10
)

// Match holds the result of a single fuzzy match operation.
type Match struct {
	// Score is the relevance score; higher is better. 0 means no match.
	Score float64

	// Positions are the indices in the target string that were matched.
	Positions []int
}

// Score computes how well pattern matches target using fzf-style heuristics.
// Returns a Match with Score == 0 if pattern is not a subsequence of target.
func Score(pattern, target string) Match {
	if pattern == "" {
		return Match{Score: 1}
	}

	// Smart case: if pattern is all lowercase, match case-insensitively.
	caseSensitive := strings.ToLower(pattern) != pattern
	cmpTarget := target
	cmpPattern := pattern
	if !caseSensitive {
		cmpTarget = strings.ToLower(target)
		cmpPattern = strings.ToLower(pattern)
	}

	runes := []rune(cmpTarget)
	patRunes := []rune(cmpPattern)
	origRunes := []rune(target)

	positions := make([]int, 0, len(patRunes))

	pi := 0
	for ti := 0; ti < len(runes) && pi < len(patRunes); ti++ {
		if runes[ti] == patRunes[pi] {
			positions = append(positions, ti)
			pi++
		}
	}

	if pi < len(patRunes) {
		return Match{}
	}

	score := 0
	prev := -2

	for i, pos := range positions {
		ch := origRunes[pos]
		s := scoreBase

		if pos == 0 {
			s += scorePrefix
		} else {
			if isSep(origRunes[pos-1]) {
				s += scoreWordBound
			}
			if isLower(origRunes[pos-1]) && isUpper(ch) {
				s += scoreWordBound
			}
		}

		if pos == prev+1 {
			s += scoreConsecutive
			if i >= 2 && positions[i-2]+1 == prev {
				s += scoreConsecutive
			}
		}

		score += s
		prev = pos
	}

	maxScore := len(patRunes) * (scoreBase + scorePrefix + scoreWordBound + scoreConsecutive*2)
	if maxScore == 0 {
		maxScore = 1
	}
	norm := float64(score) / float64(maxScore)
	if norm > 1 {
		norm = 1
	}

	return Match{
		Score:     norm,
		Positions: positions,
	}
}

func isSep(r rune) bool {
	return r == ' ' || r == '-' || r == '_' || r == '/' || r == '.' || r == ':'
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isLower(r rune) bool { return r >= 'a' && r <= 'z' }
