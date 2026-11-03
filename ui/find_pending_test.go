package ui

import (
	"runtime"
	"testing"

	"github.com/egoist/mygo/internal/l10n"
	"github.com/egoist/mygo/internal/platform"
)

func TestFindBarPendingKeepsInputAndBlocksNavigation(t *testing.T) {
	open, pending := true, true
	query, current, matches := "original", 1, 3
	tt := coreNewTester(func(c *context) {
		coreFindBar(c, &open, &query, matches, &current, FindBarOptions{Pending: pending}).FillWidth()
	}, 640, 240)
	if !tt.Focused("Find") || !tt.HasText("Searching…") || tt.HasText("No matches") || tt.HasText("2 of 3") {
		t.Fatalf("pending state: focused=%v texts=%q", tt.Focused("Find"), tt.Texts())
	}
	tt.send(platform.SurfaceEvent{Kind: platform.AccessibilityOn})
	accessNode(t, tt.h.access, platform.RoleStatus, "Searching…")
	tt.Key(0, KeyEnter)
	tt.Key(Shift, KeyEnter)
	if runtime.GOOS == "darwin" {
		tt.Key(Cmd, KeyG)
		tt.Key(Shift|Cmd, KeyG)
	} else {
		tt.Key(0, KeyF3)
		tt.Key(Shift, KeyF3)
	}
	tt.Click("Next")
	tt.Click("Previous")
	if current != 1 {
		t.Fatalf("pending navigation changed index to %d", current)
	}
	tt.Click("Find")
	if runtime.GOOS == "darwin" {
		tt.Key(Cmd, KeyA)
	} else {
		tt.Key(Ctrl, KeyA)
	}
	tt.Type("replacement")
	if query != "replacement" || !tt.HasText("Searching…") {
		t.Fatalf("pending input: query=%q texts=%q", query, tt.Texts())
	}
	pending = false
	matches = 0
	tt.Frame()
	if !tt.HasText("No matches") || tt.HasText("Searching…") {
		t.Fatalf("completed empty search: %q", tt.Texts())
	}
	matches = 2
	tt.Frame()
	tt.Key(0, KeyEnter)
	if current != 1 || !tt.HasText("2 of 2") {
		t.Fatalf("completed navigation: %d %q", current, tt.Texts())
	}
	pending = true
	tt.Frame()
	tt.Key(0, KeyEscape)
	if open {
		t.Fatal("Escape did not close pending search")
	}
	open = true
	tt.Frame()
	tt.Click("Done")
	if open {
		t.Fatal("Done did not close pending search")
	}
}

func TestFindBarPendingLocalizedStatus(t *testing.T) {
	open, query, current := true, "value", 0
	tt := coreNewTester(func(c *context) { coreFindBar(c, &open, &query, 0, &current, FindBarOptions{Pending: true}) }, 640, 200)
	for _, lang := range []string{"en", "ja", "zh-Hans", "zh-Hant"} {
		tt.SetLocale(lang)
		expected := l10n.In(lang).Searching()
		if !tt.HasText(expected) || tt.HasText(l10n.In(lang).NoMatches()) {
			t.Fatalf("locale %s: %q", lang, tt.Texts())
		}
		tt.send(platform.SurfaceEvent{Kind: platform.AccessibilityOn})
		accessNode(t, tt.h.access, platform.RoleStatus, expected)
	}
}
