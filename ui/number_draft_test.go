package ui

import (
	"strconv"
	"testing"

	"github.com/egoist/mygo/internal/platform"
)

var _ func(*Context, *float64, float64, float64, float64) Element = NumberInput

func draftStep(raw string, direction int) string {
	value, err := strconv.Atoi(raw)
	if err != nil {
		value = 14
	}
	return strconv.Itoa(max(12, min(32, value+direction)))
}

func draftTester(text *string, submits *int) *Tester {
	return NewTester(func(c *Context) {
		Column(c).Padding(10).AlignItems(Start).Children(func() {
			p := NumberInputDraft(c, text, draftStep)
			p.Root.Width(248).Label("Number")
			p.Input.Label("Count").OnSubmit(func() { *submits++ })
		})
	}, 400, 100)
}

func TestNumberInputDraftPreservesRawTextAndChildFocus(t *testing.T) {
	for _, raw := range []string{"", "abc", " 40 ", "12.5", "1e100"} {
		t.Run(raw, func(t *testing.T) {
			text, submits := "14", 0
			tt := draftTester(&text, &submits)
			defer tt.rt.close()
			if err := tt.Click("Count"); err != nil {
				t.Fatal(err)
			}
			tt.Key(Cmd, KeyA)
			tt.Key(0, KeyBackspace)
			tt.Type(raw)
			tt.Key(0, KeyEnter)
			tt.Frame()
			if text != raw || numberText(t, tt) != raw || submits != 1 {
				t.Fatalf("Enter text=%q editor=%q submits=%d", text, numberText(t, tt), submits)
			}
			tt.Key(0, KeyTab)
			tt.Frame()
			if text != raw || !tt.Focused("Decrease") || submits != 1 {
				t.Fatalf("child focus text=%q submits=%d", text, submits)
			}
			tt.Frame()
			if submits != 1 {
				t.Fatal("rebuild repeated submit")
			}
		})
	}
}

func TestNumberInputDraftUsesCurrentNativeTextBetweenFrames(t *testing.T) {
	for _, operation := range []string{"replace", "paste", "selection"} {
		t.Run(operation, func(t *testing.T) {
			text, submits := "14", 0
			tt := draftTester(&text, &submits)
			defer tt.rt.close()
			if err := tt.Click("Count"); err != nil {
				t.Fatal(err)
			}
			switch operation {
			case "replace":
				tt.h.ime.Client.ReplaceText(&TextInputRange{Start: 0, End: 2}, "21")
			case "paste":
				tt.SetClipboard("21")
				tt.rt.event(platform.SurfaceEvent{Kind: platform.KeyPressed, Mods: platform.Modifiers(Cmd), Key: platform.KeyA})
				tt.rt.event(platform.SurfaceEvent{Kind: platform.KeyPressed, Mods: platform.Modifiers(Cmd), Key: platform.KeyV})
			case "selection":
				tt.rt.event(platform.SurfaceEvent{Kind: platform.KeyPressed, Mods: platform.Modifiers(Cmd), Key: platform.KeyA})
				tt.h.ime.Client.ReplaceText(nil, "21")
			}
			tt.rt.event(platform.SurfaceEvent{Kind: platform.KeyPressed, Key: platform.KeyUp})
			tt.rt.event(platform.SurfaceEvent{Kind: platform.KeyPressed, Key: platform.KeyDown})
			tt.Frame()
			if text != "21" || numberText(t, tt) != "21" || submits != 0 {
				t.Fatalf("batch text=%q editor=%q submits=%d", text, numberText(t, tt), submits)
			}
		})
	}
}

func TestNumberInputDraftButtonsUseBoundTextAndNeverSubmit(t *testing.T) {
	for _, activation := range []string{"pointer", "composition", "enter", "space"} {
		t.Run(activation, func(t *testing.T) {
			text, submits := "14", 0
			tt := draftTester(&text, &submits)
			defer tt.rt.close()
			if err := tt.Click("Count"); err != nil {
				t.Fatal(err)
			}
			if activation == "pointer" || activation == "composition" {
				client := tt.h.ime.Client
				if activation == "composition" {
					client.SetMarkedText(&TextInputRange{Start: 0, End: 2}, "候補", TextInputRange{Start: 2, End: 2})
					var composition platform.InputComposition
					composition.End(client)
					composition.Replace(client, nil, "21")
				} else {
					client.ReplaceText(&TextInputRange{Start: 0, End: 2}, "21")
				}
				bounds, _ := tt.Find("Increase")
				tt.rt.event(platform.SurfaceEvent{Kind: platform.PointerDown, X: float64(bounds.X + bounds.W/2), Y: float64(bounds.Y + bounds.H/2)})
				tt.rt.event(platform.SurfaceEvent{Kind: platform.PointerUp, X: float64(bounds.X + bounds.W/2), Y: float64(bounds.Y + bounds.H/2)})
				tt.Frame()
			} else {
				tt.Key(Cmd, KeyA)
				tt.Type("21")
				tt.Key(0, KeyTab)
				tt.Key(0, KeyTab)
				if !tt.Focused("Increase") {
					t.Fatal("increase did not take focus")
				}
				key := KeyEnter
				if activation == "space" {
					key = KeySpace
				}
				tt.Key(0, key)
			}
			if text != "22" || numberText(t, tt) != "22" || submits != 0 {
				t.Fatalf("button text=%q editor=%q submits=%d", text, numberText(t, tt), submits)
			}
			tt.Frame()
			if text != "22" || submits != 0 {
				t.Fatal("button applied twice")
			}
			if err := tt.Click("Count"); err != nil {
				t.Fatal(err)
			}
			tt.Key(Cmd, KeyZ)
			if text != "14" || numberText(t, tt) != "14" || submits != 0 {
				t.Fatalf("button undo text=%q editor=%q submits=%d", text, numberText(t, tt), submits)
			}
			tt.Key(Cmd|Shift, KeyZ)
			if text != "22" || numberText(t, tt) != "22" || submits != 0 {
				t.Fatalf("button redo text=%q editor=%q submits=%d", text, numberText(t, tt), submits)
			}
		})
	}
}

func TestNumberInputDraftBoundsAndPreview(t *testing.T) {
	for _, boundary := range []struct {
		raw, want, outward, inward string
		first, second              Key
	}{
		{"32", "31", "Increase", "Decrease", KeyUp, KeyDown},
		{"12", "13", "Decrease", "Increase", KeyDown, KeyUp},
	} {
		for _, method := range []string{"arrows", "buttons"} {
			t.Run(boundary.raw+method, func(t *testing.T) {
				text, submits := boundary.raw, 0
				tt := draftTester(&text, &submits)
				defer tt.rt.close()
				if method == "arrows" {
					if err := tt.Click("Count"); err != nil {
						t.Fatal(err)
					}
					tt.rt.event(platform.SurfaceEvent{Kind: platform.KeyPressed, Key: platform.Key(boundary.first)})
					tt.rt.event(platform.SurfaceEvent{Kind: platform.KeyPressed, Key: platform.Key(boundary.second)})
					tt.Frame()
				} else {
					for _, node := range tt.Accessibility() {
						if node.Label == boundary.outward && !node.Disabled {
							t.Fatal("outward button enabled")
						}
					}
					if err := tt.Click(boundary.outward); err != nil {
						t.Fatal(err)
					}
					if err := tt.Click(boundary.inward); err != nil {
						t.Fatal(err)
					}
				}
				if text != boundary.want || submits != 0 {
					t.Fatalf("got %q submits=%d want %q", text, submits, boundary.want)
				}
			})
		}
	}
}

func TestNumberInputDraftNativeCompositionUndoRedo(t *testing.T) {
	text, submits := "14", 0
	tt := draftTester(&text, &submits)
	defer tt.rt.close()
	if err := tt.Click("Count"); err != nil {
		t.Fatal(err)
	}
	client := tt.h.ime.Client
	client.SetMarkedText(&TextInputRange{Start: 0, End: 2}, "候補", TextInputRange{Start: 2, End: 2})
	tt.Frame()
	if submits != 0 || numberText(t, tt) == "14候補" {
		t.Fatal("composition submitted or duplicated")
	}
	var composition platform.InputComposition
	composition.End(client)
	composition.Replace(client, nil, "21")
	tt.Frame()
	if text != "21" || submits != 0 {
		t.Fatalf("composition text=%q submits=%d", text, submits)
	}
	tt.Key(Cmd, KeyZ)
	if text != "14" {
		t.Fatalf("undo %q", text)
	}
	tt.Key(Cmd|Shift, KeyZ)
	if text != "21" {
		t.Fatalf("redo %q", text)
	}
}

func TestNumberInputDraftPartsExpireWithBuild(t *testing.T) {
	text := "14"
	var previous NumberInputParts
	tt := NewTester(func(c *Context) {
		if previous.Root.owner != nil {
			for _, part := range []Element{previous.Root, previous.Input, previous.Decrease, previous.Increase} {
				if part.Valid() {
					t.Fatal("old part became valid")
				}
			}
		}
		previous = NumberInputDraft(c, &text, draftStep)
	}, 400, 100)
	defer tt.rt.close()
	tt.Frame()
	tt.Frame()
}

func TestNumberInputDraftRecomputesDisabledButtons(t *testing.T) {
	text, submits := "32", 0
	tt := draftTester(&text, &submits)
	defer tt.rt.close()
	assertIncrease := func(disabled bool) {
		t.Helper()
		for _, n := range tt.Accessibility() {
			if n.Label == "Increase" {
				if n.Disabled != disabled {
					t.Fatalf("increase disabled=%t want %t, text=%q", n.Disabled, disabled, text)
				}
				return
			}
		}
		t.Fatal("missing Increase")
	}
	assertIncrease(true)
	text = "31"
	tt.Frame()
	assertIncrease(false)
	if err := tt.Click("Increase"); err != nil {
		t.Fatal(err)
	}
	if text != "32" || numberText(t, tt) != "32" {
		t.Fatalf("increase text=%q editor=%q", text, numberText(t, tt))
	}
	assertIncrease(true)
	if err := tt.Click("Decrease"); err != nil {
		t.Fatal(err)
	}
	assertIncrease(false)
	if text != "31" || numberText(t, tt) != "31" || submits != 0 {
		t.Fatalf("decrease text=%q editor=%q submits=%d", text, numberText(t, tt), submits)
	}
}

func TestNumberInputDraftKeepsSharedWidthAndColumnHeight(t *testing.T) {
	text, value := "14", 14.0
	column := false
	tt := NewTester(func(c *Context) {
		Column(c).Padding(10).AlignItems(Start).Children(func() {
			NumberInput(c, &value, 12, 32, 1).Label("Eager").Width(248)
			p := NumberInputDraft(c, &text, draftStep)
			p.Root.Label("Draft").Width(248)
			p.Input.Label("Draft text")
			if column {
				p.Root.Column()
			}
		})
	}, 400, 300)
	defer tt.rt.close()
	eager, _ := tt.Find("Eager")
	draft, _ := tt.Find("Draft")
	if eager.W != draft.W || eager.H != draft.H {
		t.Fatalf("shared layout eager=%v draft=%v", eager, draft)
	}
	input, _ := tt.Find("Draft text")
	column = true
	tt.Frame()
	after, _ := tt.Find("Draft text")
	if after.H != input.H {
		t.Fatalf("column input height %v want %v", after, input)
	}
}

func TestNumberInputDraftPendingReplacementMatchesTextInput(t *testing.T) {
	for _, omit := range []bool{false, true} {
		results := []string{}
		for _, draft := range []bool{false, true} {
			text, open := "14", true
			tt := NewTester(func(c *Context) {
				if !open {
					return
				}
				if draft {
					NumberInputDraft(c, &text, draftStep).Input.Label("Count")
				} else {
					TextInput(c, &text).Label("Count")
				}
			}, 400, 100)
			if err := tt.Click("Count"); err != nil {
				t.Fatal(err)
			}
			tt.h.ime.Client.ReplaceText(&TextInputRange{Start: 0, End: 2}, "21")
			open = !omit
			tt.Frame()
			results = append(results, text)
			tt.rt.close()
		}
		if results[0] != results[1] || !omit && results[1] != "21" {
			t.Fatalf("omitted=%t plain=%q draft=%q", omit, results[0], results[1])
		}
	}
}
