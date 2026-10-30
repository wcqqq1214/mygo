package ui

import (
	"fmt"
	"testing"
	"time"
)

func TestColorReversalMatchesCSS(t *testing.T) {
	const duration = 250 * time.Millisecond
	for _, shorten := range []bool{false, true} {
		for _, repeated := range []bool{false, true} {
			t.Run(fmt.Sprintf("shorten=%v/repeated=%v", shorten, repeated), func(t *testing.T) {
				a, b := RGB(223, 227, 225), RGB(240, 244, 241)
				bg := a
				var element *node
				tt := coreNewTester(func(c *context) {
					element = coreBox(c).Size(40, 40).Background(bg).Transition(ElementTransition{
						Colors: true, Duration: duration, Ease: CubicBezier(0, 0, 0.58, 1), ShortenColorReversals: shorten,
					})
				}, 40, 40)
				start := tt.Image().RGBAAt(20, 20)
				currentDuration := func() time.Duration {
					if shorten {
						return element.st.trec.colorCSS[0].k.duration
					}
					return element.st.trec.colorK.duration
				}
				bg = b
				tt.Frame()
				tt.Advance(100 * time.Millisecond)
				tt.Frame()
				shown := tt.Image().RGBAAt(20, 20)
				bg = a
				tt.Frame()
				if tt.Image().RGBAAt(20, 20) != shown {
					t.Fatal("reversal jumped from its current color")
				}
				// CSS easing progress at 100/250 is 0.57088023. Actual
				// WKWebView reports 142.720 ms and 192.212 ms for this sequence.
				want := duration
				if shorten {
					want = 142720 * time.Microsecond
				}
				if got := currentDuration(); got < want-200*time.Microsecond || got > want+200*time.Microsecond {
					t.Fatalf("first reversal duration = %v, want %v", got, want)
				}
				if repeated {
					tt.Advance(60 * time.Millisecond)
					tt.Frame()
					shown = tt.Image().RGBAAt(20, 20)
					bg = b
					tt.Frame()
					if tt.Image().RGBAAt(20, 20) != shown {
						t.Fatal("second reversal jumped")
					}
					if shorten {
						want = 192212 * time.Microsecond
					}
					if got := currentDuration(); got < want-200*time.Microsecond || got > want+200*time.Microsecond {
						t.Fatalf("second reversal duration = %v, want %v", got, want)
					}
				}
				tt.Advance(want + time.Millisecond)
				tt.Frame()
				if element.st.trec.colorMoving {
					t.Fatal("transition outlived its expected end")
				}
				if !repeated && tt.Image().RGBAAt(20, 20) != start {
					t.Fatal("reverse did not restore the original pixels")
				}
				// Once settled, an ordinary new transition takes its full time.
				bg = RGB(90, 125, 110)
				tt.Frame()
				if currentDuration() != duration {
					t.Fatal("a new target inherited a completed reversal's shortening")
				}
			})
		}
	}
}

func TestCSSColorTransitionsKeepIndependentDeadlines(t *testing.T) {
	bg, border := RGB(223, 227, 225), RGB(184, 192, 188)
	var element *node
	tt := coreNewTester(func(c *context) {
		element = coreBox(c).Size(40, 40).Background(bg).Border(1, border).Transition(ElementTransition{
			Colors: true, Duration: 250 * time.Millisecond, Ease: CubicBezier(0, 0, 0.58, 1), ShortenColorReversals: true,
		})
	}, 40, 40)
	bg, border = RGB(240, 244, 241), RGB(148, 163, 156)
	tt.Frame()
	tt.Advance(100 * time.Millisecond)
	tt.Frame()
	// Pressing changes the fill but leaves the hovered border target alone.
	bg = RGB(232, 237, 234)
	tt.Frame()
	if element.st.trec.colorCSS[0].k.duration != 250*time.Millisecond {
		t.Fatal("a third target inherited reversal shortening")
	}
	tt.Advance(151 * time.Millisecond)
	tt.Frame()
	if element.borderC != border || element.st.trec.colorCSS[1].k.moving {
		t.Fatal("changing the background restarted the border transition")
	}
	if !element.st.trec.colorCSS[0].k.moving {
		t.Fatal("background did not retain its own later deadline")
	}
	tt.Advance(100 * time.Millisecond)
	tt.Frame()
	if element.st.trec.colorMoving {
		t.Fatal("independent color transitions did not settle")
	}
}
