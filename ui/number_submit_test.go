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
	layout := func(width float32, align Align) (field, input, increase Rect) {
		value := 5.0
		tt := NewTester(func(c *Context) {
			Column(c).Padding(10).AlignItems(align).Children(func() {
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
	field, input, increase := layout(0, Start)
	if edge := increase.X + increase.W; edge < field.X+field.W-0.5 || edge > field.X+field.W+0.5 || input.W < 80 || input.W > 90 {
		t.Fatalf("natural layout changed: field=%v input=%v increase=%v", field, input, increase)
	}
	naturalInput, naturalIncrease := input.W, increase.X-field.X
	field, input, increase = layout(0, Stretch)
	if input.W != naturalInput || increase.X-field.X != naturalIncrease {
		t.Fatalf("stretch without Width changed the children: field=%v input=%v increase=%v; natural input=%v increase offset=%v", field, input, increase, naturalInput, naturalIncrease)
	}
	for _, align := range []Align{Start, Stretch} {
		field, input, increase = layout(248, align)
		if field.W != 248 || increase.X+increase.W < field.X+field.W-0.5 || increase.X+increase.W > field.X+field.W+0.5 || input.W <= 90 {
			t.Fatalf("explicit width not filled with alignment %v: field=%v input=%v increase=%v", align, field, input, increase)
		}
	}
}

func TestNumberInputWidthOptionsAndRemovingWidth(t *testing.T) {
	value := 5.0
	var size func(Element) Element
	tt := NewTester(func(c *Context) {
		Column(c).Padding(10).Children(func() {
			number := NumberInput(c, &value, 0, 10, 1).Label("Count")
			if size != nil {
				size(number)
			}
		})
	}, 400, 100)
	tt.Frame()
	naturalDecrease, _ := tt.Find("Decrease")
	naturalIncrease, _ := tt.Find("Increase")
	for _, option := range []struct {
		name  string
		apply func(Element) Element
		width float32
	}{
		{"Width", func(e Element) Element { return e.Width(248) }, 248},
		{"Size", func(e Element) Element { return e.Size(248, 34) }, 248},
		{"WidthPercent", func(e Element) Element { return e.WidthPercent(50) }, 190},
		{"FillWidth", func(e Element) Element { return e.FillWidth() }, 380},
		{"Fill", func(e Element) Element { return e.Fill() }, 380},
	} {
		t.Run(option.name, func(t *testing.T) {
			size = option.apply
			tt.Frame()
			field, _ := tt.Find("Count")
			increase, _ := tt.Find("Increase")
			if field.W != option.width || increase.X+increase.W != field.X+field.W {
				t.Fatalf("width option not filled: field=%v increase=%v", field, increase)
			}
			size = nil
			tt.Frame()
			decrease, _ := tt.Find("Decrease")
			increase, _ = tt.Find("Increase")
			if decrease != naturalDecrease || increase != naturalIncrease {
				t.Fatalf("removing width changed natural buttons: decrease=%v increase=%v; want %v and %v", decrease, increase, naturalDecrease, naturalIncrease)
			}
		})
	}
}

func numberText(t *testing.T, tt *Tester) string {
	t.Helper()
	for _, node := range tt.Accessibility() {
		if node.Role == RoleTextField {
			return node.Value
		}
	}
	t.Fatal("no number input")
	return ""
}

func TestNumberInputShowsTheValueAfterEnterAndWhenTheAppSetsIt(t *testing.T) {
	value := 5.0
	var clamp bool
	tt := NewTester(func(c *Context) {
		Column(c).Padding(10).AlignItems(Start).Children(func() {
			NumberInput(c, &value, 0, 1000, 1).Label("Count").OnSubmit(func() {
				if clamp {
					value = min(value, 10)
				}
			})
		})
	}, 400, 100)
	tt.Frame()
	field, _ := tt.Find("Count")
	tt.ClickAt(field.X+20, field.Y+field.H/2)
	tt.Key(Cmd, KeyA)
	tt.Type("abc")
	if got := numberText(t, tt); got != "abc" {
		t.Fatalf("typing replaced the text: %q", got)
	}
	tt.Key(0, KeyEnter)
	tt.Frame()
	if got := numberText(t, tt); got != "5" || value != 5 {
		t.Fatalf("Enter on text that is not a number shows %q with value %v", got, value)
	}
	clamp = true
	tt.Key(Cmd, KeyA)
	tt.Type("40")
	if got := numberText(t, tt); got != "40" || value != 40 {
		t.Fatalf("typed 40 shows %q with value %v", got, value)
	}
	tt.Key(0, KeyEnter)
	tt.Frame()
	tt.Frame()
	if got := numberText(t, tt); got != "10" || value != 10 {
		t.Fatalf("a value the app set while focused shows %q with value %v", got, value)
	}
}

func TestNumberInputKeepsSignedDeltaButtonSemantics(t *testing.T) {
	for _, test := range []struct {
		step, value float64
		labels      [2]string
		disabled    [2]bool
	}{
		{1, 0, [2]string{"Decrease", "Increase"}, [2]bool{true, false}},
		{-1, 0, [2]string{"Increase", "Decrease"}, [2]bool{false, true}},
		{0, 0, [2]string{"Decrease", "Decrease"}, [2]bool{false, false}},
	} {
		value := test.value
		tt := NewTester(func(c *Context) { NumberInput(c, &value, 0, 10, test.step) }, 400, 100)
		found := 0
		for _, node := range tt.Accessibility() {
			if node.Role == RoleButton {
				if found >= 2 || node.Label != test.labels[found] || node.Disabled != test.disabled[found] {
					t.Fatalf("step %v button %d: %+v", test.step, found, node)
				}
				found++
			}
		}
		if found != 2 {
			t.Fatalf("step %v found %d buttons", test.step, found)
		}
		tt.rt.close()
	}
}
