//go:build darwin

package darwin

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/internal/platform"
)

// Exercise the AppKit object bridge without creating a window. The modern
// expanded property alone does not provide the legacy AXExpanded attribute.
func TestAccessibilityExpansionBridge(t *testing.T) {
	previous := theBackend
	backend := New()
	theBackend = backend
	defer func() { theBackend = previous }()
	registerClasses()
	withPool(func() {
		obj := send(send(class("MyGoAccessibilityElement"), "alloc"), "init")
		defer release(obj)
		el := &accessElement{obj: obj, s: &surface{w: &window{b: backend}, updating: true}}
		backend.byAccess[obj] = el
		defer delete(backend.byAccess, obj)
		n := platform.AccessNode{ID: 1, Role: platform.RoleButton, Label: "Disclosure", Actions: platform.ActionPress | platform.ActionExpand}
		for i, open := range []bool{true, false, true} {
			n.States = platform.AccessExpandable
			if open {
				n.States |= platform.AccessExpanded
			}
			el.apply(n, i == 0)
			if sendBool(obj, "isAccessibilityExpanded") != open || sendBool(obj, "isAccessibilityDisclosed") != open {
				t.Errorf("modern disclosure state differs from %v", open)
			}
			for _, name := range []string{"AXExpanded", "AXDisclosing"} {
				value := send(obj, "accessibilityAttributeValue:", uintptr(nsString(name)))
				if value == 0 || sendBool(value, "boolValue") != open {
					t.Errorf("%s=%s, want %v", name, goString(send(value, "description")), open)
				}
			}
			var names []string
			for _, name := range arrayItems(send(obj, "accessibilityAttributeNames")) {
				names = append(names, goString(name))
			}
			if !slices.Contains(names, "AXExpanded") || !sendBool(obj, "accessibilityIsAttributeSettable:", uintptr(nsString("AXDisclosing"))) {
				t.Error("expansion attribute or action is unavailable")
			}
		}
		n.Actions = 0
		n.States |= platform.AccessDisabled
		el.apply(n, false)
		if sendBool(obj, "accessibilityIsAttributeSettable:", uintptr(nsString("AXDisclosing"))) {
			t.Error("disabled disclosure is settable")
		}
	})
}
