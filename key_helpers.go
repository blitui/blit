package blit

// KeyMsg helper functions for the most common key-checking patterns.
// These reduce the boilerplate of type-asserting Msg and comparing strings.

// KeyString returns the string representation of a KeyMsg, or "" if msg
// is not a KeyMsg.
func KeyString(msg Msg) string {
	if k, ok := msg.(KeyMsg); ok {
		return k.String()
	}
	return ""
}

// IsKey reports whether msg is a KeyMsg whose string representation
// equals key.
func IsKey(msg Msg, key string) bool {
	if k, ok := msg.(KeyMsg); ok {
		return k.String() == key
	}
	return false
}

// IsEnter reports whether msg is the Enter key.
func IsEnter(msg Msg) bool {
	if k, ok := msg.(KeyMsg); ok {
		return k.Type == KeyEnter
	}
	return false
}

// IsEscape reports whether msg is the Escape key.
func IsEscape(msg Msg) bool {
	if k, ok := msg.(KeyMsg); ok {
		return k.Type == KeyEscape
	}
	return false
}

// IsTab reports whether msg is the Tab key.
func IsTab(msg Msg) bool {
	if k, ok := msg.(KeyMsg); ok {
		return k.Type == KeyTab
	}
	return false
}

// IsBackspace reports whether msg is the Backspace key.
func IsBackspace(msg Msg) bool {
	if k, ok := msg.(KeyMsg); ok {
		return k.Type == KeyBackspace
	}
	return false
}

// IsRunes reports whether msg is a KeyRunes event (regular character input).
func IsRunes(msg Msg) bool {
	if k, ok := msg.(KeyMsg); ok {
		return k.Type == KeyRunes
	}
	return false
}

// Runes returns the typed rune characters if msg is a KeyRunes event,
// or nil otherwise.
func Runes(msg Msg) []rune {
	if k, ok := msg.(KeyMsg); ok {
		if k.Type == KeyRunes {
			return k.Runes
		}
	}
	return nil
}
