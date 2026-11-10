package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/internal/platform"
)

type tooltipFixture struct {
	tt       *Tester
	label    string
	disabled bool
	removed  bool
	scroll   ScrollState
}

func newTooltipFixture(width int) *tooltipFixture {
	f := &tooltipFixture{label: "Saves the file"}
	theme := LightTheme()
	theme.TooltipPolicy = &TooltipPolicy{Delay: 300 * time.Millisecond, PreferBelow: true, Hoverable: true, DismissOnInput: true, Gap: 4, ViewportInset: 8}
	theme.TooltipStyle = func(tip Element) {
		tip.Label("tooltip box").MaxWidth(320).Padding(6, 10).Radius(6).Background(theme.Background).FontSize(12)
	}
	f.tt = coreNewTester(func(c *context) {
		c.SetTheme(theme)
		coreColumn(c).Fill().AlignItems(Start).Gap(12).Children(func() {
			coreRow(c).Gap(12).Children(func() {
				if !f.removed {
					coreButton(c, "Save").Label("anchor: Save").Disabled(f.disabled).Tooltip(f.label)
				}
				coreButton(c, "Open").Label("anchor: Open").Tooltip("Opens a file")
				coreButton(c, "Plain")
			})
			coreScroll(c).Width(200).Height(100).TrackScroll(&f.scroll).Children(func() { coreBox(c).Height(400) })
			coreButton(c, "Bottom").Label("anchor: Bottom").Absolute().Bottom(0).Left(0).Tooltip("Bottom hint")
		})
	}, width, 520)
	return f
}

func (f *tooltipFixture) hover(t *testing.T, label string) Rect {
	t.Helper()
	r := at(t, f.tt, "anchor: "+label)
	f.tt.Move(r.X+r.W/2, r.Y+r.H/2)
	return r
}

func TestTooltipPolicyTimingAndEngagement(t *testing.T) {
	f := newTooltipFixture(760)
	tt := f.tt
	save := f.hover(t, "Save")
	tt.Advance(299 * time.Millisecond)
	if tt.HasText(f.label) {
		t.Fatal("pointer hint appeared before 300 ms")
	}
	tt.Advance(time.Millisecond)
	if !tt.HasText(f.label) {
		t.Fatal("pointer hint did not appear at 300 ms")
	}
	r := at(t, tt, "tooltip box")
	if r.X != 8 || r.Y != save.Y+save.H+4 {
		t.Fatalf("hint is not below the anchor with an 8 px viewport inset: %v by %v", r, save)
	}
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(time.Second)
	if !tt.HasText(f.label) {
		t.Fatal("moving into the hint dismissed it")
	}
	f.hover(t, "Open")
	tt.Advance(299 * time.Millisecond)
	if tt.HasText("Opens a file") || tt.HasText(f.label) {
		t.Fatal("warm movement skipped the next hint's delay")
	}
	tt.Advance(time.Millisecond)
	if !tt.HasText("Opens a file") {
		t.Fatal("second pointer hint did not appear")
	}
	tt.Move(500, 400)
	tt.Key(0, KeyTab)
	tt.Advance(299 * time.Millisecond)
	if tt.HasText(f.label) {
		t.Fatal("keyboard hint appeared before 300 ms")
	}
	tt.Advance(time.Millisecond)
	if !tt.HasText(f.label) {
		t.Fatal("keyboard hint did not appear")
	}
	f.label = "Updated hint"
	tt.Frame()
	if !tt.HasText(f.label) {
		t.Fatal("visible hint did not follow its changed label")
	}
	f.hover(t, "Open")
	tt.Advance(300 * time.Millisecond)
	tt.Key(0, KeyEscape)
	tt.Advance(time.Second)
	if tt.HasText("Opens a file") || tt.HasText(f.label) || tt.h.escapes != 0 {
		t.Fatal("Escape either reached the parent or fell back to the older focused hint")
	}
	tt.Key(0, KeyEscape)
	if tt.h.escapes != 1 {
		t.Fatal("a second Escape did not reach the parent")
	}
	tt.Move(500, 400)
	tt.Advance(300 * time.Millisecond)
	f.disabled = true
	tt.Frame()
	if tt.HasText(f.label) {
		t.Fatal("disabled anchor retained its hint")
	}
	f.disabled = false
	f.hover(t, "Open")
	tt.Advance(300 * time.Millisecond)
	f.hover(t, "Save")
	tt.Advance(300 * time.Millisecond)
	f.removed = true
	tt.Frame()
	if tt.HasText(f.label) {
		t.Fatal("removed anchor retained its hint")
	}
}

func TestTooltipPolicyDismissalAndPlacement(t *testing.T) {
	t.Run("keyboard supersedes stationary pointer", func(t *testing.T) {
		f := newTooltipFixture(760)
		f.hover(t, "Save")
		f.tt.Advance(300 * time.Millisecond)
		f.tt.Key(0, KeyTab)
		f.tt.Key(0, KeyTab)
		f.tt.Advance(300 * time.Millisecond)
		if !f.tt.HasText("Opens a file") {
			t.Fatal("keyboard did not supersede the stationary pointer")
		}
		f.tt.Key(0, KeyTab)
		f.tt.Advance(time.Second)
		if f.tt.HasText(f.label) || f.tt.HasText("Opens a file") {
			t.Fatal("stationary pointer restored an old hint after keyboard focus left")
		}
	})
	t.Run("disabled after tooltip", func(t *testing.T) {
		for _, ancestor := range []bool{false, true} {
			disabled := false
			tt := coreNewTester(func(c *context) {
				theme := LightTheme()
				theme.TooltipPolicy = &TooltipPolicy{Delay: 300 * time.Millisecond}
				c.SetTheme(theme)
				row := coreRow(c).Children(func() {
					button := coreButton(c, "Save").Tooltip("Save hint")
					if !ancestor {
						button.Disabled(disabled)
					}
				})
				if ancestor {
					row.Disabled(disabled)
				}
			}, 760, 520)
			r := at(t, tt, "Save")
			tt.Move(r.X+r.W/2, r.Y+r.H/2)
			tt.Advance(300 * time.Millisecond)
			if !tt.HasText("Save hint") {
				t.Fatal("fixture did not show its hint")
			}
			disabled = true
			tt.Frame()
			if tt.HasText("Save hint") {
				t.Fatal("disabling after building retained the hint")
			}
		}
	})
	for name, event := range map[string]platform.SurfaceEvent{
		"escape pending": {Kind: platform.KeyPressed, Key: platform.KeyEscape},
		"enter":          {Kind: platform.KeyPressed, Key: platform.KeyEnter},
		"space":          {Kind: platform.KeyPressed, Key: platform.KeySpace},
		"resize":         {Kind: platform.SurfaceResize},
		"blur":           {Kind: platform.SurfaceBlur},
		"scroll":         {Kind: platform.PointerScroll, X: 10, Y: 10, DY: 20},
		"pointer":        {Kind: platform.PointerDown, X: 10, Y: 10},
	} {
		t.Run(name, func(t *testing.T) {
			f := newTooltipFixture(760)
			f.hover(t, "Save")
			if name != "escape pending" {
				f.tt.Advance(300 * time.Millisecond)
				if !f.tt.HasText(f.label) {
					t.Fatal("fixture did not show a hint")
				}
			}
			f.tt.send(event)
			if name == "blur" {
				f.tt.send(platform.SurfaceEvent{Kind: platform.SurfaceFocus})
			}
			f.tt.Advance(time.Second)
			if f.tt.HasText(f.label) {
				t.Fatal("dismissed hint returned without a new engagement")
			}
		})
	}
	t.Run("programmatic scroll", func(t *testing.T) {
		f := newTooltipFixture(760)
		f.hover(t, "Save")
		f.tt.Advance(300 * time.Millisecond)
		f.scroll.Y = 40
		f.tt.Frame()
		f.tt.Frame()
		if f.tt.HasText(f.label) {
			t.Fatal("programmatic scrolling did not dismiss the hint")
		}
	})
	for _, width := range []int{240, 760, 1320} {
		f := newTooltipFixture(width)
		f.label = strings.Repeat("Long hint ", 12)
		f.tt.Frame()
		f.hover(t, "Save")
		f.tt.Advance(300 * time.Millisecond)
		r := at(t, f.tt, "tooltip box")
		if r.X < 8 || r.X+r.W > float32(width)-8 || r.W > 320 {
			t.Fatalf("hint exceeds the viewport at width %d: %v", width, r)
		}
		bottom := f.hover(t, "Bottom")
		f.tt.Advance(300 * time.Millisecond)
		r = at(t, f.tt, "tooltip box")
		if r.Y+r.H != bottom.Y-4 {
			t.Fatalf("hint did not flip above a bottom anchor: %v by %v", r, bottom)
		}
	}
}
