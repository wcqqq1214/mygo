package ui

import (
	"testing"

	"github.com/egoist/mygo/internal/platform"
)

func TestDisabledColorsKeepInputAndAccessibilitySemantics(t *testing.T) {
	for _, explicitColors := range []bool{false, true} {
		clicks := 0
		var paragraph *node
		tt := coreNewTester(func(c *context) {
			theme := LightTheme()
			theme.DisableAutoDimming = explicitColors
			c.SetTheme(theme)
			coreColumn(c).Padding(10).Gap(4).Children(func() {
				if coreButtonBase(c).Label("Disabled").Size(30, 20).Background(RGB(0, 0, 0)).Disabled(true).Clicked() {
					clicks++
				}
				paragraph = coreRichText(c).Children(func() {
					coreText(c, "Inline").TextColor(RGB(0, 0, 0)).Disabled(true)
				})
				coreButton(c, "Enabled")
			})
		}, 140, 140)
		pixel := tt.Image().RGBAAt(15, 15)
		if explicitColors && pixel.R != 0 || !explicitColors && pixel.R == 0 {
			t.Fatalf("disabled block opacity: explicit=%v, pixel=%v", explicitColors, pixel)
		}
		if len(paragraph.spans) != 1 || (paragraph.spans[0].Color.A == 255) != explicitColors {
			t.Fatalf("disabled inline opacity: explicit=%v, spans=%+v", explicitColors, paragraph.spans)
		}
		tt.Click("Disabled")
		if clicks != 0 || tt.Focused("Disabled") {
			t.Fatal("disabled button accepted input")
		}
		tt.Key(0, KeyTab)
		if !tt.Focused("Enabled") {
			t.Fatal("Tab did not skip disabled button")
		}
		tt.send(platform.SurfaceEvent{Kind: platform.AccessibilityOn})
		n := accessNode(t, tt.h.access, platform.RoleButton, "Disabled")
		if n.States&platform.AccessDisabled == 0 {
			t.Fatal("disabled button lost accessibility state")
		}
		tt.send(platform.SurfaceEvent{Kind: platform.AccessAction, ID: n.ID, Action: platform.AccessPress})
		if clicks != 0 {
			t.Fatal("accessibility activated disabled button")
		}
	}
}
