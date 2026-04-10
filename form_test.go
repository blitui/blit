package blit_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	blit "github.com/blitui/blit"
	"github.com/blitui/blit/btest"
)

type formModel struct {
	form     *blit.Form
	lastVals map[string]string
}

func newFormModel(opts blit.FormOpts) *formModel {
	m := &formModel{}
	extra := opts.OnSubmit
	opts.OnSubmit = func(values map[string]string) {
		m.lastVals = values
		if extra != nil {
			extra(values)
		}
	}
	m.form = blit.NewForm(opts)
	m.form.SetTheme(blit.DefaultTheme())
	m.form.SetFocused(true)
	return m
}

func (m *formModel) Init() tea.Cmd { return m.form.Init() }

func (m *formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.form.SetSize(ws.Width, ws.Height)
		return m, nil
	}
	comp, cmd := m.form.Update(msg, blit.Context{})
	m.form = comp.(*blit.Form)
	return m, cmd
}

func (m *formModel) View() string { return m.form.View() }

func TestFormValidators(t *testing.T) {
	if err := blit.Required()("hi"); err != nil {
		t.Errorf("Required: unexpected error: %v", err)
	}
	if err := blit.Required()(""); err == nil {
		t.Error("Required: expected error for empty")
	}
	if err := blit.Required()("  "); err == nil {
		t.Error("Required: expected error for spaces")
	}
	if err := blit.MinLength(3)("abc"); err != nil {
		t.Errorf("MinLength: unexpected error: %v", err)
	}
	if err := blit.MinLength(5)("ab"); err == nil {
		t.Error("MinLength: expected error")
	}
	if err := blit.MaxLength(10)("hello"); err != nil {
		t.Errorf("MaxLength: unexpected error: %v", err)
	}
	if err := blit.MaxLength(3)("toolong"); err == nil {
		t.Error("MaxLength: expected error")
	}
	if err := blit.EmailValidator()("x@y.z"); err != nil {
		t.Errorf("Email: unexpected error: %v", err)
	}
	if err := blit.EmailValidator()("bad"); err == nil {
		t.Error("Email: expected error")
	}
	if err := blit.URLValidator()("https://x.com"); err != nil {
		t.Errorf("URL: unexpected error: %v", err)
	}
	if err := blit.URLValidator()("ftp://x.com"); err == nil {
		t.Error("URL: expected error for ftp")
	}
}

func TestFormComposeValidators(t *testing.T) {
	v := blit.ComposeValidators(blit.Required(), blit.MinLength(8))
	if v("") == nil {
		t.Error("expected required error")
	}
	if v("short") == nil {
		t.Error("expected min-length error")
	}
	if v("longpass") != nil {
		t.Error("unexpected error for valid value")
	}
}

func TestFormRegexValidator(t *testing.T) {
	v := blit.RegexValidator(`^\d{4}$`, "must be 4 digits")
	if v("1234") != nil {
		t.Error("unexpected error for 1234")
	}
	if v("abc") == nil {
		t.Error("expected error for abc")
	}
}

func TestFormNavigation(t *testing.T) {
	opts := blit.FormOpts{
		Groups: []blit.FormGroup{{Fields: []blit.Field{
			blit.NewTextField("a", "FieldA"),
			blit.NewTextField("b", "FieldB"),
		}}},
	}
	tm := btest.NewTestModel(t, newFormModel(opts), 80, 24)
	if !tm.Screen().Contains("FieldA") {
		t.Error("expected FieldA on screen")
	}
	tm.SendKey("tab")
	if !tm.Screen().Contains("FieldB") {
		t.Error("expected FieldB after tab")
	}
	tm.SendKey("shift+tab")
	if !tm.Screen().Contains("FieldA") {
		t.Error("expected FieldA after shift+tab")
	}
}

func TestFormSubmit(t *testing.T) {
	var captured map[string]string
	opts := blit.FormOpts{
		Groups: []blit.FormGroup{{Fields: []blit.Field{
			blit.NewTextField("name", "Name").WithDefault("Alice"),
		}}},
		OnSubmit: func(v map[string]string) { captured = v },
	}
	tm := btest.NewTestModel(t, newFormModel(opts), 80, 24)
	tm.SendKey("enter")
	if captured == nil {
		t.Fatal("OnSubmit not called")
	}
	if captured["name"] != "Alice" {
		t.Errorf("got %q, want Alice", captured["name"])
	}
}

func TestFormInlineValidation(t *testing.T) {
	var submitted bool
	opts := blit.FormOpts{
		Groups: []blit.FormGroup{{Fields: []blit.Field{
			blit.NewTextField("email", "Email").
				WithRequired().
				WithValidator(blit.EmailValidator()),
		}}},
		OnSubmit: func(_ map[string]string) { submitted = true },
	}
	tm := btest.NewTestModel(t, newFormModel(opts), 80, 24)
	tm.SendKey("enter")
	if submitted {
		t.Error("should not submit empty required field")
	}
	tm.Type("bad")
	tm.SendKey("enter")
	if submitted {
		t.Error("should not submit invalid email")
	}
	if !tm.Screen().Contains("valid email") {
		t.Errorf("expected validation error, got:\n%s", tm.Screen().String())
	}
}

func TestFormWizardMode(t *testing.T) {
	var submitted bool
	opts := blit.FormOpts{
		WizardMode: true,
		Groups: []blit.FormGroup{{Fields: []blit.Field{
			blit.NewTextField("s1", "Step 1"),
			blit.NewTextField("s2", "Step 2"),
			blit.NewTextField("s3", "Step 3"),
		}}},
		OnSubmit: func(_ map[string]string) { submitted = true },
	}
	tm := btest.NewTestModel(t, newFormModel(opts), 80, 24)
	if !tm.Screen().Contains("Step 1 of 3") {
		t.Errorf("expected Step 1 of 3, got:\n%s", tm.Screen().String())
	}
	tm.SendKey("enter")
	if !tm.Screen().Contains("Step 2 of 3") {
		t.Errorf("expected Step 2 of 3, got:\n%s", tm.Screen().String())
	}
	tm.SendKey("shift+tab")
	if !tm.Screen().Contains("Step 1 of 3") {
		t.Errorf("expected back to step 1, got:\n%s", tm.Screen().String())
	}
	tm.SendKey("enter")
	tm.SendKey("enter")
	tm.SendKey("enter")
	if !submitted {
		t.Error("expected form submitted")
	}
}

func TestFormFieldTypes(t *testing.T) {
	opts := blit.FormOpts{
		Groups: []blit.FormGroup{{
			Title: "Account",
			Fields: []blit.Field{
				blit.NewTextField("u", "Username"),
				blit.NewPasswordField("p", "Password"),
				blit.NewSelectField("r", "Role", []string{"Admin", "User"}),
				blit.NewMultiSelectField("x", "Permissions", []string{"Read", "Write"}),
				blit.NewConfirmField("c", "Confirm"),
				blit.NewNumberField("n", "Age").WithMin(18).WithMax(120),
			},
		}},
	}
	tm := btest.NewTestModel(t, newFormModel(opts), 80, 40)
	scr := tm.Screen()
	for _, lbl := range []string{"Username", "Password", "Role", "Permissions", "Confirm", "Age", "Account"} {
		if !scr.Contains(lbl) {
			t.Errorf("missing label %q on screen", lbl)
		}
	}
}

func TestFormNumberField(t *testing.T) {
	type tc struct {
		val     string
		wantErr bool
	}
	cases := []tc{
		{"42", false}, {"25.5", false}, {"abc", true},
		{"5", true}, {"200", true}, {"50", false},
	}
	f := blit.NewNumberField("n", "N").WithMin(18).WithMax(120)
	for _, c := range cases {
		t.Run(c.val, func(t *testing.T) {
			f.SetValue(c.val)
			err := f.Validate()
			if c.wantErr && err == nil {
				t.Errorf("expected error for %q", c.val)
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error for %q: %v", c.val, err)
			}
		})
	}
}

func TestFormSelectField(t *testing.T) {
	f := blit.NewSelectField("l", "L", []string{"Go", "Rust", "Python"})
	if f.Value() != "Go" {
		t.Errorf("want Go, got %q", f.Value())
	}
	f.WithDefault("Rust")
	if f.Value() != "Rust" {
		t.Errorf("want Rust, got %q", f.Value())
	}
}

func TestFormMultiSelectField(t *testing.T) {
	f := blit.NewMultiSelectField("t", "T", []string{"Go", "TUI", "CLI"})
	if f.Value() != "" {
		t.Errorf("expected empty, got %q", f.Value())
	}
	f.SetValue("Go,CLI")
	v := f.Value()
	if !strings.Contains(v, "Go") || !strings.Contains(v, "CLI") {
		t.Errorf("expected Go and CLI in %q", v)
	}
}

func TestFormConfirmField(t *testing.T) {
	f := blit.NewConfirmField("c", "C")
	if f.Value() != "false" {
		t.Errorf("want false, got %q", f.Value())
	}
	f.WithDefault(true)
	if f.Value() != "true" {
		t.Errorf("want true, got %q", f.Value())
	}
}

func TestFormGroupSeparator(t *testing.T) {
	opts := blit.FormOpts{
		Groups: []blit.FormGroup{
			{Title: "Personal", Fields: []blit.Field{blit.NewTextField("n", "Name")}},
			{Title: "Account", Fields: []blit.Field{blit.NewTextField("e", "Email")}},
		},
	}
	tm := btest.NewTestModel(t, newFormModel(opts), 80, 24)
	scr := tm.Screen()
	if !scr.Contains("Personal") {
		t.Error("expected Personal group title")
	}
	if !scr.Contains("Account") {
		t.Error("expected Account group title")
	}
}
