# Tomato runtime patches for MyGo

This branch preserves the complete upstream source, tests, documentation, examples,
website and MIT license of [`github.com/egoist/mygo` v0.6.4](https://github.com/egoist/mygo/tree/v0.6.4),
commit `6255ea40139d06e0762a7c4a32626148ba410478`. It carries the existing Tomato
runtime fixes without upgrading the framework API. The module path remains
`github.com/egoist/mygo`; Tomato pins an immutable commit of this fork through a
Go module replacement. The module cache is never modified.
The upstream module checksum is
`h1:VzCqZKgvXcym5cSYKoC6IkTgcVeGZquPiui/a3oE36o=`.

The runtime and generated-code changes are:

- `ui/element.go` adds `Element.Inert(bool)` using the existing inert subtree
  mechanism. `ui/api_gen.go` is regenerated with `go generate ./ui`.
  Read-only SVG web views remain visible but cannot receive pointer, keyboard,
  or accessibility interaction. Their native wrapper supplies the image label.
- `internal/scene/scene.go` detects wide colors in nested and retained canvases.
  `ui/window.go` includes those colors when choosing the output format.
- `internal/gpu/metal/metal.go` renders wide canvases into float16 textures only
  when the output supports wide color. sRGB output uses the same gamut-mapped
  colors as direct drawing. Texture format participates in the cache key so
  moving between screens, or changing a canvas's colors, cannot reuse a texture
  with the wrong format. Other renderers keep their existing sRGB behavior.
- `ui/transition.go` adds opt-in `ShortenColorReversals` for fixed-duration color
  transitions without a delay. Background and border keep independent clocks;
  interrupted reversals follow the [CSS shortening rules](https://www.w3.org/TR/css-transitions-1/#reversing),
  including repeated reversals. Default, spring and delayed transitions retain
  upstream behavior. Tomato enables the option for its shared control styles.
- `ui/theme.go`, `ui/paint.go` and `ui/inline.go` add the opt-in
  `DisableAutoDimming` theme setting. Tomato supplies official Radix disabled
  colors instead of multiplying block or inline opacity. Disabled input,
  keyboard focus and accessibility behavior remain intact; other themes keep
  the default dimming. `ui/disabled_dimming_test.go` covers both appearances
  and the retained disabled semantics in the native runtime checks.
- `ui/theme.go`, `ui/tooltip.go`, `ui/input.go` and `ui/layout.go` add an opt-in
  tooltip policy and appearance callback. Tomato uses a fixed 300 ms delay for
  pointer and keyboard hints, hoverable content, a below-first position with
  an 8 px viewport inset, and its shared Radix surface/border/shadow tokens.
  Input, window changes and programmatic scrolling dismiss an engaged hint;
  Escape consumes a pending or visible hint before reaching its parent.
  Default timing, warm acceleration, appearance and positioning stay intact.
  `ui/tooltip_policy_test.go` covers engagement, live labels, disabled/removed
  anchors, scrolling, dismissal and narrow/edge placement. It is included in
  the native runtime checks alongside the existing default-tooltip tests.

- `ui/theme.go`, `ui/element.go`, `ui/runtime.go` and `ui/scope.go` add the opt-in
  `FocusFollowsTreeOrder` setting. Tomato places overlay splitters between their
  neighboring panes in the element tree while retaining their full pointer hit
  area above both panes. Tab follows that tree, including rows constructed during
  layout. Paint/hit order, dialog containment, popover anchors, disabled/inert
  exclusion and roving groups keep their existing semantics. The default remains
  paint order. `ui/focus_tree_order_test.go` exercises both settings and these
  boundaries in the native runtime checks.

- `ui/access.go` marks an explicitly expanded/collapsed control as expandable
  in both states. This lets macOS distinguish a collapsed disclosure from an
  ordinary button. Custom click controls consume assistive expand/collapse
  requests only when the requested state changes; passive labels advertise no
  action. `internal/darwin/access.go` exposes the legacy `AXExpanded` and
  `AXDisclosing` attributes, synchronizes disclosure state for custom controls
  and makes disclosure settable only while the action is enabled. The AppKit
  object bridge is checked without opening windows in
  `internal/darwin/access_expansion_test.go`.
  `ui/headless.go` exposes immutable accessibility snapshots
  through `Tester.Accessibility`, so application tests verify actual semantic
  roles and states rather than infer them from pixels. `ui/access_snapshot_test.go`
  checks transitions through both states in the native runtime checks.

`ui/inert_test.go` covers visibility, pointer exclusion, focus, reactivation,
effects and retained-canvas output selection. `internal/gpu/metal/canvas_wide_test.go`
compares actual GPU pixel readback with direct drawing through nested, retained,
wide → sRGB → wide and content-change cases. `bun run verify` runs these focused
regressions. `bun run verify:native` requires Metal for the pixel test; it cannot
silently count an unavailable GPU as passing.

`ui/color_reversal_test.go` compares interrupted and repeated reversal timings,
ordinary new targets, independent background/border deadlines and the unchanged
default. The application also checks hover pixels, reduced motion and folder
rotation. These runtime regressions are included in `bun run verify`.

The full upstream tree and all framework regression tests are retained. The
`bun run verify` and `bun run verify:native` commands above run in Tomato and test
these packages through their canonical Go import paths. When upgrading MyGo,
compare these changes with upstream and remove the replacement once equivalent
behavior is available. This branch is maintained separately from upstream main;
upstream commits and authors are preserved without rewriting history.
