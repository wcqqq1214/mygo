# Tomato runtime patches for MyGo

This branch preserves the complete upstream source, tests, documentation, examples,
website and MIT license of [`github.com/egoist/mygo` v0.7.4](https://github.com/egoist/mygo/tree/v0.7.4),
commit `9bd34e1eb9525d9a1fe698aadc964c49c8ff8e7e`. It carries the Tomato runtime
fixes that v0.7.4 does not provide. The module path remains
`github.com/egoist/mygo`; Tomato pins an immutable commit of this fork through a
Go module replacement. The module cache is never modified.
The upstream module checksum is
`h1:0qkDpDRo7C9g8PM8eu4V6FVouhLiTrUdqKD6I31Lv3Q=`.

The previous maintenance branch, `tomato/v0.6.4-patches`, is kept unchanged.
Its commits were cherry-picked onto v0.7.4; upstream commits and authors are
preserved without rewriting history.

## Patches

Each entry lists its upstream status at v0.7.4.

- **Pending find feedback.** Native `FindBar` accepts optional
  `FindBarOptions{Pending: true}`. Pending work displays a localized searching
  status, blocks match navigation and keeps typing, Escape and Done available.
  Existing calls without options retain their behavior.
- **Retained layout accounting.** `TextLayout.MemoryCost()` estimates owned Go
  layout data for cache admission, including a conservative reservation for
  lazy caret geometry. Shared fonts, glyph atlases and platform caches are
  excluded; it is not process RSS. Strings are charged by logical length, so
  consumers with strict retention budgets must own substring storage. Estimation
  is safe alongside geometry calls.

- **Inert subtrees.** `ui/element.go` adds `Element.Inert(bool)` using the existing
  inert subtree mechanism; `ui/api_gen.go` is regenerated with `go generate ./ui`.
  Read-only SVG web views remain visible but cannot receive pointer, keyboard,
  or accessibility interaction. Their native wrapper supplies the image label.
  Upstream's `IgnoreInput` (#927) keeps the subtree in the accessibility tree
  and does not cover embedded web views, so it is not equivalent.
- **Wide colors in canvases.** `internal/scene/scene.go` detects wide colors in
  nested and retained canvases, and `ui/window.go` includes them when choosing
  the output format. `internal/gpu/metal/metal.go` renders wide canvases into
  float16 textures only when the output supports wide color; sRGB output uses
  the same gamut-mapped colors as direct drawing. Texture format participates
  in the cache key. Upstream still reads only the top-level scene's wide colors.
- **Color transition reversals.** `ui/transition.go` adds opt-in
  `ShortenColorReversals` for fixed-duration color transitions without a delay.
  Background and border keep independent clocks; interrupted reversals follow
  the [CSS shortening rules](https://www.w3.org/TR/css-transitions-1/#reversing).
  Default, spring and delayed transitions retain upstream behavior. Not upstream.
- **Disabled colors.** `ui/theme.go`, `ui/paint.go` and `ui/inline.go` add the
  opt-in `DisableAutoDimming` theme setting, so Tomato supplies its own disabled
  colors instead of opacity. Disabled input, focus and accessibility behavior
  remain intact. Not upstream.
- **Tooltip policy.** `ui/theme.go`, `ui/tooltip.go`, `ui/input.go` and
  `ui/layout.go` add an opt-in `TooltipPolicy` and `TooltipStyle` callback: a fixed
  delay, hoverable content, below-first placement with a viewport inset, and
  dismissal on input, window changes, programmatic scrolling and Escape.
  Default timing and appearance stay intact. Upstream has only `TooltipBase`.
- **Focus in tree order.** `ui/theme.go`, `ui/element.go`, `ui/runtime.go` and
  `ui/scope.go` add the opt-in `FocusFollowsTreeOrder` setting. Tab follows the
  element tree, including rows constructed during layout, while paint and hit
  order, dialog containment, popover anchors and roving groups keep their
  semantics. Not upstream.
- **Disclosure accessibility.** `ui/access.go` marks an explicitly
  expanded/collapsed control as expandable in both states.
  `internal/darwin/access.go` exposes `AXExpanded` and `AXDisclosing` for custom
  controls and makes disclosure settable only while the action is enabled.
  `ui/headless.go` exposes immutable snapshots through `Tester.Accessibility`.
  Upstream exposes disclosure only for outline rows.
- **Unscaled image compositing.** `internal/raster/image_unscaled.go` and
  `internal/raster/raster.go` avoid bilinear interpolation for pixel-aligned
  images drawn at their original size without opacity, tint or rounded corners.
  They preserve premultiplied blending and fractional clip coverage; other
  images use the general sampler. Adapted for v0.7.4: an image whose pixels were
  released (`Bitmap.ReleasePixels`) and not restored falls back to the sampler,
  which draws nothing. `internal/raster/canvas.go` keeps byte slices local during
  channel conversion. Not upstream.
- **Per-axis scroll bars.** `ui/element.go`, `ui/context.go`, `ui/runtime.go`,
  `ui/input.go` and `ui/paint.go` add `Element.ScrollbarAxes(horizontal, vertical)`.
  A hidden axis keeps scrolling but has no painted thumb, pointer hit target or
  cursor override, and hiding the axis being dragged ends the drag. Adapted for
  v0.7.4: upstream's `Scrollbars(visible bool)` (#972) hides both bars at once,
  so the v0.6.4 branch's `Scrollbars(horizontal, vertical)` is renamed to
  `ScrollbarAxes` and `Scrollbars(false)` overrides it. Tomato hides only the code
  list's vertical bar while its overview is visible.
- **Markdown theme cache.** `ui/markdown/theme.go` and `ui/markdown/view.go`
  compare only the window-theme inputs of `DefaultTheme` and
  `editor.DefaultTheme` (`themeKey`). The tooltip callback makes `ui.Theme`
  incomparable, and callback changes must not flush Markdown caches. Required
  by the tooltip patch; not needed upstream.
- **Damage flag reuse.** `internal/scene/damage.go` grows and clears the
  per-op enclosing flags instead of appending a temporary slice, which allocates
  under `-race`. Upstream v0.7.4 still appends the temporary slice.
- **Test stability.** `ui/bench_test.go` warms the full sidebar range before
  counting allocations. `internal/darwin/access_expansion_test.go` registers
  only its accessibility class: `registerClasses` would define `MyGoWebView` as
  a `WKWebView`, which upstream's embedded web view test replaces with a plain
  `NSView`. The v0.6.4 branch's change to `surface_drag_test.go` is no longer
  needed and upstream's test is kept.

## Regression tests

`ui/inert_test.go`, `ui/color_reversal_test.go`, `ui/disabled_dimming_test.go`,
`ui/tooltip_policy_test.go`, `ui/focus_tree_order_test.go`,
`ui/access_snapshot_test.go`, `ui/scrollbar_visibility_test.go`,
`ui/markdown/theme_cache_test.go`, `internal/raster/image_unscaled_test.go`,
`internal/scene/damage_test.go`, `internal/darwin/access_expansion_test.go` and
`internal/gpu/metal/canvas_wide_test.go` cover these patches. The Metal test
compares actual GPU pixel readback and requires a Metal device.

Tomato's `bun run verify` and `bun run verify:native` run these packages through
their canonical Go import paths and fail when a pinned test does not run. When
upgrading MyGo, compare each patch with upstream and drop it once equivalent
behavior is available.
