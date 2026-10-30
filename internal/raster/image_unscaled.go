package raster

import "github.com/egoist/mygo/internal/scene"

// imageUnscaled composites pixel-aligned images at their original size. Cached
// UI canvases commonly take this path: sampling four neighbors at each pixel
// center would read the same source pixel and do unnecessary float arithmetic.
// Scaling, subpixel placement, tinting and image corners retain the sampler.
func (r *renderer) imageUnscaled(op *scene.Op) bool {
	src, dst, img := op.Src, op.Rect, op.Image
	if op.Grayscale || op.Color != (scene.Color{}) || hasRadii(op.Radii) ||
		src.W != dst.W || src.H != dst.H || src.X < 0 || src.Y < 0 {
		return false
	}
	for _, v := range [...]float32{src.X, src.Y, dst.X, dst.Y, dst.W, dst.H} {
		if v != float32(int(v)) {
			return false
		}
	}
	sx, sy, w, h := int(src.X), int(src.Y), int(src.W), int(src.H)
	if sx > img.W || sy > img.H || w > img.W-sx || h > img.H-sy {
		return false
	}
	x0, y0, x1, y1 := r.pixelBounds(dst)
	if x0 >= x1 || y0 >= y1 {
		return true
	}
	ox, oy := sx-int(dst.X), sy-int(dst.Y)
	stride, pixels := img.W*4, img.Pix
	for y := y0; y < y1; y++ {
		at := (y+oy)*stride + (x0+ox)*4
		from := pixels[at : at+4*(x1-x0)]
		to := r.dst.Pix[y*r.dst.Stride+4*x0:][:len(from)]
		cl, ch := r.clipSolid(y)
		for i := 0; i < len(from); i += 4 {
			s := from[i : i+4]
			a := s[3]
			if a == 0 {
				continue
			}
			d := to[i : i+4]
			if x := x0 + i/4; x < cl || x >= ch {
				c := [4]float32{float32(s[0]) / 255, float32(s[1]) / 255, float32(s[2]) / 255, float32(a) / 255}
				blend(d, c, r.clipCoverage(x, y))
				continue
			}
			if a == 255 {
				d[0], d[1], d[2], d[3] = s[2], s[1], s[0], a
				continue
			}
			inv := uint32(255 - a)
			d[0], d[1], d[2], d[3] = over(s[2], d[0], inv), over(s[1], d[1], inv), over(s[0], d[2], inv), over(a, d[3], inv)
		}
	}
	return true
}
