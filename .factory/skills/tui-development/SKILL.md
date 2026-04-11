---
name: tui-development
description: Develop and test blit TUI components using the Component interface, btest harness, and theme system
---

# TUI Component Development

## Creating a New Component

1. Define a struct implementing the `Component` interface:
   - `Init() tea.Cmd`
   - `Update(msg tea.Msg, ctx Context) (Component, tea.Cmd)` — return `Consumed()` or `notConsumed()`
   - `View() string`
   - `KeyBindings() []KeyBind`
   - `SetSize(width, height int)`
   - `Focused() bool` / `SetFocused(bool)`

2. Thread `Context` through all Update calls — it carries Theme, Size, Focus, Hotkeys, Clock, Logger.

3. Write tests alongside the component file:
   - `component_test.go` — basic behavior
   - `component_coverage_test.go` — edge cases and comprehensive coverage
   - `component_edge_test.go` — boundary conditions

4. Use `btest.NewHarness()` for integration tests with golden file comparison.

5. Never call `os.Exit`, `log.Fatal`, or `panic` in library code.
