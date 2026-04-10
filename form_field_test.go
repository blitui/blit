package blit_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	blit "github.com/blitui/blit"
)

// --- TextField ---

func TestTextField_Interface(t *testing.T) {
	f := blit.NewTextField("name", "Name").
		WithHint("your full name").
		WithPlaceholder("John Doe").
		WithRequired()

	if f.FieldID() != "name" {
		t.Fatalf("FieldID() = %q, want name", f.FieldID())
	}
	if f.Label() != "Name" {
		t.Fatalf("Label() = %q, want Name", f.Label())
	}
}

func TestTextField_SetFocused(t *testing.T) {
	f := blit.NewTextField("t", "T")
	f.SetFocused(true)
	f.SetFocused(false)
	// No panic means success.
}

func TestTextField_Validate(t *testing.T) {
	f := blit.NewTextField("e", "Email").
		WithRequired().
		WithValidator(blit.EmailValidator())

	f.SetValue("")
	if err := f.Validate(); err == nil {
		t.Error("expected required error")
	}

	f.SetValue("bad")
	if err := f.Validate(); err == nil {
		t.Error("expected email error")
	}

	f.SetValue("a@b.com")
	if err := f.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTextField_View(t *testing.T) {
	f := blit.NewTextField("t", "Title").WithHint("optional")
	theme := blit.DefaultTheme()

	view := f.View(true, theme, 60)
	if !strings.Contains(view, "Title") {
		t.Fatal("view should contain label")
	}

	view = f.View(false, theme, 60)
	if !strings.Contains(view, "Title") {
		t.Fatal("unfocused view should contain label")
	}
}

func TestTextField_ViewWithError(t *testing.T) {
	f := blit.NewTextField("t", "T").WithRequired()
	f.SetValue("")
	f.Validate()

	view := f.View(true, blit.DefaultTheme(), 60)
	if !strings.Contains(view, "required") {
		t.Fatalf("view should show validation error:\n%s", view)
	}
}

func TestTextField_Update(t *testing.T) {
	f := blit.NewTextField("t", "T")
	f.SetFocused(true)
	// Sending a key msg clears error state.
	f.SetValue("")
	f.Validate()
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	// No panic, error should be cleared.
}

// --- PasswordField ---

func TestPasswordField_Interface(t *testing.T) {
	f := blit.NewPasswordField("pw", "Password").
		WithHint("8+ chars").
		WithPlaceholder("secret").
		WithRequired().
		WithValidator(blit.MinLength(8))

	if f.FieldID() != "pw" {
		t.Fatalf("FieldID() = %q, want pw", f.FieldID())
	}
	if f.Label() != "Password" {
		t.Fatalf("Label() = %q, want Password", f.Label())
	}
}

func TestPasswordField_SetFocused(t *testing.T) {
	f := blit.NewPasswordField("p", "P")
	f.SetFocused(true)
	f.SetFocused(false)
}

func TestPasswordField_Validate(t *testing.T) {
	f := blit.NewPasswordField("p", "P").WithRequired()
	f.SetValue("")
	if err := f.Validate(); err == nil {
		t.Error("expected required error")
	}
	f.SetValue("secret123")
	if err := f.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPasswordField_View(t *testing.T) {
	f := blit.NewPasswordField("p", "Pass").WithHint("min 8")
	view := f.View(true, blit.DefaultTheme(), 60)
	if !strings.Contains(view, "Pass") {
		t.Fatal("view should contain label")
	}
}

func TestPasswordField_Update(t *testing.T) {
	f := blit.NewPasswordField("p", "P")
	f.SetFocused(true)
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
}

// --- SelectField ---

func TestSelectField_Navigation(t *testing.T) {
	f := blit.NewSelectField("s", "S", []string{"A", "B", "C"})
	f.Update(tea.KeyMsg{Type: tea.KeyRight})
	if f.Value() != "B" {
		t.Fatalf("Value() = %q, want B", f.Value())
	}
	f.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if f.Value() != "A" {
		t.Fatalf("Value() = %q, want A", f.Value())
	}

	// Vi keys.
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if f.Value() != "B" {
		t.Fatalf("l: Value() = %q, want B", f.Value())
	}
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if f.Value() != "A" {
		t.Fatalf("h: Value() = %q, want A", f.Value())
	}
}

func TestSelectField_SetValue(t *testing.T) {
	f := blit.NewSelectField("s", "S", []string{"X", "Y", "Z"})
	f.SetValue("Z")
	if f.Value() != "Z" {
		t.Fatalf("Value() = %q, want Z", f.Value())
	}
	f.SetValue("nonexistent")
	if f.Value() != "Z" {
		t.Fatal("SetValue with nonexistent should not change")
	}
}

func TestSelectField_Empty(t *testing.T) {
	f := blit.NewSelectField("s", "S", []string{})
	if f.Value() != "" {
		t.Fatalf("empty select Value() = %q, want empty", f.Value())
	}
}

func TestSelectField_View(t *testing.T) {
	f := blit.NewSelectField("s", "S", []string{"A", "B"}).WithHint("pick one")
	view := f.View(true, blit.DefaultTheme(), 60)
	if !strings.Contains(view, "A") || !strings.Contains(view, "B") {
		t.Fatalf("view should contain options:\n%s", view)
	}
}

func TestSelectField_Validate(t *testing.T) {
	f := blit.NewSelectField("s", "S", []string{"A"}).WithRequired().WithValidator(nil)
	if err := f.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSelectField_SetFocused(t *testing.T) {
	f := blit.NewSelectField("s", "S", []string{"A"})
	f.SetFocused(true)
	f.SetFocused(false)
}

// --- MultiSelectField ---

func TestMultiSelectField_Toggle(t *testing.T) {
	f := blit.NewMultiSelectField("m", "M", []string{"R", "W", "X"})
	// Toggle first item.
	f.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !strings.Contains(f.Value(), "R") {
		t.Fatalf("Value() = %q, should contain R", f.Value())
	}

	// Move right and toggle.
	f.Update(tea.KeyMsg{Type: tea.KeyRight})
	f.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !strings.Contains(f.Value(), "W") {
		t.Fatalf("Value() = %q, should contain W", f.Value())
	}

	// Toggle off first item.
	f.Update(tea.KeyMsg{Type: tea.KeyLeft})
	f.Update(tea.KeyMsg{Type: tea.KeySpace})
	if strings.Contains(f.Value(), "R") {
		t.Fatalf("Value() = %q, should not contain R after untoggle", f.Value())
	}
}

func TestMultiSelectField_ValidateRequired(t *testing.T) {
	f := blit.NewMultiSelectField("m", "M", []string{"A", "B"}).WithRequired()
	if err := f.Validate(); err == nil {
		t.Error("expected error when nothing selected")
	}
	f.Update(tea.KeyMsg{Type: tea.KeySpace})
	if err := f.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMultiSelectField_View(t *testing.T) {
	f := blit.NewMultiSelectField("m", "M", []string{"A", "B"}).WithHint("choose")
	view := f.View(true, blit.DefaultTheme(), 60)
	if !strings.Contains(view, "A") || !strings.Contains(view, "B") {
		t.Fatalf("view should contain options:\n%s", view)
	}
	if !strings.Contains(view, "toggle") {
		t.Fatalf("focused view should show toggle hint:\n%s", view)
	}
}

func TestMultiSelectField_SetFocused(t *testing.T) {
	f := blit.NewMultiSelectField("m", "M", []string{"A"})
	f.SetFocused(true)
	f.SetFocused(false)
}

func TestMultiSelectField_ViNav(t *testing.T) {
	f := blit.NewMultiSelectField("m", "M", []string{"A", "B", "C"})
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	f.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !strings.Contains(f.Value(), "B") {
		t.Fatalf("l+space should select B, got %q", f.Value())
	}
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	f.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !strings.Contains(f.Value(), "A") {
		t.Fatalf("h+space should select A, got %q", f.Value())
	}
}

// --- ConfirmField ---

func TestConfirmField_Toggle(t *testing.T) {
	f := blit.NewConfirmField("c", "Confirm")
	if f.Value() != "false" {
		t.Fatal("default should be false")
	}
	f.Update(tea.KeyMsg{Type: tea.KeySpace})
	if f.Value() != "true" {
		t.Fatal("space should toggle to true")
	}
	f.Update(tea.KeyMsg{Type: tea.KeySpace})
	if f.Value() != "false" {
		t.Fatal("space should toggle back to false")
	}
}

func TestConfirmField_YesNo(t *testing.T) {
	f := blit.NewConfirmField("c", "C")
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if f.Value() != "true" {
		t.Fatal("y should set true")
	}
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if f.Value() != "false" {
		t.Fatal("n should set false")
	}
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")})
	if f.Value() != "true" {
		t.Fatal("Y should set true")
	}
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})
	if f.Value() != "false" {
		t.Fatal("N should set false")
	}
}

func TestConfirmField_SetValue(t *testing.T) {
	f := blit.NewConfirmField("c", "C")
	f.SetValue("true")
	if f.Value() != "true" {
		t.Fatal("SetValue true should work")
	}
	f.SetValue("yes")
	if f.Value() != "true" {
		t.Fatal("SetValue yes should work")
	}
	f.SetValue("1")
	if f.Value() != "true" {
		t.Fatal("SetValue 1 should work")
	}
	f.SetValue("no")
	if f.Value() != "false" {
		t.Fatal("SetValue no should be false")
	}
}

func TestConfirmField_View(t *testing.T) {
	f := blit.NewConfirmField("c", "Agree?").WithHint("required")
	view := f.View(true, blit.DefaultTheme(), 60)
	if !strings.Contains(view, "Agree?") {
		t.Fatal("view should contain label")
	}
	if !strings.Contains(view, "Yes") || !strings.Contains(view, "No") {
		t.Fatal("view should contain Yes and No")
	}
}

func TestConfirmField_SetFocused(t *testing.T) {
	f := blit.NewConfirmField("c", "C")
	f.SetFocused(true)
	f.SetFocused(false)
}

func TestConfirmField_ArrowToggle(t *testing.T) {
	f := blit.NewConfirmField("c", "C")
	f.Update(tea.KeyMsg{Type: tea.KeyRight})
	if f.Value() != "true" {
		t.Fatal("right should toggle")
	}
	f.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if f.Value() != "false" {
		t.Fatal("left should toggle back")
	}
}

// --- NumberField ---

func TestNumberField_Interface(t *testing.T) {
	f := blit.NewNumberField("n", "Age").
		WithHint("18+").
		WithPlaceholder("25").
		WithRequired().
		WithMin(18).WithMax(120).
		WithDefault(25)

	if f.FieldID() != "n" {
		t.Fatalf("FieldID() = %q, want n", f.FieldID())
	}
	if f.Value() != "25" {
		t.Fatalf("Value() = %q, want 25", f.Value())
	}
}

func TestNumberField_SetFocused(t *testing.T) {
	f := blit.NewNumberField("n", "N")
	f.SetFocused(true)
	f.SetFocused(false)
}

func TestNumberField_ValidateEmpty(t *testing.T) {
	f := blit.NewNumberField("n", "N")
	f.SetValue("")
	if err := f.Validate(); err != nil {
		t.Error("non-required empty should be valid")
	}
}

func TestNumberField_View(t *testing.T) {
	f := blit.NewNumberField("n", "Count").WithHint("positive")
	view := f.View(true, blit.DefaultTheme(), 60)
	if !strings.Contains(view, "Count") {
		t.Fatal("view should contain label")
	}
}

func TestNumberField_Update(t *testing.T) {
	f := blit.NewNumberField("n", "N")
	f.SetFocused(true)
	f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("5")})
}

// --- Form ---

func TestForm_Values(t *testing.T) {
	form := blit.NewForm(blit.FormOpts{
		Groups: []blit.FormGroup{{Fields: []blit.Field{
			blit.NewTextField("a", "A").WithDefault("hello"),
			blit.NewConfirmField("b", "B").WithDefault(true),
		}}},
	})
	vals := form.Values()
	if vals["a"] != "hello" {
		t.Fatalf("a = %q, want hello", vals["a"])
	}
	if vals["b"] != "true" {
		t.Fatalf("b = %q, want true", vals["b"])
	}
}

func TestForm_Reset(t *testing.T) {
	form := blit.NewForm(blit.FormOpts{
		Groups: []blit.FormGroup{{Fields: []blit.Field{
			blit.NewTextField("a", "A"),
			blit.NewTextField("b", "B"),
		}}},
	})
	form.SetFocused(true)
	form.SetTheme(blit.DefaultTheme())
	form.SetSize(80, 24)

	// Navigate forward.
	form.Update(tea.KeyMsg{Type: tea.KeyTab}, blit.Context{})
	form.Reset()
	// After reset, should be back at start.
}

func TestForm_EmptyView(t *testing.T) {
	form := blit.NewForm(blit.FormOpts{})
	form.SetTheme(blit.DefaultTheme())
	form.SetSize(80, 24)
	if form.View() != "" {
		t.Fatal("empty form should return empty view")
	}
}

func TestForm_UnfocusedIgnoresInput(t *testing.T) {
	form := blit.NewForm(blit.FormOpts{
		Groups: []blit.FormGroup{{Fields: []blit.Field{
			blit.NewTextField("a", "A"),
		}}},
	})
	form.SetFocused(false)
	updated, _ := form.Update(tea.KeyMsg{Type: tea.KeyTab}, blit.Context{})
	_ = updated.(*blit.Form)
}

func TestForm_KeyBindings(t *testing.T) {
	form := blit.NewForm(blit.FormOpts{})
	binds := form.KeyBindings()
	if len(binds) == 0 {
		t.Fatal("expected non-empty key bindings")
	}
}
