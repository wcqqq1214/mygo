package raster

import (
	"image"

	"github.com/egoist/mygo/internal/scene"
)

// drawCanvases draws into their pixels the canvases s draws anew or shows,
// those changed since, each before the canvases and the scene that show
// it.
func drawCanvases(s *scene.Scene) {
	s.EachCanvas(func(c *scene.Image) {
		if c.PixVersion == c.Version() && len(c.Pix) == 4*c.W*c.H {
			return
		}
		cs := c.Scene
		cs.Width, cs.Height = c.W, c.H
		img := NewImage(c.W, c.H)
		var r Renderer
		r.d.draw(img, cs, image.Rect(0, 0, c.W, c.H), r.opBounds(cs))
		if len(c.Pix) != 4*c.W*c.H {
			c.Pix = make([]byte, 4*c.W*c.H)
		}
		// From BGRA to the RGBA of images.
		from, to := img.Pix, c.Pix
		for i := 0; i+3 < len(from); i += 4 {
			s, d := from[i:i+4], to[i:i+4]
			d[0], d[1], d[2], d[3] = s[2], s[1], s[0], s[3]
		}
		c.PixVersion = c.Version()
	})
}
