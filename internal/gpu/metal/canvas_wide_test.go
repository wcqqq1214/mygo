//go:build darwin

package metal

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/egoist/mygo/internal/scene"
)

// Compare real float16 readback with direct drawing, including retained and
// nested canvases and a window moving between wide and sRGB output formats.
func TestCanvasPreservesWidePixelsAndSRGBFallback(t *testing.T) {
	r, err := newRenderer()
	if err != nil {
		unavailable(t, err)
	}
	defer r.Release()
	if !r.SetWide(true) {
		t.Fatal(r.wideErr)
	}
	direct := &scene.Scene{Width: 16, Height: 8}
	direct.Wide = []scene.WideColors{{Color: [4]float32{-.25, 1.1, -.1, 1}, Set: scene.WideColor}}
	// Deliberately different from per-channel clipping: the sRGB path must
	// use the supplied gamut-mapped fallback, as direct rendering does.
	direct.Ops = []scene.Op{{Kind: scene.OpFill, Rect: scene.Rect{W: 16, H: 8}, Color: scene.Color{R: 30, G: 230, B: 20, A: 255}, Wide: 1}}
	inner := scene.NewCanvas(16, 8)
	*inner.Scene = *direct
	wrap := func(img *scene.Image) *scene.Scene {
		return &scene.Scene{Width: 16, Height: 8, Ops: []scene.Op{{Kind: scene.OpImage, Rect: scene.Rect{W: 16, H: 8}, Image: img, Src: scene.Rect{W: 16, H: 8}}}}
	}
	nested := scene.NewCanvas(16, 8)
	*nested.Scene = *wrap(inner)
	for _, wide := range []bool{true, false, true} {
		want := renderRaw(t, r, direct, wide)
		for _, img := range []*scene.Image{inner, nested} {
			for range 2 { // Same version, sampled again: exercise retained textures.
				got := renderRaw(t, r, wrap(img), wide)
				if !wide {
					if !bytes.Equal(got, want) {
						t.Fatal("canvas differs from direct sRGB gamut mapping")
					}
					continue
				}
				for i := 0; i < len(got); i += 2 {
					actual, expected := half(binary.LittleEndian.Uint16(got[i:])), half(binary.LittleEndian.Uint16(want[i:]))
					if math.Abs(float64(actual-expected)) > .002 {
						t.Fatalf("canvas clipped wide component: got %g want %g", actual, expected)
					}
				}
			}
		}
	}
	// A same-sized canvas changes its content gamut without changing identity.
	direct.Wide, direct.Ops[0].Wide = nil, 0
	*inner.Scene = *direct
	inner.Changed()
	nested.Changed()
	want := renderRaw(t, r, direct, true)
	got := renderRaw(t, r, wrap(nested), true)
	if !bytes.Equal(got, want) {
		t.Fatal("canvas format cache did not follow content becoming sRGB")
	}
}
