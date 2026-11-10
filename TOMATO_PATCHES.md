# Tomato runtime patches for MyGo

This branch carries the Tomato runtime patches on top of [upstream commit `bd590bb4`](https://github.com/egoist/mygo/commit/bd590bb4e4f99ca814e5a1a0992296e9feb1d9aa). The module path remains `github.com/egoist/mygo`. Tomato pins an immutable fork commit through a Go module replacement.

The entries below describe the inherited patches and their original upstream comparison. MyGo 0.8 adds overlapping features, including per-element disabled dimming, tooltip delay, number-input sizing, and accessibility inspection. Those do not individually replace the full Tomato contracts covered by the regression tests below.

## Patches

Entries describe the inherited patches; upstream status statements refer to v0.7.4 unless otherwise stated.

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

- **Number input submit, width and app-set values.** `ui/number.go` makes
  `NumberInput` report Enter in its input through `Submitted` and `OnSubmit`,
  forwarded to the control as `Combobox` does, and show the value in full on
  Enter as losing the focus does. A value the app sets replaces the typed text,
  as in `DecimalInput`, even while the input has the focus. The input grows
  into an explicit `Width` so the buttons stay at the end; without a `Width`
  the layout is unchanged. Upstream's `NumberInput` has none of these, so an
  app cannot apply a typed value on Enter or clamp it. `ui/number_submit_test.go`
  covers Enter, typing, stepping, button clicks, repeated frames, both widths,
  text after Enter and a value set by the app while focused.
- **Controlled number drafts.** `NumberInputDraft(c, &text, step)` binds the
  caller's complete raw string and returns checked `NumberInputParts` with
  `Root`, `Input`, `Decrease` and `Increase`. A pure transform receives current
  committed text and direction -1 or +1 for native arrows and buttons. The
  constructor derives disabled buttons from the same transform after binding.
  It never normalizes on child blur or Enter, and steps never submit;
  `Input.OnSubmit` lets the caller commit. Style actual child borders through
  the parts, and retain `Handle`, never parts, between builds. The original
  `NumberInput` signature and eager semantics are unchanged. Both share native
  construction and width/Column layout. `ui/number_draft_test.go` covers raw
  text, native batched replacements, paste, composition, button keyboard and
  pointer activation, undo/redo, disabled transitions and checked lifetimes.
  Omission before a pending native replacement binds has the same behavior as
  plain `TextInput`; this API does not flush omitted controls.
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
- **Markdown theme cache.** `ui/markdown/theme.go` and `ui/markdown/view.go`
  compare only the window-theme inputs of `DefaultTheme` and
  `editor.DefaultTheme` (`themeKey`). The tooltip callback makes `ui.Theme`
  incomparable, and callback changes must not flush Markdown caches. Required
  by the tooltip patch; not needed upstream.
- **Damage flag reuse.** `internal/scene/damage.go` grows and clears the
  per-op enclosing flags instead of appending a temporary slice, which allocates
  under `-race`. Upstream v0.7.4 still appends the temporary slice.
- **Tester field targeting.** A named `Field` whose center reaches its associated control can be clicked by name even when an interactive ancestor handles input. Same-named covering controls remain rejected. The scrollbar hit check also respects hidden axes.
- **Test stability.** `ui/bench_test.go` warms the full sidebar range before
  counting allocations. `internal/darwin/access_expansion_test.go` registers
  only its accessibility class: `registerClasses` would define `MyGoWebView` as
  a `WKWebView`, which upstream's embedded web view test replaces with a plain
  `NSView`. The v0.6.4 branch's change to `surface_drag_test.go` is no longer
  needed and upstream's test is kept.

## Upstream scrollbar axes

Tomato uses the upstream `Element.ScrollbarAxes(horizontal, vertical)` implementation from [MyGo #1650](https://github.com/egoist/mygo/pull/1650). The diff hides its native vertical indicator while its overview is visible. Horizontal indicators and scroll input remain available.

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
