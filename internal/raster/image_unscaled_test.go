package raster

import (
	"bytes"
	"fmt"
	"image"
	"math/rand/v2"
	"testing"

	"github.com/egoist/mygo/internal/scene"
)

func imageFixture(w, h int) *scene.Image {
	rnd := rand.New(rand.NewPCG(31, 71))
	pix := make([]byte, 4*w*h)
	for i := 0; i < len(pix); i += 4 {
		a := byte(rnd.IntN(256))
		if i%12 == 0 {
			a = 0
		} else if i%12 == 4 {
			a = 255
		}
		pix[i], pix[i+1], pix[i+2], pix[i+3] = byte(rnd.IntN(int(a)+1)), byte(rnd.IntN(int(a)+1)), byte(rnd.IntN(int(a)+1)), a
	}
	return scene.NewImageRGBA(w, h, pix)
}

// A neutral tint forces the general image sampler without changing its output.
// Compare entire frames, including pixels outside the clipped image.
func TestUnscaledImageMatchesSampler(t *testing.T) {
	img := imageFixture(32, 24)
	for _, offset := range []float32{-9, 0, 7, 7.25} {
		for _, opacity := range []float32{0, 1, .4} {
			for _, radius := range []float32{0, 6, -6} {
				for _, crop := range []scene.Rect{{W: 32, H: 24}, {X: 4, Y: 3, W: 20, H: 15}, {X: 24, Y: 3, W: -20, H: 15}, {X: 20, Y: 3, W: 20, H: 15}} {
					t.Run(fmt.Sprint(offset, opacity, radius, crop), func(t *testing.T) {
						s := &scene.Scene{Width: 40, Height: 35, Clear: scene.Color{R: 31, G: 72, B: 99, A: 193}}
						s.Ops = []scene.Op{
							{Kind: scene.OpPushClip, Rect: scene.Rect{X: 1.25, Y: 2.5, W: 34, H: 29}, Radii: [4]float32{radius, radius, radius, radius}},
							{Kind: scene.OpImage, Rect: scene.Rect{X: offset, Y: 4, W: abs32(crop.W), H: crop.H}, Src: crop, Image: img, Opacity: opacity},
							{Kind: scene.OpPopClip},
						}
						got, want := NewImage(40, 35), NewImage(40, 35)
						Render(got, s)
						s.Ops[1].Color = scene.Color{R: 255, G: 255, B: 255, A: 255}
						Render(want, s)
						if !bytes.Equal(got.Pix, want.Pix) {
							t.Fatal("unscaled pixels differ from general sampling")
						}
					})
				}
			}
		}
	}
}

func TestUnscaledCanvasDamageMatchesSampler(t *testing.T) {
	canvas := scene.NewCanvas(32, 24)
	canvas.Scene.Ops = []scene.Op{{Kind: scene.OpFill, Rect: scene.Rect{X: 1.5, Y: 2.5, W: 25, H: 20}, Color: scene.Color{R: 40, G: 95, B: 180, A: 139}, Radii: [4]float32{4, 4, 4, 4}}}
	op := scene.Op{Kind: scene.OpImage, Rect: scene.Rect{X: 3, Y: 5, W: 32, H: 24}, Src: scene.Rect{W: 32, H: 24}, Image: canvas}
	s := &scene.Scene{Width: 40, Height: 35, Ops: []scene.Op{op}, Canvases: []*scene.Image{canvas}}
	drawCanvases(s)
	for _, damage := range []image.Rectangle{image.Rect(0, 0, 40, 35), image.Rect(9, 10, 25, 20), image.Rect(38, 30, 40, 35)} {
		got, want := NewImage(40, 35), NewImage(40, 35)
		for i := range got.Pix {
			got.Pix[i], want.Pix[i] = 41, 41
		}
		var a, b Renderer
		s.Ops[0] = op
		a.d.draw(got, s, damage, a.opBounds(s))
		s.Ops[0].Color = scene.Color{R: 255, G: 255, B: 255, A: 255}
		b.d.draw(want, s, damage, b.opBounds(s))
		if !bytes.Equal(got.Pix, want.Pix) {
			t.Fatalf("canvas differs for damage %v", damage)
		}
	}
}

func BenchmarkUnscaledImage(b *testing.B) {
	img := imageFixture(760, 520)
	for _, sampled := range []bool{false, true} {
		b.Run(fmt.Sprint("sampled=", sampled), func(b *testing.B) {
			op := scene.Op{Kind: scene.OpImage, Rect: scene.Rect{W: 760, H: 520}, Src: scene.Rect{W: 760, H: 520}, Image: img}
			if sampled {
				op.Color = scene.Color{R: 255, G: 255, B: 255, A: 255}
			}
			s := &scene.Scene{Width: 760, Height: 520, Ops: []scene.Op{op}}
			m := NewImage(760, 520)
			b.ResetTimer()
			for b.Loop() {
				Render(m, s)
			}
		})
	}
}
