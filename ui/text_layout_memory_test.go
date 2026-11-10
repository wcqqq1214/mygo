package ui

import (
	"strings"
	"sync"
	"testing"
	"unsafe"
)

func TestTextLayoutMemoryCostIncludesOwnedBuffers(t *testing.T) {
	plain := ShapeText("abc", Font{Size: 14}, 0)
	long := ShapeText(strings.Repeat("A\t中文🙂 ", 100), Font{Size: 14}, 0)
	wrapped := ShapeText(strings.Repeat("A\t中文🙂 ", 100), Font{Size: 14}, 80)
	rich := ShapeRichText([]Span{{Text: "abc", Font: "system-ui", Features: "liga", Lang: "en"}}, Font{Size: 14}, 0)
	for _, layout := range []*TextLayout{plain, long, wrapped, rich} {
		own := int(unsafe.Sizeof(*layout)) + (cap(layout.units)+cap(layout.boundaries))*int(unsafe.Sizeof(int(0))) + cap(layout.spans)*int(unsafe.Sizeof(Span{}))
		if layout.MemoryCost() <= own+len(layout.l.Params.Text) {
			t.Fatalf("layout data missing from estimate: %d", layout.MemoryCost())
		}
		before := layout.MemoryCost()
		layout.Caret(1)
		layout.IndexAt(Point{X: 10, Y: 10})
		layout.SelectionRects(TextInputRange{Start: 0, End: 2})
		if layout.MemoryCost() != before {
			t.Fatal("geometry invalidated the reserved memory estimate")
		}
		if allocs := testing.AllocsPerRun(20, func() { layout.MemoryCost() }); allocs != 0 {
			t.Fatalf("cost estimation allocates %v", allocs)
		}
	}
	if long.MemoryCost() <= plain.MemoryCost() || wrapped.MemoryCost() <= long.MemoryCost() || rich.MemoryCost() <= plain.MemoryCost() {
		t.Fatalf("missing text/wrap/style costs: %d %d %d %d", plain.MemoryCost(), long.MemoryCost(), wrapped.MemoryCost(), rich.MemoryCost())
	}
}

func TestTextLayoutMemoryCostConcurrentGeometry(t *testing.T) {
	layout := ShapeText("one אבג 中文🙂 two", Font{Size: 14}, 80)
	expected := layout.MemoryCost()
	var group sync.WaitGroup
	for i := 0; i < 4; i++ {
		group.Go(func() {
			for j := 0; j < 50; j++ {
				if layout.MemoryCost() != expected {
					t.Error("unstable memory estimate")
				}
				layout.Caret(j % 10)
				layout.SelectionRects(TextInputRange{Start: 0, End: 10})
			}
		})
	}
	group.Wait()
}
