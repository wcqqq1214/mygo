package ui

import "testing"

func TestNumberInputColumnKeepsItsNaturalInputHeight(t *testing.T) {
	for _, align := range []Align{Start, Stretch} {
		value, text := 5.0, "5"
		var gap float32
		tt := NewTester(func(c *Context) {
			gap = c.Theme().Space(1)
			Column(c).Padding(10).AlignItems(align).Children(func() {
				TextInput(c, &text).Width(c.Theme().Space(20)).Label("Reference")
				NumberInput(c, &value, 0, 10, 1).Column().Label("Count")
			})
		}, 400, 300)
		tt.Frame()
		reference, _ := tt.Find("Reference")
		field, _ := tt.Find("Count")
		decrease, _ := tt.Find("Decrease")
		increase, _ := tt.Find("Increase")
		if height := decrease.Y - field.Y - gap; height != reference.H {
			t.Fatalf("column input with alignment %v has height %v, want %v", align, height, reference.H)
		}
		if want := reference.H + decrease.H + increase.H + 2*gap; field.H != want {
			t.Fatalf("column with alignment %v has height %v, want %v", align, field.H, want)
		}
	}
}
