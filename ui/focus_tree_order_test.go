package ui

import (
	"slices"
	"testing"
)

func TestFocusTreeOrderPreservesPointerStacking(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		clicked := 0
		tt := coreNewTester(func(c *context) {
			theme := *c.Theme()
			theme.FocusFollowsTreeOrder = enabled
			c.SetTheme(&theme)
			coreRow(c).Size(120, 80).Children(func() {
				coreButton(c, "Before").Width(60).Height(80)
				if coreBox(c).Absolute().Left(54).Top(0).Width(12).Height(80).Focusable().Label("Splitter").Clicked() {
					clicked++
				}
				coreButton(c, "After").Width(60).Height(80)
			})
		}, 120, 80)
		labels := []string{"Before", "Splitter", "After"}
		want := []string{"Before", "After", "Splitter"}
		if enabled {
			want = labels
		}
		if got := tabs(tt, 0, 3, labels...); !slices.Equal(got, want) {
			t.Fatalf("enabled=%v: Tab=%v want=%v", enabled, got, want)
		}
		tt.Key(Shift, KeyTab)
		if !tt.Focused(want[1]) {
			t.Fatal("reverse traversal lost position")
		}
		for _, x := range []float32{55, 65} {
			tt.ClickAt(x, 40)
		}
		if clicked != 2 {
			t.Fatal("keyboard order changed the overlaid pointer target")
		}
	}
}

func TestFocusTreeOrderIncludesDeferredRows(t *testing.T) {
	var list ListState
	visible := true
	tt := coreNewTester(func(c *context) {
		theme := *c.Theme()
		theme.FocusFollowsTreeOrder = true
		c.SetTheme(&theme)
		coreColumn(c).Fill().Children(func() {
			coreButton(c, "Before")
			coreList(c, &list, 1, func(int) {
				if visible {
					coreButton(c, "Row")
				}
			}).Height(50)
			coreBox(c).Disabled(true).Children(func() { coreButton(c, "Disabled") })
			coreBox(c).Inert(true).Children(func() { coreButton(c, "Inert") })
			coreButton(c, "After")
		})
	}, 300, 300)
	labels := []string{"Before", "Row", "After", "Disabled", "Inert"}
	if got := tabs(tt, 0, 3, labels...); !slices.Equal(got, labels[:3]) {
		t.Fatalf("deferred row order=%v", got)
	}
	visible = false
	tt.Frame()
	tt.Click("Before")
	tt.Key(0, KeyTab)
	if !tt.Focused("After") {
		t.Fatal("removed row retained a stop")
	}
	visible = true
	tt.Frame()
	tt.Click("Before")
	tt.Key(0, KeyTab)
	if !tt.Focused("Row") {
		t.Fatal("returning row lost its tree position")
	}
}

func TestFocusTreeOrderRetainsScopesAndGroups(t *testing.T) {
	d := &dialogView{size: "Small"}
	tt := coreNewTester(func(c *context) {
		theme := *c.Theme()
		theme.FocusFollowsTreeOrder = true
		c.SetTheme(&theme)
		d.view(c)
	}, 500, 400)
	tt.Click("Open")
	labels := []string{"Name", "Small", "OK", "Cancel", "Before", "Open", "After"}
	if !tt.Focused("Name") {
		t.Fatal("dialog did not take focus")
	}
	if got := tabs(tt, 0, 4, labels...); !slices.Equal(got, []string{"Small", "OK", "Cancel", "Name"}) {
		t.Fatalf("dialog focus escaped: %v", got)
	}
	tt.Key(0, KeyEscape)
	if !tt.Focused("Open") {
		t.Fatal("closing dialog lost its opener")
	}

	open := false
	pt := coreNewTester(func(c *context) {
		theme := *c.Theme()
		theme.FocusFollowsTreeOrder = true
		c.SetTheme(&theme)
		coreColumn(c).Fill().Children(func() {
			anchor := coreButton(c, "Anchor")
			coreToolbar(c, func() { coreButton(c, "First"); coreButton(c, "Second") })
			coreButton(c, "After")
			corePopover(c, anchor, &open, func() { coreButton(c, "Popup") })
		})
	}, 500, 400)
	pt.Click("Anchor")
	open = true
	pt.Frame()
	labels = []string{"Anchor", "Popup", "First", "Second", "After"}
	if got := tabs(pt, 0, 3, labels...); !slices.Equal(got, []string{"Popup", "First", "After"}) {
		t.Fatalf("popover/group order=%v", got)
	}
	pt.Key(Shift, KeyTab)
	pt.Key(0, KeyRight)
	if !pt.Focused("Second") {
		t.Fatal("group arrows stopped working")
	}
}
