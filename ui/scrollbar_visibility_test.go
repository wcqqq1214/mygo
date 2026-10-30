package ui

import (
	"fmt"
	"testing"
)

func TestScrollbarVisibilityAxes(t *testing.T) {
	for _, horizontal := range []bool{false, true} {
		for _, vertical := range []bool{false, true} {
			t.Run(fmt.Sprint(horizontal, vertical), func(t *testing.T) {
				var sc *node
				var state ScrollState
				tt := coreNewTester(func(c *context) {
					sc = coreScrollBoth(c).Size(200, 200).Scrollbars(horizontal, vertical).TrackScroll(&state).Children(func() {
						coreBox(c).Size(1000, 1000).Cursor(CursorText)
					})
				}, 220, 220)
				before := tt.Image()
				tt.Move(100, 100)
				after := tt.Image()
				if painted := before.RGBAAt(194, 12) != after.RGBAAt(194, 12); painted != vertical {
					t.Fatalf("vertical painted=%v, want %v", painted, vertical)
				}
				if painted := before.RGBAAt(12, 194) != after.RGBAAt(12, 194); painted != horizontal {
					t.Fatalf("horizontal painted=%v, want %v", painted, horizontal)
				}
				tt.Move(194, 80)
				wantCursor := CursorText
				if vertical {
					wantCursor = CursorDefault
				}
				if tt.rt.cursor != wantCursor {
					t.Fatal("vertical cursor disagrees with visibility")
				}
				tt.Move(80, 194)
				wantCursor = CursorText
				if horizontal {
					wantCursor = CursorDefault
				}
				if tt.rt.cursor != wantCursor {
					t.Fatal("horizontal cursor disagrees with visibility")
				}
				tt.Press(194, 12)
				if dragging := tt.rt.scrollDrag.st == sc.st; dragging != vertical {
					t.Fatalf("vertical hit=%v, want %v", dragging, vertical)
				}
				tt.Move(194, 70)
				tt.Release(194, 70)
				if (state.Y > 0) != vertical {
					t.Fatal("vertical drag disagrees with visibility")
				}
				tt.Press(12, 194)
				if dragging := tt.rt.scrollDrag.st == sc.st; dragging != horizontal {
					t.Fatalf("horizontal hit=%v, want %v", dragging, horizontal)
				}
				tt.Move(70, 194)
				tt.Release(70, 194)
				if (state.X > 0) != horizontal {
					t.Fatal("horizontal drag disagrees with visibility")
				}
				state.X, state.Y = 0, 0
				tt.Frame()
				tt.Scroll(100, 100, 30, 40)
				if state.X != 30 || state.Y != 40 {
					t.Fatalf("wheel disabled: %+v", state)
				}
			})
		}
	}
}

func TestScrollbarVisibilityChangesDuringDrag(t *testing.T) {
	var state ScrollState
	show := true
	tt := coreNewTester(func(c *context) {
		coreScroll(c).Size(200, 200).Scrollbars(true, show).TrackScroll(&state).Children(func() {
			coreBox(c).Height(1000).Shrink(0)
		})
	}, 220, 220)
	tt.Move(100, 100)
	tt.Press(194, 12)
	if tt.rt.scrollDrag.st == nil {
		t.Fatal("visible scrollbar did not start drag")
	}
	show = false
	tt.Frame()
	tt.Move(194, 100)
	tt.Release(194, 100)
	if tt.rt.scrollDrag.st != nil || state.Y != 0 {
		t.Fatal("hidden scrollbar continued dragging")
	}
	show = true
	tt.Frame()
	tt.Press(194, 12)
	tt.Move(194, 100)
	tt.Release(194, 100)
	if state.Y <= 0 {
		t.Fatal("restored scrollbar did not drag")
	}
}
