package ui

import (
	"bytes"
	"testing"
	"time"
)

func TestInertWebViewRemainsVisibleAndSkipsInput(t *testing.T) {
	page := &testWebView{"image"}
	inert, shown := false, true
	var before, after string
	tt := coreNewTester(func(c *context) {
		coreColumn(c).Fill().Children(func() {
			coreTextInput(c, &before).Label("Before")
			if shown {
				coreWebView(c, page).Height(80).Inert(inert).Background(RGB(20, 80, 130))
			}
			coreTextInput(c, &after).Label("After")
		})
	}, 200, 180)
	normal := tt.Image()
	inert = true
	tt.Frame()
	if !bytes.Equal(normal.Pix, tt.Image().Pix) {
		t.Fatal("inert content changed appearance")
	}
	p, ok := placed(tt, page)
	if !ok || len(p.covers) != 1 || p.covers[0] != p.clip {
		t.Fatal("inert web view is hidden or can receive pointer input")
	}
	tt.Click("Before")
	tt.Key(0, KeyTab)
	if !tt.Focused("After") || tt.h.tabbedInto == page {
		t.Fatal("Tab entered an inert web view")
	}
	tt.Key(Shift, KeyTab)
	if !tt.Focused("Before") {
		t.Fatal("reverse Tab entered an inert web view")
	}
	shown = false
	tt.Frame()
	tt.Advance(time.Second)
	shown, inert = true, false
	tt.Frame()
	tt.Click("Before")
	tt.Key(0, KeyTab)
	if tt.h.tabbedInto != page {
		t.Fatal("returning interactive web view did not recover input")
	}
	inert = true
	tt.Frame()
	if s := tt.rt.states[tt.rt.focused]; s != nil && s.webView == page {
		t.Fatal("making the focused page inert left it with keyboard focus")
	}
}

type retainedWideEffect struct{ bitmap *Bitmap }

func (*retainedWideEffect) EffectMargin() Point { return Point{} }
func (fx *retainedWideEffect) PaintEffect(p *Painter, r Rect, b *Bitmap, _ Point) {
	fx.bitmap = b
	p.Image(b, r, FillBox)
}

func TestEffectAndRetainedCanvasRequestWideOutput(t *testing.T) {
	for _, mode := range []string{"direct", "effect", "nested", "retained"} {
		t.Run(mode, func(t *testing.T) {
			g := &wideGPU{can: true}
			h, _, frame := gpuHost(t, func() (gpuRenderer, error) { return g, nil })
			h.conn.Surface = &wideSurface{wide: true}
			fx, outer := &retainedWideEffect{}, &retainedWideEffect{}
			held := false
			h.rt = newRuntime(func(c *context) {
				if held {
					coreBox(c).Size(24, 16).Draw(func(p *Painter, r Rect) { p.Image(fx.bitmap, r, FillBox) })
					return
				}
				parent := coreBox(c).Size(24, 16)
				if mode == "nested" {
					parent.Effect(outer)
				}
				parent.Children(func() {
					e := coreBox(c).Size(24, 16).Background(wideGreen)
					if mode != "direct" {
						e.Effect(fx)
					}
				})
			}, h)
			frame()
			if !g.wide {
				t.Fatal("wide colors in an effect did not request wide output")
			}
			if mode == "retained" {
				held = true
				h.wideUntil, h.wide, g.wide = time.Time{}, false, false
				frame()
				if !g.wide {
					t.Fatal("retained canvas lost its gamut request")
				}
			}
		})
	}
}
