package text

import (
	"strings"
	"testing"
	"unsafe"
)

func actualLayoutMemory(l *Layout) int {
	p := l.Params
	n := int(unsafe.Sizeof(cached{})) + len(p.Text) + len(p.Spans) + len(p.Ellipsis) + len(p.Style.Family) + len(p.Style.Features) + len(p.Style.Lang)
	n += cap(l.Runes) * int(unsafe.Sizeof(rune(0)))
	if cap(l.Lines) > 1 {
		n += cap(l.Lines) * int(unsafe.Sizeof(Line{}))
	}
	n += cap(l.boundaries) * int(unsafe.Sizeof(int(0)))
	for i := range l.Lines {
		line := &l.Lines[i]
		n += cap(line.Glyphs) * int(unsafe.Sizeof(Glyph{}))
		n += (cap(line.carets) + cap(line.upstream)) * int(unsafe.Sizeof(float32(0)))
		n += cap(line.stops) * int(unsafe.Sizeof(caretStop{}))
	}
	return n
}

func TestLayoutMemoryCostReservesLazyGeometry(t *testing.T) {
	for _, source := range []string{"", "a", "ab", "abc", "abcde", "אבג one", "中文🙂", strings.Repeat("one אבג 中文🙂 ", 100)} {
		for _, width := range []float32{0, 80} {
			layout := Shared().Shape(Params{Text: source, Style: Style{Size: 14}, Width: width})
			before := layout.MemoryCost()
			for i := range layout.Lines {
				layout.visualStops(i)
			}
			if actual := actualLayoutMemory(layout); actual > before {
				t.Fatalf("text %q width %v actual %d > reserved %d", source, width, actual, before)
			}
			if layout.MemoryCost() != before {
				t.Fatal("lazy geometry changed admission estimate")
			}
		}
	}
}
