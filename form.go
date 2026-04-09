package tuikit

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormSubmitMsg is emitted when the form is submitted successfully.
type FormSubmitMsg struct {
	Values map[string]string
}

// Validator is a function that validates a field value.
// Return nil if valid, or an error describing the problem.
type Validator func(value string) error

// Required returns a Validator that fails if the value is empty.
func Required() Validator {
	return func(v string) error {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("this field is required")
		}
		return nil
	}
}

// MinLength returns a Validator that fails if the value is shorter than n characters.
func MinLength(n int) Validator {
	return func(v string) error {
		if len([]rune(v)) < n {
			return fmt.Errorf("must be at least %d characters", n)
		}
		return nil
	}
}

// MaxLength returns a Validator that fails if the value is longer than n characters.
func MaxLength(n int) Validator {
	return func(v string) error {
		if len([]rune(v)) > n {
			return fmt.Errorf("must be at most %d characters", n)
		}
		return nil
	}
}

// EmailValidator returns a Validator that checks for a basic email format.
func EmailValidator() Validator {
	re := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	return func(v string) error {
		if !re.MatchString(v) {
			return fmt.Errorf("must be a valid email address")
		}
		return nil
	}
}

// URLValidator returns a Validator that checks for http/https URL format.
func URLValidator() Validator {
	re := regexp.MustCompile(`^https?://[^\s]+$`)
	return func(v string) error {
		if !re.MatchString(v) {
			return fmt.Errorf("must be a valid URL (http:// or https://)")
		}
		return nil
	}
}

// RegexValidator returns a Validator that checks the value against a pattern.
func RegexValidator(pattern, message string) Validator {
	re := regexp.MustCompile(pattern)
	return func(v string) error {
		if !re.MatchString(v) {
			return fmt.Errorf("%s", message)
		}
		return nil
	}
}

// ComposeValidators returns a Validator that runs all validators in order.
// The first error encountered is returned.
func ComposeValidators(validators ...Validator) Validator {
	return func(v string) error {
		for _, fn := range validators {
			if err := fn(v); err != nil {
				return err
			}
		}
		return nil
	}
}

// Field is the interface all form field types implement.
type Field interface {
	FieldID() string
	Label() string
	Value() string
	SetValue(v string)
	Validate() error
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	View(focused bool, theme Theme, width int) string
	SetFocused(focused bool)
}

// fieldBase holds common field state shared by all field types.
type fieldBase struct {
	id          string
	label       string
	hint        string
	placeholder string
	required    bool
	validator   Validator
	err         error
}

func (f *fieldBase) FieldID() string { return f.id }
func (f *fieldBase) Label() string   { return f.label }

func (f *fieldBase) validate(v string) error {
	if f.required {
		if err := Required()(v); err != nil {
			return err
		}
	}
	if f.validator != nil {
		return f.validator(v)
	}
	return nil
}

func (f *fieldBase) renderLabel(focused bool, theme Theme) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text)).Bold(focused)
	label := f.label
	if f.required {
		req := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Negative)).Render("*")
		label = label + req
	}
	return style.Render(label)
}

func (f *fieldBase) renderHint(theme Theme) string {
	if f.hint == "" {
		return ""
	}
	var s lipgloss.Style
	if ss, ok := theme.Style("label.hint"); ok {
		s = ss.Base
	} else {
		s = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted))
	}
	return s.Render("  " + f.hint)
}

func (f *fieldBase) renderError(theme Theme) string {
	if f.err == nil {
		return ""
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Negative)).Render("  x " + f.err.Error())
}

// TextField is a single-line text input field.
type TextField struct {
	fieldBase
	input textinput.Model
}

// NewTextField creates a new single-line text field.
func NewTextField(id, label string) *TextField {
	ti := textinput.New()
	ti.CharLimit = 256
	return &TextField{fieldBase: fieldBase{id: id, label: label}, input: ti}
}

// WithHint sets the hint text.
func (f *TextField) WithHint(hint string) *TextField { f.hint = hint; return f }

// WithPlaceholder sets the placeholder text.
func (f *TextField) WithPlaceholder(p string) *TextField {
	f.placeholder = p
	f.input.Placeholder = p
	return f
}

// WithDefault sets the initial value.
func (f *TextField) WithDefault(v string) *TextField { f.input.SetValue(v); return f }

// WithRequired marks the field as required.
func (f *TextField) WithRequired() *TextField { f.required = true; return f }

// WithValidator attaches a validator function.
func (f *TextField) WithValidator(v Validator) *TextField { f.validator = v; return f }

func (f *TextField) Value() string     { return f.input.Value() }
func (f *TextField) SetValue(v string) { f.input.SetValue(v) }

func (f *TextField) Validate() error {
	f.err = f.validate(f.input.Value())
	return f.err
}

func (f *TextField) Init() tea.Cmd { return nil }

func (f *TextField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	if _, ok := msg.(tea.KeyMsg); ok {
		f.err = nil
	}
	return cmd
}

func (f *TextField) SetFocused(focused bool) {
	if focused {
		f.input.Focus()
	} else {
		f.input.Blur()
	}
}

func (f *TextField) View(focused bool, theme Theme, width int) string {
	labelLine := f.renderLabel(focused, theme)
	if h := f.renderHint(theme); h != "" {
		labelLine += h
	}
	if focused {
		if ss, ok := theme.Style("input.text"); ok {
			f.input.TextStyle = ss.Focus
		} else {
			f.input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
		}
		if ss, ok := theme.Style("input.focus"); ok {
			f.input.PromptStyle = ss.Focus
		} else {
			f.input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent))
		}
	}
	inputStyle := lipgloss.NewStyle().Width(width - 2)
	lines := []string{labelLine, inputStyle.Render(f.input.View())}
	if errLine := f.renderError(theme); errLine != "" {
		lines = append(lines, errLine)
	}
	return strings.Join(lines, "\n")
}

// PasswordField is a single-line text input that masks its value.
type PasswordField struct {
	fieldBase
	input textinput.Model
}

// NewPasswordField creates a new password input field.
func NewPasswordField(id, label string) *PasswordField {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '*'
	return &PasswordField{fieldBase: fieldBase{id: id, label: label}, input: ti}
}

// WithHint sets the hint text.
func (f *PasswordField) WithHint(hint string) *PasswordField { f.hint = hint; return f }

// WithPlaceholder sets the placeholder text.
func (f *PasswordField) WithPlaceholder(p string) *PasswordField {
	f.placeholder = p
	f.input.Placeholder = p
	return f
}

// WithRequired marks the field as required.
func (f *PasswordField) WithRequired() *PasswordField { f.required = true; return f }

// WithValidator attaches a validator function.
func (f *PasswordField) WithValidator(v Validator) *PasswordField { f.validator = v; return f }

func (f *PasswordField) Value() string     { return f.input.Value() }
func (f *PasswordField) SetValue(v string) { f.input.SetValue(v) }

func (f *PasswordField) Validate() error {
	f.err = f.validate(f.input.Value())
	return f.err
}

func (f *PasswordField) Init() tea.Cmd { return nil }

func (f *PasswordField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	if _, ok := msg.(tea.KeyMsg); ok {
		f.err = nil
	}
	return cmd
}

func (f *PasswordField) SetFocused(focused bool) {
	if focused {
		f.input.Focus()
	} else {
		f.input.Blur()
	}
}

func (f *PasswordField) View(focused bool, theme Theme, width int) string {
	labelLine := f.renderLabel(focused, theme)
	if h := f.renderHint(theme); h != "" {
		labelLine += h
	}
	if focused {
		f.input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent))
	}
	inputStyle := lipgloss.NewStyle().Width(width - 2)
	lines := []string{labelLine, inputStyle.Render(f.input.View())}
	if errLine := f.renderError(theme); errLine != "" {
		lines = append(lines, errLine)
	}
	return strings.Join(lines, "\n")
}

// SelectField is a single-choice field rendered as a cycling selector.
type SelectField struct {
	fieldBase
	options  []string
	selected int
}

// NewSelectField creates a new select field with the given options.
func NewSelectField(id, label string, options []string) *SelectField {
	return &SelectField{fieldBase: fieldBase{id: id, label: label}, options: options}
}

// WithHint sets the hint text.
func (f *SelectField) WithHint(hint string) *SelectField { f.hint = hint; return f }

// WithRequired marks the field as required.
func (f *SelectField) WithRequired() *SelectField { f.required = true; return f }

// WithValidator attaches a validator.
func (f *SelectField) WithValidator(v Validator) *SelectField { f.validator = v; return f }

// WithDefault sets the initially-selected option by value.
func (f *SelectField) WithDefault(v string) *SelectField {
	for i, opt := range f.options {
		if opt == v {
			f.selected = i
			break
		}
	}
	return f
}

func (f *SelectField) Value() string {
	if len(f.options) == 0 {
		return ""
	}
	return f.options[f.selected]
}

func (f *SelectField) SetValue(v string) {
	for i, opt := range f.options {
		if opt == v {
			f.selected = i
			return
		}
	}
}

func (f *SelectField) Validate() error {
	f.err = f.validate(f.Value())
	return f.err
}

func (f *SelectField) Init() tea.Cmd     { return nil }
func (f *SelectField) SetFocused(_ bool) {}

func (f *SelectField) Update(msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "left", "h":
			if f.selected > 0 {
				f.selected--
			}
			f.err = nil
		case "right", "l":
			if f.selected < len(f.options)-1 {
				f.selected++
			}
			f.err = nil
		}
	}
	return nil
}

func (f *SelectField) View(focused bool, theme Theme, _ int) string {
	labelLine := f.renderLabel(focused, theme)
	if h := f.renderHint(theme); h != "" {
		labelLine += h
	}
	var optParts []string
	for i, opt := range f.options {
		if i == f.selected {
			optParts = append(optParts, lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent)).Bold(true).Render("@ "+opt))
		} else {
			optParts = append(optParts, lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted)).Render("o "+opt))
		}
	}
	var row string
	if focused {
		row = "< " + strings.Join(optParts, "  ") + " >"
	} else {
		row = "  " + strings.Join(optParts, "  ")
	}
	lines := []string{labelLine, row}
	if errLine := f.renderError(theme); errLine != "" {
		lines = append(lines, errLine)
	}
	return strings.Join(lines, "\n")
}

// MultiSelectField is a multi-choice field where users toggle options with space.
type MultiSelectField struct {
	fieldBase
	options  []string
	selected map[int]bool
	cursor   int
}

// NewMultiSelectField creates a new multi-select field.
func NewMultiSelectField(id, label string, options []string) *MultiSelectField {
	return &MultiSelectField{
		fieldBase: fieldBase{id: id, label: label},
		options:   options,
		selected:  make(map[int]bool),
	}
}

// WithHint sets the hint text.
func (f *MultiSelectField) WithHint(hint string) *MultiSelectField { f.hint = hint; return f }

// WithRequired marks the field as required (at least one option must be selected).
func (f *MultiSelectField) WithRequired() *MultiSelectField { f.required = true; return f }

func (f *MultiSelectField) Value() string {
	var parts []string
	for i, opt := range f.options {
		if f.selected[i] {
			parts = append(parts, opt)
		}
	}
	return strings.Join(parts, ",")
}

func (f *MultiSelectField) SetValue(v string) {
	f.selected = make(map[int]bool)
	for _, part := range strings.Split(v, ",") {
		for i, opt := range f.options {
			if opt == strings.TrimSpace(part) {
				f.selected[i] = true
			}
		}
	}
}

func (f *MultiSelectField) Validate() error {
	if f.required && len(f.selected) == 0 {
		f.err = fmt.Errorf("select at least one option")
		return f.err
	}
	f.err = nil
	return nil
}

func (f *MultiSelectField) Init() tea.Cmd     { return nil }
func (f *MultiSelectField) SetFocused(_ bool) {}

func (f *MultiSelectField) Update(msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "left", "h":
			if f.cursor > 0 {
				f.cursor--
			}
			f.err = nil
		case "right", "l":
			if f.cursor < len(f.options)-1 {
				f.cursor++
			}
			f.err = nil
		case " ":
			f.selected[f.cursor] = !f.selected[f.cursor]
			f.err = nil
		}
	}
	return nil
}

func (f *MultiSelectField) View(focused bool, theme Theme, _ int) string {
	labelLine := f.renderLabel(focused, theme)
	if focused {
		labelLine += lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted)).Render("  space to toggle")
	}
	if h := f.renderHint(theme); h != "" {
		labelLine += "  " + h
	}
	var optParts []string
	for i, opt := range f.options {
		isCursor := focused && i == f.cursor
		isSelected := f.selected[i]
		check := "[ ]"
		if isSelected {
			check = "[x]"
		}
		var s lipgloss.Style
		switch {
		case isCursor:
			s = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent)).Bold(true)
		case isSelected:
			s = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Positive))
		default:
			s = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted))
		}
		optParts = append(optParts, s.Render(check+" "+opt))
	}
	row := "  " + strings.Join(optParts, "  ")
	lines := []string{labelLine, row}
	if errLine := f.renderError(theme); errLine != "" {
		lines = append(lines, errLine)
	}
	return strings.Join(lines, "\n")
}

// ConfirmField is a yes/no boolean field.
type ConfirmField struct {
	fieldBase
	value bool
}

// NewConfirmField creates a new yes/no confirm field.
func NewConfirmField(id, label string) *ConfirmField {
	return &ConfirmField{fieldBase: fieldBase{id: id, label: label}}
}

// WithHint sets the hint text.
func (f *ConfirmField) WithHint(hint string) *ConfirmField { f.hint = hint; return f }

// WithDefault sets the initial boolean value.
func (f *ConfirmField) WithDefault(v bool) *ConfirmField { f.value = v; return f }

func (f *ConfirmField) Value() string {
	if f.value {
		return "true"
	}
	return "false"
}

func (f *ConfirmField) SetValue(v string) {
	f.value = v == "true" || v == "yes" || v == "1"
}

func (f *ConfirmField) Validate() error   { f.err = nil; return nil }
func (f *ConfirmField) Init() tea.Cmd     { return nil }
func (f *ConfirmField) SetFocused(_ bool) {}

func (f *ConfirmField) Update(msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "left", "h", "right", "l", " ":
			f.value = !f.value
		case "y", "Y":
			f.value = true
		case "n", "N":
			f.value = false
		}
	}
	return nil
}

func (f *ConfirmField) View(focused bool, theme Theme, _ int) string {
	labelLine := f.renderLabel(focused, theme)
	if h := f.renderHint(theme); h != "" {
		labelLine += h
	}
	yesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted))
	noStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted))
	if f.value {
		yesStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Positive)).Bold(true)
	} else {
		noStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Negative)).Bold(true)
	}
	row := "  " + yesStyle.Render("Yes") + "  " + noStyle.Render("No")
	return strings.Join([]string{labelLine, row}, "\n")
}

// NumberField is a text field that validates numeric input.
type NumberField struct {
	fieldBase
	input textinput.Model
	min   *float64
	max   *float64
}

// NewNumberField creates a new numeric input field.
func NewNumberField(id, label string) *NumberField {
	ti := textinput.New()
	ti.CharLimit = 32
	return &NumberField{fieldBase: fieldBase{id: id, label: label}, input: ti}
}

// WithHint sets the hint text.
func (f *NumberField) WithHint(hint string) *NumberField { f.hint = hint; return f }

// WithPlaceholder sets the placeholder text.
func (f *NumberField) WithPlaceholder(p string) *NumberField {
	f.placeholder = p
	f.input.Placeholder = p
	return f
}

// WithRequired marks the field as required.
func (f *NumberField) WithRequired() *NumberField { f.required = true; return f }

// WithMin sets the minimum numeric value.
func (f *NumberField) WithMin(min float64) *NumberField { f.min = &min; return f }

// WithMax sets the maximum numeric value.
func (f *NumberField) WithMax(max float64) *NumberField { f.max = &max; return f }

// WithDefault sets the initial numeric value.
func (f *NumberField) WithDefault(v float64) *NumberField {
	f.input.SetValue(strconv.FormatFloat(v, 'f', -1, 64))
	return f
}

func (f *NumberField) Value() string     { return f.input.Value() }
func (f *NumberField) SetValue(v string) { f.input.SetValue(v) }

func (f *NumberField) Validate() error {
	v := strings.TrimSpace(f.input.Value())
	if f.required && v == "" {
		f.err = fmt.Errorf("this field is required")
		return f.err
	}
	if v != "" {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			f.err = fmt.Errorf("must be a valid number")
			return f.err
		}
		if f.min != nil && n < *f.min {
			f.err = fmt.Errorf("must be at least %g", *f.min)
			return f.err
		}
		if f.max != nil && n > *f.max {
			f.err = fmt.Errorf("must be at most %g", *f.max)
			return f.err
		}
	}
	f.err = nil
	return nil
}

func (f *NumberField) Init() tea.Cmd { return nil }

func (f *NumberField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	if _, ok := msg.(tea.KeyMsg); ok {
		f.err = nil
	}
	return cmd
}

func (f *NumberField) SetFocused(focused bool) {
	if focused {
		f.input.Focus()
	} else {
		f.input.Blur()
	}
}

func (f *NumberField) View(focused bool, theme Theme, width int) string {
	labelLine := f.renderLabel(focused, theme)
	if h := f.renderHint(theme); h != "" {
		labelLine += h
	}
	if focused {
		f.input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent))
	}
	inputStyle := lipgloss.NewStyle().Width(width - 2)
	lines := []string{labelLine, inputStyle.Render(f.input.View())}
	if errLine := f.renderError(theme); errLine != "" {
		lines = append(lines, errLine)
	}
	return strings.Join(lines, "\n")
}

// FormGroup is a named group of fields with a visual separator.
type FormGroup struct {
	Title  string
	Fields []Field
}

// FormOpts configures a Form component.
type FormOpts struct {
	Groups     []FormGroup
	OnSubmit   func(values map[string]string)
	WizardMode bool
}

// Form is a TUI component that renders a collection of fields with keyboard
// navigation, inline validation, optional grouping, and a wizard step mode.
// It implements Component and Themed.
type Form struct {
	opts       FormOpts
	theme      Theme
	focused    bool
	width      int
	height     int
	focusIndex int
	allFields  []Field
	submitted  bool
	wizardStep int
}

// NewForm creates a Form from the given options.
func NewForm(opts FormOpts) *Form {
	f := &Form{opts: opts}
	f.buildIndex()
	return f
}

func (f *Form) buildIndex() {
	f.allFields = nil
	for _, g := range f.opts.Groups {
		f.allFields = append(f.allFields, g.Fields...)
	}
	if len(f.allFields) > 0 {
		f.allFields[0].SetFocused(true)
	}
}

func (f *Form) totalFields() int { return len(f.allFields) }

func (f *Form) currentField() Field {
	if f.focusIndex < 0 || f.focusIndex >= len(f.allFields) {
		return nil
	}
	return f.allFields[f.focusIndex]
}

func (f *Form) moveFocusTo(idx int) {
	for i, field := range f.allFields {
		field.SetFocused(i == idx)
	}
	f.focusIndex = idx
}

func (f *Form) submit() tea.Cmd {
	hasErr := false
	for _, field := range f.allFields {
		if err := field.Validate(); err != nil {
			hasErr = true
		}
	}
	if hasErr {
		return nil
	}
	values := make(map[string]string, len(f.allFields))
	for _, field := range f.allFields {
		values[field.FieldID()] = field.Value()
	}
	f.submitted = true
	if f.opts.OnSubmit != nil {
		f.opts.OnSubmit(values)
	}
	return func() tea.Msg { return FormSubmitMsg{Values: values} }
}

// Init implements Component.
func (f *Form) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, field := range f.allFields {
		if cmd := field.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// Update implements Component.
func (f *Form) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !f.focused {
		return f, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if f.opts.WizardMode {
				if f.wizardStep < f.totalFields()-1 {
					if cur := f.allFields[f.wizardStep]; cur != nil {
						cur.Validate()
					}
					f.wizardStep++
					f.moveFocusTo(f.wizardStep)
				} else {
					return f, f.submit()
				}
			} else {
				if f.focusIndex < len(f.allFields)-1 {
					f.moveFocusTo(f.focusIndex + 1)
				}
			}
			return f, Consumed()
		case "shift+tab":
			if f.opts.WizardMode {
				if f.wizardStep > 0 {
					f.wizardStep--
					f.moveFocusTo(f.wizardStep)
				}
			} else {
				if f.focusIndex > 0 {
					f.moveFocusTo(f.focusIndex - 1)
				}
			}
			return f, Consumed()
		case "enter":
			if f.opts.WizardMode {
				if f.wizardStep < f.totalFields()-1 {
					if cur := f.allFields[f.wizardStep]; cur != nil {
						cur.Validate()
					}
					f.wizardStep++
					f.moveFocusTo(f.wizardStep)
					return f, Consumed()
				}
				return f, f.submit()
			}
			if f.focusIndex == len(f.allFields)-1 {
				return f, f.submit()
			}
			f.moveFocusTo(f.focusIndex + 1)
			return f, Consumed()
		}
		if cur := f.currentField(); cur != nil {
			cmd := cur.Update(msg)
			return f, cmd
		}
	case tea.WindowSizeMsg:
		f.SetSize(msg.Width, msg.Height)
	}
	if cur := f.currentField(); cur != nil {
		cur.Update(msg)
	}
	return f, nil
}

// View implements Component.
func (f *Form) View() string {
	if len(f.allFields) == 0 {
		return ""
	}
	if f.opts.WizardMode {
		return f.viewWizard()
	}
	return f.viewNormal()
}

func (f *Form) viewNormal() string {
	var sections []string
	fieldIdx := 0
	for gi, group := range f.opts.Groups {
		if group.Title != "" {
			sepStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(f.theme.Muted)).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color(f.theme.Border)).
				Width(f.width - 2)
			sections = append(sections, sepStyle.Render(group.Title))
		} else if gi > 0 {
			sep := lipgloss.NewStyle().
				Foreground(lipgloss.Color(f.theme.Border)).
				Render(strings.Repeat("-", f.width-2))
			sections = append(sections, sep)
		}
		for _, field := range group.Fields {
			isFocused := f.focused && fieldIdx == f.focusIndex
			fieldView := field.View(isFocused, f.theme, f.width-4)
			if isFocused {
				accent := lipgloss.NewStyle().Foreground(lipgloss.Color(f.theme.Accent)).Render("|")
				fieldView = lipgloss.JoinHorizontal(lipgloss.Top, accent, " ", fieldView)
			} else {
				fieldView = "   " + fieldView
			}
			sections = append(sections, fieldView)
			sections = append(sections, "")
			fieldIdx++
		}
	}
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color(f.theme.Muted)).
		Render("tab/shift+tab: navigate  enter on last field: submit")
	sections = append(sections, hint)
	return strings.Join(sections, "\n")
}

func (f *Form) viewWizard() string {
	if f.wizardStep >= len(f.allFields) {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(f.theme.Positive)).Render("Form submitted")
	}
	field := f.allFields[f.wizardStep]
	total := len(f.allFields)
	progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(f.theme.Muted))
	var dots []string
	for i := 0; i < total; i++ {
		switch {
		case i == f.wizardStep:
			dots = append(dots, lipgloss.NewStyle().Foreground(lipgloss.Color(f.theme.Accent)).Render("*"))
		case i < f.wizardStep:
			dots = append(dots, lipgloss.NewStyle().Foreground(lipgloss.Color(f.theme.Positive)).Render("v"))
		default:
			dots = append(dots, progressStyle.Render("o"))
		}
	}
	progress := strings.Join(dots, " ")
	stepLabel := progressStyle.Render(fmt.Sprintf("Step %d of %d", f.wizardStep+1, total))
	header := progress + "  " + stepLabel
	fieldView := field.View(f.focused, f.theme, f.width-4)
	var navParts []string
	if f.wizardStep > 0 {
		navParts = append(navParts, lipgloss.NewStyle().Foreground(lipgloss.Color(f.theme.Accent)).Render("< back (shift+tab)"))
	}
	if f.wizardStep < total-1 {
		navParts = append(navParts, lipgloss.NewStyle().Foreground(lipgloss.Color(f.theme.Accent)).Render("next > (enter/tab)"))
	} else {
		navParts = append(navParts, lipgloss.NewStyle().Foreground(lipgloss.Color(f.theme.Positive)).Render("submit (enter)"))
	}
	nav := strings.Join(navParts, "   ")
	return strings.Join([]string{header, "", fieldView, "", nav}, "\n")
}

// KeyBindings implements Component.
func (f *Form) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "tab", Label: "Next field", Group: "FORM"},
		{Key: "shift+tab", Label: "Previous field", Group: "FORM"},
		{Key: "enter", Label: "Submit / next", Group: "FORM"},
	}
}

// SetSize implements Component.
func (f *Form) SetSize(width, height int) { f.width = width; f.height = height }

// Focused implements Component.
func (f *Form) Focused() bool { return f.focused }

// SetFocused implements Component.
func (f *Form) SetFocused(focused bool) {
	f.focused = focused
	if cur := f.currentField(); cur != nil {
		cur.SetFocused(focused)
	}
}

// SetTheme implements Themed.
func (f *Form) SetTheme(th Theme) { f.theme = th }

// Values returns the current values of all fields as a map.
func (f *Form) Values() map[string]string {
	out := make(map[string]string, len(f.allFields))
	for _, field := range f.allFields {
		out[field.FieldID()] = field.Value()
	}
	return out
}

// Reset clears the form submission state and returns focus to the first field.
func (f *Form) Reset() {
	f.submitted = false
	f.wizardStep = 0
	f.moveFocusTo(0)
}
