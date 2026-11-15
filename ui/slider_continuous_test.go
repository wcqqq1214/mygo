package ui

import (
	"math"
	"testing"

	"github.com/egoist/mygo/internal/platform"
)

func TestSliderContinuousKeyboard(t *testing.T) {
	for _, styled := range []bool{false, true} {
		name := "base"
		if styled {
			name = "styled"
		}
		t.Run(name, func(t *testing.T) {
			value := 25.25
			tt := coreNewTester(func(c *context) {
				if styled {
					coreSlider(c, &value, 0, 100).Width(220).Label("slider")
				} else {
					coreSliderBase(c, &value, 0, 100).Size(200, 100).Label("slider")
				}
			}, 240, 160)
			tt.send(platform.SurfaceEvent{Kind: platform.AccessibilityOn})
			n := accessNode(t, tt.h.access, platform.RoleSlider, "slider")
			tt.send(platform.SurfaceEvent{Kind: platform.AccessAction, ID: n.ID, Action: platform.AccessFocus})
			for _, tc := range []struct {
				key  Key
				want float64
			}{{KeyRight, 26.25}, {KeyRight, 27.25}, {KeyLeft, 26.25}, {KeyUp, 27.25}, {KeyDown, 26.25}, {KeyPageUp, 36.25}, {KeyPageDown, 26.25}, {KeyHome, 0}, {KeyEnd, 100}} {
				tt.Key(0, tc.key)
				tt.Frame()
				if value != tc.want {
					t.Fatalf("key %v: value %v, want %v", tc.key, value, tc.want)
				}
			}
			value = 25.25
			tt.Frame()
			for _, tc := range []struct {
				action platform.AccessActionKind
				want   float64
			}{{platform.AccessIncrement, 26.25}, {platform.AccessDecrement, 25.25}} {
				tt.send(platform.SurfaceEvent{Kind: platform.AccessAction, ID: n.ID, Action: tc.action})
				if value != tc.want || accessNode(t, tt.h.access, platform.RoleSlider, "slider").Now != tc.want {
					t.Fatalf("access action %v: value %v, want %v", tc.action, value, tc.want)
				}
			}
		})
	}
}

func TestSliderHeldKeyboardSurvivesFrames(t *testing.T) {
	value := 0.0
	tt := coreNewTester(func(c *context) {
		coreSliderBase(c, &value, 0, 100).Size(200, 100).Label("slider")
	}, 240, 160)
	check := func(want float64) {
		t.Helper()
		if math.Abs(value-want) > 0.0001 {
			t.Fatalf("value %v, want %v", value, want)
		}
	}
	tt.Press(50, 50)
	check(25)
	tt.Key(0, KeyRight)
	check(26)
	tt.Frame()
	check(26)
	tt.Key(0, KeyPageUp)
	check(36)
	tt.Release(50, 50)
	check(36)
	tt.Press(50, 50)
	check(25)
	tt.Key(0, KeyRight)
	tt.Move(150.5, 50)
	check(75.25)
	tt.Key(0, KeyLeft)
	check(74.25)
	tt.Move(-20, 50)
	check(0)
	tt.Key(0, KeyRight)
	check(1)
	tt.Frame()
	check(1)
	tt.CancelPress()
	tt.Move(180, 50)
	check(1)
}

func TestSliderHeldInputOptions(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		vertical, rtl, styled bool
		step                  float64
	}{
		{name: "vertical", vertical: true},
		{name: "rtl", rtl: true},
		{name: "vertical-rtl", vertical: true, rtl: true},
		{name: "stepped", step: 5},
		{name: "styled", styled: true},
		{name: "step-slider", styled: true, step: 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value := 0.0
			disabled := false
			tt := coreNewTester(func(c *context) {
				var e *node
				switch {
				case tc.styled && tc.step > 0:
					e = coreStepSlider(c, &value, 0, 100, tc.step).Width(220)
				case tc.styled:
					e = coreSlider(c, &value, 0, 100).Width(220)
				default:
					e = coreSliderBase(c, &value, 0, 100).Size(200, 100).Step(tc.step)
				}
				if tc.vertical {
					e.Vertical()
				}
				e.Disabled(disabled).Label("slider")
			}, 240, 160)
			if tc.rtl {
				tt.SetDirection(RightToLeft)
			}
			tt.send(platform.SurfaceEvent{Kind: platform.AccessibilityOn})
			n := accessNode(t, tt.h.access, platform.RoleSlider, "slider")
			x, y := float32(n.Bounds.X+n.Bounds.W/2), float32(n.Bounds.Y+n.Bounds.H/2)
			tt.Press(x, y)
			check := func(want float64) {
				t.Helper()
				if math.Abs(value-want) > .0001 {
					t.Fatalf("value %v, want %v", value, want)
				}
			}
			check(50)
			increment := max(1, tc.step)
			tt.Key(0, KeyUp)
			tt.Frame()
			check(50 + increment)
			tt.send(platform.SurfaceEvent{Kind: platform.AccessAction, ID: n.ID, Action: platform.AccessIncrement})
			check(50 + 2*increment)
			tt.send(platform.SurfaceEvent{Kind: platform.AccessAction, ID: n.ID, Action: platform.AccessDecrement})
			check(50 + increment)
			disabled = true
			tt.Frame()
			tt.Move(x+10, y-10)
			tt.Key(0, KeyUp)
			check(50 + increment)
			tt.Release(x+10, y-10)
		})
	}
}
