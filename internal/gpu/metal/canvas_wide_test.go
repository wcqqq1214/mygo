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

// Shared descendants must propagate their current gamut through every ancestor,
// even if only the leaf changes version. Output changes also invalidate formats.
func TestCanvasWideSharedDescendants(t *testing.T) {
	r, err := newRenderer()
	if err != nil {
		unavailable(t, err)
	}
	defer r.Release()
	if !r.SetWide(true) {
		t.Fatal(r.wideErr)
	}
	leaf := scene.NewCanvas(16, 8)
	fill := scene.Op{Kind: scene.OpFill, Rect: scene.Rect{W: 16, H: 8}, Color: scene.Color{R: 30, G: 230, B: 20, A: 255}}
	leaf.Scene.Ops = []scene.Op{fill}
	imageOp := func(img *scene.Image, x float32) scene.Op {
		return scene.Op{Kind: scene.OpImage, Image: img, Rect: scene.Rect{X: x, W: 16, H: 8}, Src: scene.Rect{W: 16, H: 8}}
	}
	canvases := []*scene.Image{leaf}
	parent := leaf
	for range 16 {
		next := scene.NewCanvas(16, 8)
		next.Scene.Ops = []scene.Op{imageOp(parent, 0)}
		canvases = append(canvases, next)
		parent = next
	}
	other := scene.NewCanvas(16, 8)
	other.Scene.Ops = []scene.Op{imageOp(leaf, 0)}
	canvases = append(canvases, other)
	s := &scene.Scene{Width: 32, Height: 8, Ops: []scene.Op{imageOp(parent, 0), imageOp(other, 16)}}
	for _, contentWide := range []bool{true, false, true} {
		leaf.Scene.Wide, leaf.Scene.Ops[0].Wide = nil, 0
		if contentWide {
			leaf.Scene.Wide = []scene.WideColors{{Color: [4]float32{-.25, 1.1, -.1, 1}, Set: scene.WideColor}}
			leaf.Scene.Ops[0].Wide = 1
		}
		leaf.Changed() // Ancestors retain their identities and versions.
		direct := &scene.Scene{Width: 32, Height: 8, Wide: leaf.Scene.Wide, Ops: []scene.Op{leaf.Scene.Ops[0]}}
		direct.Ops[0].Rect.W = 32
		for _, outputWide := range []bool{true, false, true} {
			want := renderRaw(t, r, direct, outputWide)
			for range 2 {
				got := renderRaw(t, r, s, outputWide)
				if !outputWide {
					if !bytes.Equal(got, want) {
						t.Fatal("shared canvases differ from direct sRGB drawing")
					}
				} else {
					for i := 0; i < len(got); i += 2 {
						a, b := half(binary.LittleEndian.Uint16(got[i:])), half(binary.LittleEndian.Uint16(want[i:]))
						if math.Abs(float64(a-b)) > .002 {
							t.Fatalf("shared canvas pixel: got %g want %g", a, b)
						}
					}
				}
				format := uint(pixelFormatBGRA8Unorm)
				if outputWide && contentWide {
					format = pixelFormatRGBA16Float
				}
				for _, c := range canvases {
					if got := r.canvases[c.ID()].format; got != format {
						t.Fatalf("canvas format: got %d want %d", got, format)
					}
				}
			}
		}
	}
}

// Effect inputs participate too. A cycle has no child-first order, so format
// selection must not trust a texture left over from the preceding frame.
func TestCanvasWideEffectInputsAndCycles(t *testing.T) {
	r, err := newRenderer()
	if err != nil {
		unavailable(t, err)
	}
	defer r.Release()
	if !r.SetWide(true) {
		t.Fatal(r.wideErr)
	}
	a, b := scene.NewCanvas(16, 8), scene.NewCanvas(16, 8)
	a.Scene.Effects = []scene.EffectOp{{Images: [4]*scene.Image{b}}}
	root := &scene.Scene{Width: 16, Height: 8, Canvases: []*scene.Image{a}}
	for _, cycle := range []bool{false, true} {
		if cycle {
			b.Scene.Effects = []scene.EffectOp{{Images: [4]*scene.Image{a}}}
		}
		for _, wide := range []bool{true, false, true} {
			a.Scene.Wide, b.Scene.Wide = nil, nil
			if wide {
				// In the cycle, the wide node is the not-yet-prepared ancestor.
				owner := b
				if cycle {
					owner = a
				}
				owner.Scene.Wide = []scene.WideColors{{Set: scene.WideColor}}
			}
			a.Changed()
			b.Changed()
			renderRaw(t, r, root, true)
			want := uint(pixelFormatBGRA8Unorm)
			if wide {
				want = pixelFormatRGBA16Float
			}
			for _, c := range []*scene.Image{a, b} {
				if got := r.canvases[c.ID()].format; got != want {
					t.Fatalf("cycle %v, wide %v: format %d want %d", cycle, wide, got, want)
				}
			}
		}
	}
}
