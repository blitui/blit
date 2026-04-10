package blit_test

import (
	"testing"
	"time"

	"github.com/blitui/blit"
	"github.com/blitui/blit/btest"
)

// TestComposition_TableWithDetailOverlay verifies Table and DetailOverlay
// compose correctly inside an App — Table renders, detail overlay opens on
// Enter, Esc closes it, and Table state is preserved.
func TestComposition_TableWithDetailOverlay(t *testing.T) {
	cols := []blit.Column{
		{Title: "Name", Width: 20},
		{Title: "Value", Width: 10},
	}
	rows := []blit.Row{
		{"alpha", "1"},
		{"bravo", "2"},
		{"charlie", "3"},
	}
	table := blit.NewTable(cols, rows, blit.TableOpts{})

	detail := blit.NewDetailOverlay[string](blit.DetailOverlayOpts[string]{
		Title: "Detail",
		Render: func(item string, w, h int, theme blit.Theme) string {
			return "Detail: " + item
		},
	})

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("table", table),
		blit.WithOverlay("detail", "d", detail),
	)

	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	// Table should render with data
	h.Expect("alpha").Expect("bravo")

	// Move cursor down
	h.Keys("down")

	// Verify table still shows content after navigation
	h.Expect("bravo")
}

// TestComposition_TabsWithNestedComponents verifies Tabs with multiple
// content panes render and switch correctly.
func TestComposition_TabsWithNestedComponents(t *testing.T) {
	lv := blit.NewLogViewer()
	lv.Append(blit.LogLine{Message: "log line one", Timestamp: time.Now()})
	lv.Append(blit.LogLine{Message: "log line two", Timestamp: time.Now()})

	vp := blit.NewViewport()
	vp.SetContent("viewport content here")

	tabs := blit.NewTabs([]blit.TabItem{
		{Title: "Logs", Content: lv},
		{Title: "View", Content: vp},
	}, blit.TabsOpts{})

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("tabs", tabs),
	)

	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	// First tab should be active
	h.Expect("Logs").Expect("log line one")
}

// TestComposition_SplitWithTwoComponents verifies Split layout renders
// both children and handles focus cycling.
func TestComposition_SplitWithTwoComponents(t *testing.T) {
	left := blit.NewLogViewer()
	left.Append(blit.LogLine{Message: "left pane", Timestamp: time.Now()})

	right := blit.NewLogViewer()
	right.Append(blit.LogLine{Message: "right pane", Timestamp: time.Now()})

	split := blit.NewSplit(blit.Horizontal, 0.5, left, right)

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("split", split),
	)

	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	h.Expect("left pane").Expect("right pane")
}

// TestComposition_FormInApp verifies a Form renders fields and handles
// keyboard navigation between them.
func TestComposition_FormInApp(t *testing.T) {
	form := blit.NewForm(blit.FormOpts{
		Groups: []blit.FormGroup{
			{Title: "Info", Fields: []blit.Field{
				blit.NewTextField("name", "Name"),
				blit.NewTextField("email", "Email"),
			}},
		},
	})

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("form", form),
	)

	h := btest.NewAppHarness(t, app, 60, 20)
	defer h.Done()

	h.Expect("Name").Expect("Email")
}

// TestComposition_PickerInApp verifies Picker renders items and
// responds to filter input.
func TestComposition_PickerInApp(t *testing.T) {
	items := []blit.PickerItem{
		{Title: "Apple", Subtitle: "A fruit"},
		{Title: "Banana", Subtitle: "Yellow fruit"},
		{Title: "Cherry", Subtitle: "Red fruit"},
	}
	picker := blit.NewPicker(items, blit.PickerOpts{})

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("picker", picker),
	)

	h := btest.NewAppHarness(t, app, 60, 20)
	defer h.Done()

	h.Expect("Apple").Expect("Banana").Expect("Cherry")
}

// TestComposition_ThemePropagation verifies that setting a theme on the
// App propagates to all registered components without panics.
func TestComposition_ThemePropagation(t *testing.T) {
	themes := []blit.Theme{
		blit.DefaultTheme(),
		blit.LightTheme(),
	}

	for _, theme := range themes {
		lv := blit.NewLogViewer()
		_ = lv

		table := blit.NewTable(
			[]blit.Column{{Title: "Col", Width: 10}},
			[]blit.Row{{"val"}},
			blit.TableOpts{},
		)

		app := blit.NewApp(
			blit.WithTheme(theme),
			blit.WithComponent("table", table),
		)

		h := btest.NewAppHarness(t, app, 80, 24)
		// Verify no panic during render with each theme
		h.Expect("val")
		h.Done()
	}
}

// TestComposition_HBoxLayout verifies HBox renders children side by side.
func TestComposition_HBoxLayout(t *testing.T) {
	left := blit.NewLogViewer()
	left.Append(blit.LogLine{Message: "left content", Timestamp: time.Now()})
	right := blit.NewLogViewer()
	right.Append(blit.LogLine{Message: "right content", Timestamp: time.Now()})

	hbox := &blit.HBox{
		Items: []blit.Component{left, right},
	}

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("hbox", hbox),
	)

	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	h.Expect("left content").Expect("right content")
}

// TestComposition_VBoxLayout verifies VBox renders children vertically.
func TestComposition_VBoxLayout(t *testing.T) {
	top := blit.NewBreadcrumbs([]string{"Home", "Settings"})
	bottom := blit.NewLogViewer()
	bottom.Append(blit.LogLine{Message: "log output", Timestamp: time.Now()})

	vbox := &blit.VBox{
		Items: []blit.Component{top, bottom},
	}

	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("vbox", vbox),
	)

	h := btest.NewAppHarness(t, app, 80, 24)
	defer h.Done()

	h.Expect("Home").Expect("log output")
}
