package ui

import (
	"math"
	"strconv"
	"strings"
)

// NumberInput creates a text input editing *value as a number between lo
// and hi, which the arrows Up and Down, and the buttons beside it, step
// by step. What is typed applies as soon as it is a number in range, and
// shows rounded to the decimals of step once the input loses the focus or
// Enter is pressed. Changed reports a new value, and Submitted Enter in the
// input. A Width
// wider than the natural one is taken by the input. A value the app sets
// replaces what was typed, as in a DecimalInput.
//
//	ui.NumberInput(c, &app.copies, 1, 99, 1)
func coreNumberInput(c *context, value *float64, lo, hi, step float64) *node {
	row := numberInputRow(c)
	decimals := 0
	if s := strconv.FormatFloat(step, 'f', -1, 64); strings.Contains(s, ".") {
		decimals = len(s) - strings.IndexByte(s, '.') - 1
	}
	format := func(v float64) string { return strconv.FormatFloat(v, 'f', decimals, 64) }
	text := coreLocal(row, "text", func() string { return format(*value) })
	shown := coreLocal(row, "shown", func() float64 { return *value })
	set := func(v float64) {
		v = math.Max(lo, math.Min(hi, v))
		if decimals >= 0 {
			p := math.Pow(10, float64(decimals))
			v = math.Round(v*p) / p
		}
		if v != *value {
			*value, *shown = v, v
			row.st.markChanged()
			c.rt.consumed = true
		}
	}
	numberInputChildren(c, row, text, func(in *node) {
		in.afterInput(func() {
			if *value != *shown {
				*shown, *text = *value, format(*value)
				c.rt.consumed = true
			}
			if in.Changed() {
				if v, err := strconv.ParseFloat(strings.TrimSpace(*text), 64); err == nil && v >= lo && v <= hi {
					set(v)
				}
			}
			if in.Submitted() {
				row.st.markSubmitted()
				*text = format(*value)
			}
			if in.Shortcut(0, KeyUp) {
				set(*value + step)
				*text = format(*value)
			}
			if in.Shortcut(0, KeyDown) {
				set(*value - step)
				*text = format(*value)
			}
			if !in.Focused() {
				// What the app set, or what was typed, shown in full.
				*text = format(*value)
			}
		})
		in.hasRange, in.accRange, in.accStep = true, [3]float64{lo, hi, *value}, step
	}, func(btn *node, in *node, direction int) {
		delta := float64(direction) * step
		btn.Label(map[bool]string{true: "Increase", false: "Decrease"}[delta > 0])
		btn.Disabled(delta < 0 && *value <= lo || delta > 0 && *value >= hi)
		btn.afterInput(func() {
			if btn.Clicked() {
				set(*value + delta)
				*text = format(*value)
			}
		})
	})
	return row
}

// NumberInputDraft creates a native number editor bound to the complete raw
// *text, including empty and invalid drafts. It never normalizes on Enter or
// editor blur. Input.OnSubmit reports Enter; stepping never submits.
// step must be non-nil and pure. It receives the current committed text and
// direction -1 or +1 for Up/Down and the native buttons. A button is disabled
// when previewing step returns the same text. The callback must not retain
// Elements or TextKey. Parts are checked handles for the current build only.
func coreNumberInputDraft(c *context, text *string, step func(string, int) string) numberInputParts {
	row := numberInputRow(c)
	p := numberInputChildren(c, row, text, func(in *node) {
		in.OnTextKey(func(k TextKey) bool {
			if k.Mods != 0 {
				return false
			}
			direction := 0
			switch k.Key {
			case KeyUp:
				direction = 1
			case KeyDown:
				direction = -1
			default:
				return false
			}
			k.SetText(step(k.Text(), direction))
			return true
		})
	}, func(*node, *node, int) {})
	row.afterInput(func() {
		for _, b := range []struct {
			node      *node
			direction int
		}{{p.Decrease, -1}, {p.Increase, 1}} {
			if b.node.Clicked() {
				next := step(*text, b.direction)
				if next != *text {
					*text = next
					p.Input.st.editor.setText(next)
					p.Input.st.editor.compose = ""
					p.Input.st.markChanged()
					c.rt.consumed = true
				}
			}
		}
		p.Decrease.Disabled(step(*text, -1) == *text)
		p.Increase.Disabled(step(*text, 1) == *text)
	})
	return p
}

type numberInputParts struct {
	Root, Input, Decrease, Increase *node
	c                               *context
}

func numberInputRow(c *context) *node {
	row := coreRow(c).Gap(c.theme.Space(1)).Shrink(0).AlignItems(Center)
	row.widget = "NumberInput"
	return row
}

func numberInputChildren(c *context, row *node, text *string, input func(*node), button func(*node, *node, int)) numberInputParts {
	t := c.theme
	p := numberInputParts{Root: row, c: c}
	row.Children(func() {
		in := coreTextInput(c, text).Width(t.Space(20))
		// Keep the automatic basis so Column uses the input's height.
		in.grow = 1
		in.growWithParentWidth = true
		in.widget = "NumberInput"
		p.Input = in
		input(in)
		for _, b := range []struct {
			label     string
			direction int
		}{{"−", -1}, {"+", 1}} {
			btn := coreButton(c, b.label).Padding(t.Space(1), t.Space(2)).Label(map[bool]string{true: "Increase", false: "Decrease"}[b.direction > 0])
			if b.direction < 0 {
				p.Decrease = btn
			} else {
				p.Increase = btn
			}
			button(btn, in, b.direction)
			btn.TextColor(t.Text)
		}
	})
	return p
}
