package blit

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type mockMetrics struct {
	renders    []time.Duration
	dispatches []dispatchRecord
	panics     []panicRecord
}

type dispatchRecord struct {
	component string
	duration  time.Duration
}

type panicRecord struct {
	component string
	recovered any
}

func (m *mockMetrics) OnRender(d time.Duration) { m.renders = append(m.renders, d) }
func (m *mockMetrics) OnDispatch(c string, d time.Duration) {
	m.dispatches = append(m.dispatches, dispatchRecord{c, d})
}
func (m *mockMetrics) OnPanic(c string, r any) { m.panics = append(m.panics, panicRecord{c, r}) }

func TestMetricsCollectorInterface(t *testing.T) {
	t.Parallel()
	mm := &mockMetrics{}
	ff := NewFeatureFlag(FlagDef{Name: "TEST", Default: false, Description: "test"})

	app := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("test", &stubComponent{name: "test"}),
		WithMetrics(mm),
		WithFeatureFlags(ff),
	)

	// View should trigger OnRender
	app.width = 80
	app.height = 24
	_ = app.View()
	if len(mm.renders) == 0 {
		t.Error("expected OnRender to be called after View()")
	}
}

func TestLoggingMetrics(t *testing.T) {
	t.Parallel()
	lm := &LoggingMetrics{} // nil logger — should not panic
	lm.OnRender(time.Millisecond)
	lm.OnDispatch("table", time.Millisecond)
	lm.OnPanic("table", "oops")
}

func TestOnErrorHook(t *testing.T) {
	t.Parallel()
	var capturedRecovered any

	hook := func(component string, recovered any) {
		_ = component
		capturedRecovered = recovered
	}

	app := newAppModel(
		WithTheme(DefaultTheme()),
		WithComponent("panicker", &panicStubComponent{}),
		WithOnError(hook),
	)
	app.width = 80
	app.height = 24

	// safeView should recover and call onError
	_ = safeView(&panicStubComponent{}, app.theme, app.onError)
	if capturedRecovered == nil {
		t.Error("expected onError to be called with panic value")
	}
}

func TestTraceIDIncrements(t *testing.T) {
	t.Parallel()
	app := newAppModel(WithTheme(DefaultTheme()))
	ctx1 := app.ctx()
	ctx2 := app.ctx()
	if ctx1.TraceID == 0 {
		t.Error("TraceID should be non-zero")
	}
	if ctx1.TraceID == ctx2.TraceID {
		t.Error("consecutive ctx() calls should produce different TraceIDs")
	}
}

func TestContextFlagsField(t *testing.T) {
	t.Parallel()
	ff := NewFeatureFlag(FlagDef{Name: "BETA", Default: true, Description: "beta feature"})
	app := newAppModel(
		WithTheme(DefaultTheme()),
		WithFeatureFlags(ff),
	)
	ctx := app.ctx()
	if ctx.Flags == nil {
		t.Fatal("Context.Flags should not be nil when WithFeatureFlags is set")
	}
	if !ctx.Flags.Enabled("BETA") {
		t.Error("BETA flag should be enabled")
	}
}

// panicStubComponent panics on View/Update for testing error boundaries.
type panicStubComponent struct{}

func (p *panicStubComponent) Init() tea.Cmd                                        { return nil }
func (p *panicStubComponent) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) { panic("boom") }
func (p *panicStubComponent) View() string                                         { panic("view boom") }
func (p *panicStubComponent) KeyBindings() []KeyBind                               { return nil }
func (p *panicStubComponent) SetSize(w, h int)                                     {}
func (p *panicStubComponent) Focused() bool                                        { return false }
func (p *panicStubComponent) SetFocused(bool)                                      {}
