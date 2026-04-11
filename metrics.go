package blit

import (
	"fmt"
	"log/slog"
	"time"
)

// MetricsCollector receives timing and event data from the App's internal
// lifecycle. Implement this interface to integrate with Prometheus,
// Datadog, OpenTelemetry, or any custom metrics backend.
//
// All methods are optional — nil receivers are safe. The App never blocks
// on metric emission; collectors should offload work if needed.
type MetricsCollector interface {
	// OnRender is called after each frame render with the time it took
	// to produce the View output for all visible components.
	OnRender(duration time.Duration)

	// OnDispatch is called after a tea.Msg is dispatched to a component,
	// with the component name and the time taken by its Update method.
	OnDispatch(component string, duration time.Duration)

	// OnPanic is called when a component panics during Update or View
	// and is recovered by the App's error boundary.
	OnPanic(component string, recovered any)
}

// LoggingMetrics is a MetricsCollector that emits structured log entries
// for each event. Useful for development and debugging without an
// external metrics backend.
type LoggingMetrics struct {
	Logger *slog.Logger
}

// OnRender logs the render duration at debug level.
func (m *LoggingMetrics) OnRender(duration time.Duration) {
	if m.Logger != nil {
		m.Logger.Debug("render", "duration_ms", duration.Milliseconds())
	}
}

// OnDispatch logs the dispatch duration at debug level.
func (m *LoggingMetrics) OnDispatch(component string, duration time.Duration) {
	if m.Logger != nil {
		m.Logger.Debug("dispatch", "component", component, "duration_ms", duration.Milliseconds())
	}
}

// OnPanic logs a recovered panic at error level.
func (m *LoggingMetrics) OnPanic(component string, recovered any) {
	if m.Logger != nil {
		m.Logger.Error("panic recovered", "component", component, "recovered", fmt.Sprintf("%v", recovered))
	}
}
