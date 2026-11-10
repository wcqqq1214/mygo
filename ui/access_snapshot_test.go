package ui

import (
	"testing"

	"github.com/egoist/mygo/internal/platform"
)

func TestAccessibilitySnapshotRetainsCollapsedState(t *testing.T) {
	open := false
	tt := coreNewTester(func(c *context) {
		if coreButton(c, "Disclosure").Expanded(open).Clicked() {
			open = !open
		}
	}, 200, 100)
	for _, want := range []bool{false, true, false} {
		found := false
		for _, node := range tt.Accessibility() {
			if node.Label == "Disclosure" && node.Role == RoleButton {
				found = true
				if !node.Expandable || node.Expanded != want {
					t.Fatalf("state=%+v want expanded=%v", node, want)
				}
			}
		}
		if !found {
			t.Fatal("button missing from semantic snapshot")
		}
		tt.Click("Disclosure")
	}
}

func TestAccessibilitySnapshotCustomDisclosureActions(t *testing.T) {
	open, disabled, clicks := false, false, 0
	tt := NewTester(func(c *Context) {
		Button(c, "Disclosure").Expanded(open).Disabled(disabled).OnClick(func() {
			open = !open
			clicks++
		})
		if open {
			Text(c, "Child")
		}
	}, 200, 100)
	tt.Accessibility()
	for i, want := range []bool{true, true, false, false, true} {
		n := accessNode(t, tt.h.access, platform.RoleButton, "Disclosure")
		if n.Actions&platform.ActionExpand == 0 {
			t.Fatal("custom disclosure has no accessibility action")
		}
		action := platform.AccessCollapse
		if want {
			action = platform.AccessExpand
		}
		before, changed := clicks, open != want
		wantClicks := before
		if changed {
			wantClicks++
		}
		tt.send(platform.SurfaceEvent{Kind: platform.AccessAction, ID: n.ID, Action: action})
		if open != want || tt.HasText("Child") != want || clicks != wantClicks {
			t.Fatalf("step %d: open=%v child=%v clicks=%d", i, open, tt.HasText("Child"), clicks-before)
		}
		tt.Frame()
		if clicks != wantClicks {
			t.Fatal("action repeated on another frame")
		}
	}
	disabled = true
	tt.Frame()
	n := accessNode(t, tt.h.access, platform.RoleButton, "Disclosure")
	if n.Actions&platform.ActionExpand != 0 {
		t.Fatal("disabled disclosure advertises expansion")
	}
	tt.send(platform.SurfaceEvent{Kind: platform.AccessAction, ID: n.ID, Action: platform.AccessCollapse})
	if !open {
		t.Fatal("disabled disclosure changed")
	}
}

func TestAccessibilitySnapshotPassiveExpansionHasNoAction(t *testing.T) {
	tt := coreNewTester(func(c *context) {
		coreBox(c).Size(100, 30).Role(RoleGroup).Label("Section").Expanded(true)
	}, 200, 100)
	tt.Accessibility()
	n := accessNode(t, tt.h.access, platform.RoleGroup, "Section")
	if n.States&platform.AccessExpanded == 0 || n.Actions&platform.ActionExpand != 0 {
		t.Fatalf("passive expansion state: %+v", n)
	}
}
