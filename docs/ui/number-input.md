# Number input

<!-- preview number-input 160 -->

`ui.NumberInput` creates a text input editing a `*float64` as a number
between two bounds, which Up and Down, and the buttons beside it, change by
a step. What is typed applies as soon as it is a number in range, and shows
rounded to the decimals of the step once the input loses the focus.
`Changed` reports a new value, and `Submitted` (or `OnSubmit`) Enter in the
input, as for a [text input](text-input.md), so that an app can apply what
was typed only then. A `Width` wider than the natural one is taken by the
input, which keeps the buttons at the end. Enter shows the value in full, as
losing the focus does, and a value the app sets replaces what was typed.

```go
ui.NumberInput(c, &app.copies, 1, 99, 1).Label("Copies")
ui.NumberInput(c, &app.opacity, 0, 1, 0.05).Label("Opacity")

ui.NumberInput(c, &app.draft, 0, 1000, 1).Width(248).OnSubmit(app.apply)
```

For a value the user only steps, use a [stepper](stepper.md); for one
chosen roughly, a [slider](slider.md); for an amount of money, which a
`float64` cannot hold exactly, a [decimal input](decimal-input.md).

## Accessibility

Assistive technology sees a spin button of the value, named by its `Label`
or its [field](form.md), which it increments and decrements by the step.
