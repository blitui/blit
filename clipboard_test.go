package blit

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestCopyToClipboardCmd(t *testing.T) {
	cmd := CopyToClipboardCmd("hello")
	if cmd == nil {
		t.Fatal("CopyToClipboardCmd returned nil")
	}
}

func TestOsc52Sequence(t *testing.T) {
	seq := osc52Sequence("hello")
	encoded := base64.StdEncoding.EncodeToString([]byte("hello"))
	if !strings.Contains(seq, encoded) {
		t.Errorf("sequence missing base64 payload: %q", seq)
	}
	if !strings.HasPrefix(seq, "\033]52;c;") {
		t.Errorf("sequence missing OSC 52 prefix: %q", seq)
	}
	if !strings.HasSuffix(seq, "\a") {
		t.Errorf("sequence missing BEL terminator: %q", seq)
	}
}

func TestOsc52Sequence_Empty(t *testing.T) {
	seq := osc52Sequence("")
	encoded := base64.StdEncoding.EncodeToString([]byte(""))
	if !strings.Contains(seq, encoded) {
		t.Errorf("empty string sequence wrong: %q", seq)
	}
}

func TestCopyToClipboardWithNotify(t *testing.T) {
	cmd := CopyToClipboardWithNotify("test")
	if cmd == nil {
		t.Fatal("CopyToClipboardWithNotify returned nil")
	}
}

