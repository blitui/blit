package blit_test

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	blit "github.com/blitui/blit"
)

func ExampleNewApp() {
	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithHelp(),
	)
	// app.Run() starts the TUI event loop.
	_ = app
	// Output:
}

func ExampleNewApp_withStatusBar() {
	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithStatusBar(
			func() string { return " ? help  q quit" },
			func() string { return "" },
		),
		blit.WithHelp(),
	)
	_ = app
	// Output:
}

func ExampleNewTable() {
	columns := []blit.Column{
		{Title: "Name", Width: 20, Sortable: true},
		{Title: "Value", Width: 10, Align: blit.Right},
		{Title: "Status", Width: 10},
	}
	rows := []blit.Row{
		{"Alice", "100", "Active"},
		{"Bob", "200", "Inactive"},
		{"Carol", "300", "Active"},
	}
	table := blit.NewTable(columns, rows, blit.TableOpts{
		Sortable:   true,
		Filterable: true,
	})
	_ = table
	// Output:
}

func ExampleNewPoller() {
	fetchData := func() tea.Cmd {
		return func() tea.Msg {
			return "fetched"
		}
	}
	poller := blit.NewPoller(30*time.Second, fetchData)

	// In your component's Update:
	//   case blit.TickMsg:
	//       return m, poller.Check(msg.(blit.TickMsg))
	_ = poller
	// Output:
}

func ExampleNewPollerWithOpts() {
	poller := blit.NewPollerWithOpts(blit.PollerOpts{
		Name:     "api-events",
		Interval: 30 * time.Second,
		Fetch: func() (tea.Msg, error) {
			return "events fetched", nil
		},
		Backoff:    blit.ExponentialBackoff(time.Second, 5*time.Minute),
		MaxRetries: 3,
	})
	_ = poller
	// Output:
}

func ExampleNewPollerWithOpts_fetchCtx() {
	poller := blit.NewPollerWithOpts(blit.PollerOpts{
		Name:     "api-events",
		Interval: 30 * time.Second,
		FetchCtx: func(ctx context.Context) (tea.Msg, error) {
			// Use ctx for HTTP request cancellation, timeouts, etc.
			return "events fetched", nil
		},
		Backoff:    blit.ExponentialBackoff(time.Second, 5*time.Minute),
		MaxRetries: 3,
	})
	_ = poller
	// Output:
}

func ExampleLoadConfig() {
	type AppConfig struct {
		Interval int    `yaml:"interval" blit:"label=Interval (sec),group=Polling,default=30,min=5"`
		Theme    string `yaml:"theme"    blit:"label=Theme,group=Appearance,default=dark"`
	}

	cfg, err := blit.LoadConfig[AppConfig]("myapp")
	if err != nil {
		// Handle missing or invalid config file.
		fmt.Println("config error:", err)
	}
	_ = cfg
}

func ExampleConfig_Editor() {
	type AppConfig struct {
		Interval int    `yaml:"interval" blit:"label=Interval (sec),group=Polling,default=30,min=5"`
		Theme    string `yaml:"theme"    blit:"label=Theme,group=Appearance,default=dark"`
	}

	cfg := &blit.Config[AppConfig]{
		Value: AppConfig{Interval: 30, Theme: "dark"},
	}

	// Editor auto-generates a ConfigEditor from struct tags.
	editor := cfg.Editor()
	_ = editor
	// Output:
}

func ExampleNewSignal() {
	status := blit.NewSignal("starting...")

	// Read from any goroutine.
	fmt.Println(status.Get())

	// Write from any goroutine; subscribers fire on the UI goroutine.
	status.Set("ready")
	fmt.Println(status.Get())
	// Output:
	// starting...
	// ready
}

func ExampleSignal_Subscribe() {
	count := blit.NewSignal(0)

	unsub := count.Subscribe(func(v int) {
		fmt.Println("count is now", v)
	})
	defer unsub()

	// Set triggers the subscriber on the next signal flush.
	count.Set(42)
	_ = count
	// Output:
}

func ExampleExponentialBackoff() {
	backoff := blit.ExponentialBackoff(time.Second, 5*time.Minute)

	fmt.Println(backoff.NextBackoff(0)) // 1s
	fmt.Println(backoff.NextBackoff(1)) // 2s
	fmt.Println(backoff.NextBackoff(2)) // 4s
	// Output:
	// 1s
	// 2s
	// 4s
}

func ExampleFixedBackoff() {
	backoff := blit.FixedBackoff(2*time.Second, 3)

	fmt.Println(backoff.NextBackoff(0)) // 2s
	fmt.Println(backoff.NextBackoff(2)) // 2s
	fmt.Println(backoff.NextBackoff(3)) // -1ns (no more retries)
	// Output:
	// 2s
	// 2s
	// -1ns
}

func ExampleNewLogViewer() {
	lv := blit.NewLogViewer()
	lv.Append(blit.LogLine{Level: blit.LogInfo, Message: "Application started"})
	lv.Append(blit.LogLine{Level: blit.LogInfo, Message: "Connected to database"})
	_ = lv
	// Output:
}

func ExampleNewForm() {
	form := blit.NewForm(blit.FormOpts{
		Groups: []blit.FormGroup{
			{
				Title: "Settings",
				Fields: []blit.Field{
					blit.NewTextField("name", "Name"),
					blit.NewSelectField("role", "Role", []string{"admin", "user", "viewer"}),
					blit.NewConfirmField("active", "Active"),
				},
			},
		},
		OnSubmit: func(values map[string]string) {
			fmt.Println("submitted")
		},
	})
	_ = form
	// Output:
}

func ExampleNewPicker() {
	items := []blit.PickerItem{
		{Title: "Dark", Value: "dark"},
		{Title: "Light", Value: "light"},
		{Title: "Retro", Value: "retro"},
	}
	picker := blit.NewPicker(items, blit.PickerOpts{
		Placeholder: "Choose a theme...",
	})
	_ = picker
	// Output:
}

func ExampleWithModule() {
	// Modules extend the App with custom lifecycle behavior.
	// Built-in modules include the DevConsole and Poller.
	app := blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithDevConsole(),
	)
	_ = app
	// Output:
}

func ExampleNewTree() {
	nodes := []*blit.Node{
		{Title: "src", Children: []*blit.Node{
			{Title: "main.go"},
			{Title: "app.go"},
		}},
		{Title: "go.mod"},
		{Title: "README.md"},
	}
	tree := blit.NewTree(nodes, blit.TreeOpts{})
	_ = tree
	// Output:
}

func ExampleNewSplit() {
	// Create a vertical split with 30/70 ratio.
	left := blit.NewLogViewer()
	right := blit.NewLogViewer()
	split := blit.NewSplit(blit.Vertical, 0.3, left, right)
	_ = split
	// Output:
}

func ExampleDefaultTheme() {
	theme := blit.DefaultTheme()
	fmt.Println("Theme accent:", theme.Accent)
	_ = theme
}
