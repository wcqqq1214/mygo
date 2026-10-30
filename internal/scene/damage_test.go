package scene

import (
	"image"
	"slices"
	"testing"
)

// TestDamageOfNestedTransformedClip checks that a clip under a
// transform, nested in a rounded one, changes all the rounded one cuts as
// it comes and goes: the GPU's renderers then cut the outer clip with a
// mask rather than in the shader, at its edges a unit in the last place
// apart.
func TestDamageOfNestedTransformedClip(t *testing.T) {
	outer := Op{Kind: OpPushClip, Rect: Rect{X: 10, Y: 10, W: 150, H: 150}, Radii: Uniform(8)}
	corner := Op{Kind: OpFill, Rect: Rect{X: 10, Y: 10, W: 20, H: 20}, Color: Color{R: 255, A: 255}}
	nested := []Op{
		{Kind: OpPushClip, Rect: Rect{W: 40, H: 40}, Transform: Translation(100, 100)},
		{Kind: OpFill, Rect: Rect{X: 100, Y: 100, W: 20, H: 20}, Color: Color{G: 255, A: 255}},
		{Kind: OpPopClip},
	}
	without := &Scene{Width: 200, Height: 200, Ops: []Op{outer, corner, {Kind: OpPopClip}}}
	with := &Scene{Width: 200, Height: 200, Ops: append(append([]Op{outer, corner}, nested...), Op{Kind: OpPopClip})}
	for _, frames := range [][2]*Scene{{without, with}, {with, without}} {
		d := Differ{Reach: 2}
		d.Diff(frames[0])
		d.Remember(frames[0])
		damage, ok := d.Diff(frames[1])
		if !ok {
			t.Fatal("no damage found")
		}
		// The fill in the outer clip's corner, which the inner clip
		// does not reach.
		at := image.Rect(10, 10, 30, 30)
		hit := false
		for _, r := range damage {
			hit = hit || at.In(r)
		}
		if !hit {
			t.Errorf("%d to %d ops: damage %v misses %v", len(frames[0].Ops), len(frames[1].Ops), damage, at)
		}
	}
}

func damageClipScene(groups int, transformed bool) *Scene {
	s := &Scene{Width: 200, Height: 200}
	for range groups {
		fill := Op{Kind: OpFill, Rect: Rect{X: 50, Y: 50, W: 10, H: 10}, Color: Color{G: 255, A: 255}}
		if transformed {
			fill.Transform = Scaling(2, 2)
		}
		s.Ops = append(s.Ops,
			Op{Kind: OpPushClip, Rect: Rect{W: 160, H: 160}, Radii: Uniform(8)},
			Op{Kind: OpFill, Rect: Rect{X: 10, Y: 10, W: 20, H: 20}, Color: Color{R: 255, A: 255}},
			fill, Op{Kind: OpPopClip})
	}
	return s
}

func TestDamageEnclosingFlagsReuse(t *testing.T) {
	var d Differ
	var previous *Scene
	for frame, tc := range []struct {
		groups      int
		transformed bool
	}{
		{0, false}, {1, true}, {16, true}, {16, true}, {16, false},
		{16, true}, {16, true}, {1, false}, {1, false},
		{16, false}, {16, false}, {0, false}, {0, false}, {16, false},
	} {
		s := damageClipScene(tc.groups, tc.transformed)
		capacity := cap(d.encloses2)
		var backing *bool
		if capacity > 0 {
			backing = &d.encloses2[:capacity][0]
		}
		got, ok := d.Diff(s)
		// A fresh differ has no flags left over from earlier frames.
		var fresh Differ
		if previous != nil {
			fresh.Diff(previous)
			fresh.Remember(previous)
		}
		want, wantOK := fresh.Diff(s)
		if ok != wantOK || !slices.Equal(got, want) {
			t.Fatalf("frame %d: reused flags changed damage: %v, %v; want %v, %v", frame, got, ok, want, wantOK)
		}
		flags := make([]bool, len(s.Ops))
		for i := 0; i < len(flags); i += 4 {
			flags[i] = tc.transformed
		}
		if !slices.Equal(d.encloses2, flags) {
			t.Fatalf("frame %d: enclosing flags %v, want %v", frame, d.encloses2, flags)
		}
		if len(s.Ops) <= capacity {
			if cap(d.encloses2) != capacity || capacity > 0 && &d.encloses2[:capacity][0] != backing {
				t.Fatalf("frame %d: flags did not reuse sufficient capacity", frame)
			}
		}
		d.Remember(s)
		if !slices.Equal(d.encloses, flags) {
			t.Fatalf("frame %d: Remember did not preserve the new enclosing flags", frame)
		}
		previous = s
	}
}

func TestDamageSteadyAllocs(t *testing.T) {
	s := damageClipScene(64, true)
	var d Differ
	frame := func() {
		d.Diff(s)
		d.Remember(s)
	}
	// Both flag buffers, swapped by Remember, have enough capacity.
	frame()
	frame()
	if allocs := testing.AllocsPerRun(100, frame); allocs != 0 {
		t.Fatalf("a steady damage comparison allocates %v times, want 0", allocs)
	}
}
