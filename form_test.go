package tuikit_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	tuikit "github.com/moneycaringcoder/tuikit-go"
	"github.com/moneycaringcoder/tuikit-go/tuitest"
)

type formModel struct {
	form     *tuikit.Form
	lastVals map[string]string
}

func newFormModel(opts tuikit.FormOpts) *formModel {
	m := &formModel{}
	extra := opts.OnSubmit
	opts.OnSubmit = func(values map[string]string) {
		m.lastVals = values
		if extra != nil {
			extra(values)
		}
	}
	m.form = tuikit.NewForm(opts)
	m.form.SetTheme(tuikit.DefaultTheme())
	m.form.SetFocused(true)
	return m
}

func (m *formModel) Init() tea.Cmd { return m.form.Init() }

func (m *formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.form.SetSize(ws.Width, ws.Height)
		return m, nil
	}
	comp, cmd := m.form.Update(msg, tuikit.Context{})
	m.form = comp.(*tuikit.Form)
	return m, cmd
}

func (m *formModel) View() string { return m.form.View() }

func TestFormValidators(t *testing.T) {
	if err := tuikit.Required()("hi"); err != nil {
		t.Errorf("Required: unexpected error: %v", err)
	}
	if err := tuikit.Required()(""); err == nil {
		t.Error("Required: expected error for empty")
	}
	if err := tuikit.Required()("  "); err == nil {
		t.Error("Required: expected error for spaces")
	}
	if err := tuikit.MinLength(3)("abc"); err != nil {
		t.Errorf("MinLength: unexpected error: %v", err)
	}
	if err := tuikit.MinLength(5)("ab"); err == nil {
		t.Error("MinLength: expected error")
	}
	if err := tuikit.MaxLength(10)("hello"); err != nil {
		t.Errorf("MaxLength: unexpected error: %v", err)
	}
	if err := tuikit.MaxLength(3)("toolong"); err == nil {
		t.Error("MaxLength: expected error")
	}
	if err := tuikit.EmailValidator()("x@y.z"); err != nil {
		t.Errorf("Email: unexpected error: %v", err)
	}
	if err := tuikit.EmailValidator()("bad"); err == nil {
		t.Error("Email: expected error")
	}
	if err := tuikit.URLValidator()("https://x.com"); err != nil {
		t.Errorf("URL: unexpected error: %v", err)
	}
	if err := tuikit.URLValidator()("ftp://x.com"); err == nil {
		t.Error("URL: expected error for ftp")
	}
}

func TestFormComposeValidators(t *testing.T) {
	v := tuikit.ComposeValidators(tuikit.Required(), tuikit.MinLength(8))
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
	v := tuikit.RegexValidator(`^\d{4}$`, "must be 4 digits")
	if v("1234") != nil {
		t.Error("unexpected error for 1234")
	}
	if v("abc") == nil {
		t.Error("expected error for abc")
	}
}

func TestFormNavigation(t *testing.T) {
	opts := tuikit.FormOpts{
		Groups: []tuikit.FormGroup{{Fields: []tuikit.Field{
			tuikit.NewTextField("a", "FieldA"),
			tuikit.NewTextField("b", "FieldB"),
		}}},
	}
	tm := tuitest.NewTestModel(t, newFormModel(opts), 80, 24)
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
	opts := tuikit.FormOpts{
		Groups: []tuikit.FormGroup{{Fields: []tuikit.Field{
			tuikit.NewTextField("name", "Name").WithDefault("Alice"),
		}}},
		OnSubmit: func(v map[string]string) { captured = v },
	}
	tm := tuitest.NewTestModel(t, newFormModel(opts), 80, 24)
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
	opts := tuikit.FormOpts{
		Groups: []tuikit.FormGroup{{Fields: []tuikit.Field{
			tuikit.NewTextField("email", "Email").
				WithRequired().
				WithValidator(tuikit.EmailValidator()),
		}}},
		OnSubmit: func(_ map[string]string) { submitted = true },
	}
	tm := tuitest.NewTestModel(t, newFormModel(opts), 80, 24)
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
	opts := tuikit.FormOpts{
		WizardMode: true,
		Groups: []tuikit.FormGroup{{Fields: []tuikit.Field{
			tuikit.NewTextField("s1", "Step 1"),
			tuikit.NewTextField("s2", "Step 2"),
			tuikit.NewTextField("s3", "Step 3"),
		}}},
		OnSubmit: func(_ map[string]string) { submitted = true },
	}
	tm := tuitest.NewTestModel(t, newFormModel(opts), 80, 24)
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
	opts := tuikit.FormOpts{
		Groups: []tuikit.FormGroup{{
			Title: "Account",
			Fields: []tuikit.Field{
				tuikit.NewTextField("u", "Username"),
				tuikit.NewPasswordField("p", "Password"),
				tuikit.NewSelectField("r", "Role", []string{"Admin", "User"}),
				tuikit.NewMultiSelectField("x", "Permissions", []string{"Read", "Write"}),
				tuikit.NewConfirmField("c", "Confirm"),
				tuikit.NewNumberField("n", "Age").WithMin(18).WithMax(120),
			},
		}},
	}
	tm := tuitest.NewTestModel(t, newFormModel(opts), 80, 40)
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
	f := tuikit.NewNumberField("n", "N").WithMin(18).WithMax(120)
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
	f := tuikit.NewSelectField("l", "L", []string{"Go", "Rust", "Python"})
	if f.Value() != "Go" {
		t.Errorf("want Go, got %q", f.Value())
	}
	f.WithDefault("Rust")
	if f.Value() != "Rust" {
		t.Errorf("want Rust, got %q", f.Value())
	}
}

func TestFormMultiSelectField(t *testing.T) {
	f := tuikit.NewMultiSelectField("t", "T", []string{"Go", "TUI", "CLI"})
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
	f := tuikit.NewConfirmField("c", "C")
	if f.Value() != "false" {
		t.Errorf("want false, got %q", f.Value())
	}
	f.WithDefault(true)
	if f.Value() != "true" {
		t.Errorf("want true, got %q", f.Value())
	}
}

func TestFormGroupSeparator(t *testing.T) {
	opts := tuikit.FormOpts{
		Groups: []tuikit.FormGroup{
			{Title: "Personal", Fields: []tuikit.Field{tuikit.NewTextField("n", "Name")}},
			{Title: "Account", Fields: []tuikit.Field{tuikit.NewTextField("e", "Email")}},
		},
	}
	tm := tuitest.NewTestModel(t, newFormModel(opts), 80, 24)
	scr := tm.Screen()
	if !scr.Contains("Personal") {
		t.Error("expected Personal group title")
	}
	if !scr.Contains("Account") {
		t.Error("expected Account group title")
	}
}
