package markdown

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestThemeWindowMutations(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*ui.Theme)
	}{
		{"dark", func(th *ui.Theme) { th.Dark = !th.Dark }},
		{"background", func(th *ui.Theme) { th.Background = ui.RGB(1, 2, 3) }},
		{"text", func(th *ui.Theme) { th.Text = ui.RGB(2, 3, 4) }},
		{"muted", func(th *ui.Theme) { th.TextMuted = ui.RGB(3, 4, 5) }},
		{"border", func(th *ui.Theme) { th.Border = ui.RGB(4, 5, 6) }},
		{"accent", func(th *ui.Theme) { th.Accent = ui.RGB(5, 6, 7) }},
		{"selection", func(th *ui.Theme) { th.Selection = ui.RGB(6, 7, 8) }},
		{"radius", func(th *ui.Theme) { th.Radius++ }},
		{"font size", func(th *ui.Theme) { th.FontSize++ }},
		{"font", func(th *ui.Theme) { th.Font = "monospace" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			th := ui.LightTheme()
			st := newState()
			st.o = &Options{}
			tt := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(th)
				st.theme(c)
			}, 300, 100)
			def := st.def
			// The same theme object changes in place, so pointer identity
			// cannot establish whether the Markdown defaults are current.
			st.paras = []paraCache{{ok: true}}
			cs := &codeState{chunks: []codeChunk{{ok: true}}}
			st.codes = map[int]*codeState{0: cs}
			tc.mutate(th)
			tt.Frame()
			want := DefaultTheme(th)
			got := st.th
			if *got.CodeColors != *want.CodeColors {
				t.Error("code defaults did not follow the changed window theme")
			}
			got.CodeColors = want.CodeColors
			if got != *want {
				t.Error("Markdown defaults did not follow the changed window theme")
			}
			if st.def == def || st.paras[0].ok || len(cs.chunks) != 0 {
				t.Fatal("theme mutation did not invalidate the cached defaults, paragraphs and code chunks")
			}
			def = st.def
			st.paras[0].ok = true
			cs.chunks = append(cs.chunks, codeChunk{ok: true})
			tt.Frame()
			if st.def != def || !st.paras[0].ok || len(cs.chunks) != 1 {
				t.Fatal("unchanged theme invalidated caches")
			}
		})
	}
}

func TestThemeTooltipCallbacks(t *testing.T) {
	th := ui.LightTheme()
	th.TooltipPolicy = &ui.TooltipPolicy{Delay: time.Millisecond}
	// Both styles come from the same function literal, with different
	// captures. Their behavior must follow the current callback.
	style := func(label string) func(ui.Element) {
		return func(tip ui.Element) {
			tip.Label(label).Padding(6).Background(th.Background)
		}
	}
	th.TooltipStyle = style("first style")
	var st *viewState
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(th)
		root := View(c, "A **paragraph**.\n\n```go\nvar n = 1\n```", nil)
		st = *ui.Local(root, stateKey{}, newState)
		ui.Button(c, "Hint").Tooltip("A hint")
	}, 400, 300)
	def := st.def
	paragraphSpan := &st.paras[1].p.pieces[0].spans[0]
	var code *codeState
	for _, cs := range st.codes {
		code = cs
	}
	codeSpan := &code.chunks[0].spans[0]
	checkCaches := func() {
		t.Helper()
		if st.def != def || &st.paras[1].p.pieces[0].spans[0] != paragraphSpan || &code.chunks[0].spans[0] != codeSpan {
			t.Fatal("tooltip callback change rebuilt Markdown defaults, paragraph spans or code spans")
		}
	}
	r, ok := tt.Find("Hint")
	if !ok {
		t.Fatal("missing tooltip anchor")
	}
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Advance(time.Millisecond)
	if _, ok := tt.Find("first style"); !ok {
		t.Fatal("first tooltip callback did not style the hint")
	}
	for range 3 {
		tt.Frame()
		checkCaches()
	}
	th.TooltipStyle = style("second style")
	tt.Frame()
	if _, ok := tt.Find("second style"); !ok {
		t.Fatal("tooltip did not use the replacement closure")
	}
	checkCaches()
	th.TooltipStyle = nil
	tt.Frame()
	if !tt.HasText("A hint") || st.def != def {
		t.Fatal("clearing the callback lost the hint or invalidated Markdown defaults")
	}
	checkCaches()
}
