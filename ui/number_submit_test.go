package ui

import "testing"

func TestNumberInputReportsSubmitOnlyForEnter(t *testing.T) {
	value, submits, changes := 5.0, 0, 0
	tt := NewTester(func(c *Context) {
		Column(c).Padding(10).AlignItems(Start).Children(func() {
			NumberInput(c, &value, 0, 10, 1).Label("Count").OnSubmit(func() { submits++ }).OnChange(func() { changes++ })
		})
	}, 400, 100)
	tt.Frame()
	field, _ := tt.Find("Count")
	tt.ClickAt(field.X+20, field.Y+field.H/2)
	tt.Key(Cmd, KeyA)
	tt.Type("7")
	tt.Key(0, KeyUp)
	if submits != 0 || value != 8 {
		t.Fatalf("typing and stepping submitted %d times, value %v", submits, value)
	}
	if err := tt.Click("Increase"); err != nil {
		t.Fatal(err)
	}
	if submits != 0 || value != 9 {
		t.Fatalf("button submitted %d times, value %v", submits, value)
	}
	tt.ClickAt(field.X+20, field.Y+field.H/2)
	tt.Key(0, KeyEnter)
	if submits != 1 {
		t.Fatalf("Enter submitted %d times", submits)
	}
	tt.Frame()
	if submits != 1 {
		t.Fatalf("a later frame repeated the submission: %d", submits)
	}
	tt.Key(0, KeyEnter)
	if submits != 2 {
		t.Fatalf("a second Enter submitted %d times in total", submits)
	}
}

func TestNumberInputFillsAnExplicitWidthAndKeepsItsNaturalWidth(t *testing.T) {
	layout := func(width float32) (field, input, increase Rect) {
		value := 5.0
		tt := NewTester(func(c *Context) {
			Column(c).Padding(10).AlignItems(Start).Children(func() {
				number := NumberInput(c, &value, 0, 10, 1).Label("Count")
				if width > 0 {
					number.Width(width)
				}
			})
		}, 400, 100)
		tt.Frame()
		field, _ = tt.Find("Count")
		decrease, _ := tt.Find("Decrease")
		increase, _ = tt.Find("Increase")
		return field, Rect{X: field.X, W: decrease.X - field.X}, increase
	}
	field, input, increase := layout(0)
	if edge := increase.X + increase.W; edge < field.X+field.W-0.5 || edge > field.X+field.W+0.5 || input.W < 80 || input.W > 90 {
		t.Fatalf("natural layout changed: field=%v input=%v increase=%v", field, input, increase)
	}
	field, input, increase = layout(248)
	if field.W != 248 || increase.X+increase.W < field.X+field.W-0.5 || increase.X+increase.W > field.X+field.W+0.5 || input.W <= 90 {
		t.Fatalf("explicit width not filled: field=%v input=%v increase=%v", field, input, increase)
	}
}
