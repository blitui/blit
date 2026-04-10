package blit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- TextField ---

func TestTextField_Label(t *testing.T) {
	f := NewTextField("name", "Full Name")
	if f.Label() != "Full Name" {
		t.Errorf("expected 'Full Name', got %q", f.Label())
	}
}

func TestTextField_WithHint(t *testing.T) {
	f := NewTextField("x", "X").WithHint("some hint")
	if f.hint != "some hint" {
		t.Errorf("expected hint 'some hint', got %q", f.hint)
	}
}

func TestTextField_WithPlaceholder(t *testing.T) {
	f := NewTextField("x", "X").WithPlaceholder("enter value")
	if f.input.Placeholder != "enter value" {
		t.Errorf("expected placeholder 'enter value', got %q", f.input.Placeholder)
	}
}

// --- PasswordField ---

func TestPasswordField_WithHint(t *testing.T) {
	f := NewPasswordField("pw", "Password").WithHint("min 8 chars")
	if f.hint != "min 8 chars" {
		t.Errorf("expected hint, got %q", f.hint)
	}
}

func TestPasswordField_WithPlaceholder(t *testing.T) {
	f := NewPasswordField("pw", "Password").WithPlaceholder("****")
	if f.input.Placeholder != "****" {
		t.Errorf("expected placeholder '****', got %q", f.input.Placeholder)
	}
}

func TestPasswordField_WithRequired(t *testing.T) {
	f := NewPasswordField("pw", "Password").WithRequired()
	if !f.required {
		t.Error("expected required=true")
	}
}

func TestPasswordField_WithValidator(t *testing.T) {
	f := NewPasswordField("pw", "Password").WithValidator(func(v string) error { return nil })
	if f.validator == nil {
		t.Error("expected validator to be set")
	}
}

func TestPasswordField_ValueAndSetValue(t *testing.T) {
	f := NewPasswordField("pw", "Password")
	f.SetValue("secret")
	if f.Value() != "secret" {
		t.Errorf("expected 'secret', got %q", f.Value())
	}
}

func TestPasswordField_Validate(t *testing.T) {
	f := NewPasswordField("pw", "Password").WithRequired()
	err := f.Validate()
	if err == nil {
		t.Error("expected validation error for empty required field")
	}
	f.SetValue("something")
	err = f.Validate()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestPasswordField_Update(t *testing.T) {
	f := NewPasswordField("pw", "Password")
	f.SetFocused(true)
	cmd := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	_ = cmd
}

func TestPasswordField_SetFocused(t *testing.T) {
	f := NewPasswordField("pw", "Password")
	f.SetFocused(true)
	f.SetFocused(false)
}

// --- SelectField ---

func TestSelectField_WithHint(t *testing.T) {
	f := NewSelectField("color", "Color", []string{"red", "blue"}).WithHint("pick one")
	if f.hint != "pick one" {
		t.Errorf("expected hint, got %q", f.hint)
	}
}

func TestSelectField_WithRequired(t *testing.T) {
	f := NewSelectField("color", "Color", []string{"red"}).WithRequired()
	if !f.required {
		t.Error("expected required=true")
	}
}

func TestSelectField_WithValidator(t *testing.T) {
	f := NewSelectField("color", "Color", []string{"red"}).WithValidator(func(v string) error { return nil })
	if f.validator == nil {
		t.Error("expected validator set")
	}
}

func TestSelectField_SetValue(t *testing.T) {
	f := NewSelectField("color", "Color", []string{"red", "blue", "green"})
	f.SetValue("blue")
	if f.Value() != "blue" {
		t.Errorf("expected 'blue', got %q", f.Value())
	}
	// Non-existent value should not change
	f.SetValue("purple")
	if f.Value() != "blue" {
		t.Errorf("expected still 'blue', got %q", f.Value())
	}
}

func TestSelectField_Validate(t *testing.T) {
	f := NewSelectField("color", "Color", []string{"red", "blue"})
	err := f.Validate()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSelectField_Update(t *testing.T) {
	f := NewSelectField("color", "Color", []string{"red", "blue", "green"})
	f.SetFocused(true)

	// Move right
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if f.Value() != "blue" {
		t.Errorf("expected 'blue' after right, got %q", f.Value())
	}

	// Move left
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if f.Value() != "red" {
		t.Errorf("expected 'red' after left, got %q", f.Value())
	}
}

// --- MultiSelectField ---

func TestMultiSelectField_WithHint(t *testing.T) {
	f := NewMultiSelectField("tags", "Tags", []string{"a", "b"}).WithHint("choose")
	if f.hint != "choose" {
		t.Errorf("expected hint, got %q", f.hint)
	}
}

func TestMultiSelectField_WithRequired(t *testing.T) {
	f := NewMultiSelectField("tags", "Tags", []string{"a"}).WithRequired()
	if !f.required {
		t.Error("expected required=true")
	}
}

func TestMultiSelectField_Validate(t *testing.T) {
	f := NewMultiSelectField("tags", "Tags", []string{"a", "b"}).WithRequired()
	err := f.Validate()
	if err == nil {
		t.Error("expected error when nothing selected and required")
	}
	// Select one
	f.selected[0] = true
	err = f.Validate()
	if err != nil {
		t.Errorf("expected no error after selection, got %v", err)
	}
}

func TestMultiSelectField_SetValue(t *testing.T) {
	f := NewMultiSelectField("tags", "Tags", []string{"a", "b", "c"})
	f.SetValue("a,c")
	if f.Value() != "a,c" {
		t.Errorf("expected 'a,c', got %q", f.Value())
	}
}

func TestMultiSelectField_Update(t *testing.T) {
	f := NewMultiSelectField("tags", "Tags", []string{"a", "b", "c"})
	f.SetFocused(true)

	// Toggle space
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !f.selected[0] {
		t.Error("expected first option selected after space")
	}

	// Move right
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !f.selected[1] {
		t.Error("expected second option selected")
	}
}

// --- ConfirmField ---

func TestConfirmField_WithHint(t *testing.T) {
	f := NewConfirmField("ok", "Confirm").WithHint("are you sure?")
	if f.hint != "are you sure?" {
		t.Errorf("expected hint, got %q", f.hint)
	}
}

func TestConfirmField_SetValue(t *testing.T) {
	f := NewConfirmField("ok", "Confirm")
	f.SetValue("true")
	if f.Value() != "true" {
		t.Errorf("expected 'true', got %q", f.Value())
	}
	f.SetValue("no")
	if f.Value() != "false" {
		t.Errorf("expected 'false', got %q", f.Value())
	}
}

func TestConfirmField_Validate(t *testing.T) {
	f := NewConfirmField("ok", "Confirm")
	if err := f.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestConfirmField_SetFocused(t *testing.T) {
	f := NewConfirmField("ok", "Confirm")
	f.SetFocused(true)
	f.SetFocused(false)
}

func TestConfirmField_Update(t *testing.T) {
	f := NewConfirmField("ok", "Confirm")
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if f.Value() != "true" {
		t.Errorf("expected true after 'y', got %q", f.Value())
	}
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if f.Value() != "false" {
		t.Errorf("expected false after 'n', got %q", f.Value())
	}
	// Toggle with space
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if f.Value() != "true" {
		t.Errorf("expected toggle to true after space, got %q", f.Value())
	}
}

// --- NumberField ---

func TestNumberField_WithHint(t *testing.T) {
	f := NewNumberField("age", "Age").WithHint("years")
	if f.hint != "years" {
		t.Errorf("expected hint, got %q", f.hint)
	}
}

func TestNumberField_WithPlaceholder(t *testing.T) {
	f := NewNumberField("age", "Age").WithPlaceholder("0")
	if f.input.Placeholder != "0" {
		t.Errorf("expected placeholder, got %q", f.input.Placeholder)
	}
}

func TestNumberField_WithRequired(t *testing.T) {
	f := NewNumberField("age", "Age").WithRequired()
	if !f.required {
		t.Error("expected required=true")
	}
}

func TestNumberField_WithDefault(t *testing.T) {
	f := NewNumberField("age", "Age").WithDefault(25)
	if f.Value() != "25" {
		t.Errorf("expected '25', got %q", f.Value())
	}
}

func TestNumberField_Value(t *testing.T) {
	f := NewNumberField("n", "N")
	f.SetValue("42")
	if f.Value() != "42" {
		t.Errorf("expected '42', got %q", f.Value())
	}
}

func TestNumberField_Update(t *testing.T) {
	f := NewNumberField("n", "N")
	f.SetFocused(true)
	cmd := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	_ = cmd
	f.SetFocused(false)
}

// --- Form ---

func TestForm_Focused(t *testing.T) {
	f := NewForm(FormOpts{
		Groups: []FormGroup{{Title: "G", Fields: []Field{NewTextField("name", "Name")}}},
	})
	if f.Focused() {
		t.Error("form should not be focused initially")
	}
	f.SetFocused(true)
	if !f.Focused() {
		t.Error("form should be focused after SetFocused(true)")
	}
}

func TestForm_Reset(t *testing.T) {
	f := NewForm(FormOpts{
		Groups: []FormGroup{{Title: "G", Fields: []Field{NewTextField("name", "Name")}}},
	})
	f.SetTheme(DefaultTheme())
	f.SetSize(80, 24)
	f.SetFocused(true)
	f.submitted = true
	f.wizardStep = 2
	f.Reset()
	if f.submitted {
		t.Error("submitted should be false after Reset")
	}
	if f.wizardStep != 0 {
		t.Error("wizardStep should be 0 after Reset")
	}
}
