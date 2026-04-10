# Form

Multi-field form with keyboard navigation, inline validation, optional grouping, and wizard step mode. Implements `Component` and `Themed`.

## Construction

```go
form := blit.NewForm(blit.FormOpts{
    Groups: []blit.FormGroup{
        {
            Title: "Account",
            Fields: []blit.Field{
                blit.TextField("username", "Username", blit.Required()),
                blit.TextField("email", "Email", blit.MatchRegex(`^.+@.+$`, "invalid email")),
                blit.PasswordField("password", "Password", blit.MinLen(8)),
            },
        },
    },
    OnSubmit: func(values map[string]string) {
        fmt.Println("Submitted:", values)
    },
})
```

## FormOpts

```go
type FormOpts struct {
    Groups     []FormGroup              // Field groups (rendered with headers)
    OnSubmit   func(values map[string]string) // Called on Enter at the last field
    WizardMode bool                     // Show one group at a time with step indicator
}
```

## FormGroup

```go
type FormGroup struct {
    Title  string  // Group header text
    Fields []Field // Fields in this group
}
```

## Field Constructors

| Function | Description |
|----------|-------------|
| `TextField(key, label, validators...)` | Single-line text input |
| `PasswordField(key, label, validators...)` | Masked password input |
| `SelectField(key, label, options)` | Dropdown selection |

## Built-In Validators

| Validator | Description |
|-----------|-------------|
| `Required()` | Field must not be empty |
| `MinLen(n)` | Minimum character count |
| `MaxLen(n)` | Maximum character count |
| `MatchRegex(pattern, msg)` | Must match the given regex |

## Keyboard

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Next / previous field |
| `Enter` | Submit (on last field) or next field |
| `Esc` | Cancel / blur form |

## Wizard Mode

Set `WizardMode: true` to show one group at a time. The form renders a step indicator and navigates between groups with Enter/Backspace.

## Example

```bash
go run ./examples/form/
```
