# Testing with blit

`blit` is a virtual-terminal testing framework for Bubble Tea models. It renders your model into an in-memory screen buffer and provides 30+ assertions — no real terminal required.

## Install the blit CLI

```bash
# Homebrew
brew install blitui/tap/blit

# Scoop
scoop bucket add blitui https://github.com/blitui/scoop-bucket
scoop install blit

# Go
go install github.com/blitui/blit/cmd/blit@latest
```

## Writing a Test

```go
import "github.com/blitui/blit/blit"

func TestMyApp(t *testing.T) {
    tm := blit.NewTestModel(t, myModel{}, 80, 24)

    // Interact
    tm.SendKey("down")
    tm.SendKeys("j", "j", "enter")
    tm.Type("hello")
    tm.SendResize(120, 40)
    tm.SendMsg(myCustomMsg{})

    // Assert on rendered screen
    scr := tm.Screen()
    blit.AssertContains(t, scr, "Expected text")
    blit.AssertRowContains(t, scr, 0, "Header")
    blit.AssertMatches(t, scr, `\d+ items`)
    blit.AssertRowCount(t, scr, 5)
}
```

## Assertions Reference

| Assertion | Description |
|-----------|-------------|
| `AssertContains` | Screen contains substring |
| `AssertNotContains` | Screen does not contain substring |
| `AssertRowContains` | Row N contains substring |
| `AssertMatches` | Screen matches regexp |
| `AssertRowCount` | Screen has exactly N non-empty rows |
| `AssertFgAt` | Cell at (row, col) has foreground color |
| `AssertBgAt` | Cell at (row, col) has background color |
| `AssertBoldAt` | Cell at (row, col) is bold |
| `AssertRegionContains` | Rectangular region contains text |
| `AssertScreensEqual` | Two screen snapshots are identical |
| `AssertScreensNotEqual` | Two screen snapshots differ |
| `AssertGolden` | Compare against golden file in `testdata/` |

## Golden File Testing

```go
blit.AssertGolden(t, scr, "my-test")
// compares against testdata/my-test.golden
```

Regenerate snapshots:

```bash
blit -update ./...
```

## Waiting for Async State

```go
ok := tm.WaitFor(blit.UntilContains("loaded"), 10)
if !ok {
    t.Fatal("timed out waiting for 'loaded'")
}
```

## blit CLI Flags

```bash
blit                                    # go test ./...
blit -filter TestHarness ./blit/...  # run tests matching a regexp
blit -update ./blit/...              # regenerate golden snapshots
blit -junit out/junit.xml -parallel 4   # parallel run + JUnit report
blit -html out/report.html              # HTML report
blit -watch                             # re-run on file changes (1s poll)
```

## Vitest-Style Reporter

Run with `-v` for grouped, color-coded output:

```
  blit · terminal test toolkit

  Screen
    ✓ PlainText 0.000ms
    ✓ Contains 0.000ms
  Assert
    ✓ ContainsPass 0.000ms
    ✓ RowMatchesPass 0.000ms

  PASS 96 tests (3ms)
```
