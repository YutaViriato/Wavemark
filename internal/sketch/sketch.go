// Package sketch draws geometry as hand-drawn SVG: slightly wobbly ink
// strokes over a flat, slightly offset fill, in the spirit of the line
// illustrations on claude.com and the Claude docs (but original artwork).
//
// Output is deterministic: the same shape and seed always produce the same
// path, so pages keep stable ETags and tests can compare markup.
package sketch

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand/v2"
	"strings"
)

// Point on the 48×48 drawing grid.
type Point struct{ X, Y float64 }

// Shape is something that can be traced as a hand-drawn path.
type Shape interface {
	points(r *rand.Rand, wobble float64) (pts []Point, closed bool)
}

type line struct{ a, b Point }
type poly struct {
	pts    []Point
	closed bool
}
type ellipse struct {
	c      Point
	rx, ry float64
	a0, a1 float64 // degrees; full ellipse when a0 == a1
}
type blob struct{ pts []Point } // smooth closed curve through points
type wave struct {
	x0, x1, y, amp, period float64
}

// L is a straight line.
func L(x1, y1, x2, y2 float64) Shape { return line{Point{x1, y1}, Point{x2, y2}} }

// P is an open polyline through xy pairs.
func P(xy ...float64) Shape { return poly{pts: pairs(xy), closed: false} }

// G is a closed polygon through xy pairs.
func G(xy ...float64) Shape { return poly{pts: pairs(xy), closed: true} }

// R is a rectangle.
func R(x, y, w, h float64) Shape { return G(x, y, x+w, y, x+w, y+h, x, y+h) }

// E is a full ellipse.
func E(cx, cy, rx, ry float64) Shape { return ellipse{c: Point{cx, cy}, rx: rx, ry: ry} }

// C is a circle.
func C(cx, cy, r float64) Shape { return E(cx, cy, r, r) }

// A is an elliptical arc from a0 to a1 degrees (0 = east, 90 = south).
func A(cx, cy, rx, ry, a0, a1 float64) Shape {
	return ellipse{c: Point{cx, cy}, rx: rx, ry: ry, a0: a0, a1: a1}
}

// B is a smooth closed blob through xy pairs.
func B(xy ...float64) Shape { return blob{pts: pairs(xy)} }

// W is a horizontal sine wave.
func W(x0, x1, y, amp, period float64) Shape { return wave{x0, x1, y, amp, period} }

func pairs(xy []float64) []Point {
	pts := make([]Point, 0, len(xy)/2)
	for i := 0; i+1 < len(xy); i += 2 {
		pts = append(pts, Point{xy[i], xy[i+1]})
	}
	return pts
}

func jitter(r *rand.Rand, amount float64) float64 { return (r.Float64()*2 - 1) * amount }

// subdivide splits a segment into pieces of about step units and nudges the
// inner points perpendicular to the segment.
func subdivide(r *rand.Rand, a, b Point, step, wobble float64) []Point {
	dx, dy := b.X-a.X, b.Y-a.Y
	length := math.Hypot(dx, dy)
	n := max(1, int(math.Round(length/step)))
	nx, ny := -dy/length, dx/length
	if length == 0 {
		nx, ny = 0, 0
	}
	out := make([]Point, 0, n+1)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		p := Point{a.X + dx*t, a.Y + dy*t}
		if i > 0 && i < n {
			off := jitter(r, wobble)
			p.X += nx * off
			p.Y += ny * off
		}
		out = append(out, p)
	}
	return out
}

func (s line) points(r *rand.Rand, wobble float64) ([]Point, bool) {
	// A hand-drawn line slightly overshoots its ends.
	dx, dy := s.b.X-s.a.X, s.b.Y-s.a.Y
	length := math.Hypot(dx, dy)
	if length == 0 {
		return []Point{s.a, s.b}, false
	}
	ux, uy := dx/length, dy/length
	o0, o1 := 0.2+r.Float64()*0.4, 0.2+r.Float64()*0.4
	a := Point{s.a.X - ux*o0*wobble, s.a.Y - uy*o0*wobble}
	b := Point{s.b.X + ux*o1*wobble, s.b.Y + uy*o1*wobble}
	return []Point{a, b}, false
}

func (s poly) points(r *rand.Rand, wobble float64) ([]Point, bool) {
	out := make([]Point, 0, len(s.pts)+2)
	for _, p := range s.pts {
		out = append(out, Point{p.X + jitter(r, wobble*0.35), p.Y + jitter(r, wobble*0.35)})
	}
	if s.closed && len(out) > 1 {
		// Close past the start, like a pen finishing a shape.
		first, second := out[0], out[1]
		out = append(out, first, Point{first.X + (second.X-first.X)*0.25, first.Y + (second.Y-first.Y)*0.25})
	}
	return out, false
}

func (s ellipse) points(r *rand.Rand, wobble float64) ([]Point, bool) {
	full := s.a0 == s.a1
	start, end := s.a0, s.a1
	if full {
		start = r.Float64() * 360
		end = start + 360 + 12 + r.Float64()*14 // overlap the start
	}
	span := end - start
	n := max(8, int(math.Abs(span)/18))
	out := make([]Point, 0, n+1)
	for i := 0; i <= n; i++ {
		a := (start + span*float64(i)/float64(n)) * math.Pi / 180
		k := 1 + jitter(r, wobble*0.035)
		out = append(out, Point{s.c.X + math.Cos(a)*s.rx*k, s.c.Y + math.Sin(a)*s.ry*k})
	}
	return out, false
}

func (s blob) points(r *rand.Rand, wobble float64) ([]Point, bool) {
	out := make([]Point, len(s.pts))
	for i, p := range s.pts {
		out[i] = Point{p.X + jitter(r, wobble*0.5), p.Y + jitter(r, wobble*0.5)}
	}
	return out, true
}

func (s wave) points(r *rand.Rand, wobble float64) ([]Point, bool) {
	var out []Point
	for x := s.x0; x <= s.x1+0.01; x += s.period / 6 {
		y := s.y + math.Sin((x-s.x0)/s.period*2*math.Pi)*s.amp + jitter(r, wobble*0.25)
		out = append(out, Point{x, y})
	}
	return out, false
}

func f(v float64) string {
	return strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
}

// smoothPath turns points into a quadratic curve through segment midpoints.
// Closed paths use a Catmull-Rom style loop.
func smoothPath(pts []Point, closed bool) string {
	if len(pts) < 2 {
		return ""
	}
	var b strings.Builder
	if closed {
		n := len(pts)
		mid := func(i int) Point {
			a, c := pts[(i+n)%n], pts[(i+1+n)%n]
			return Point{(a.X + c.X) / 2, (a.Y + c.Y) / 2}
		}
		m := mid(0)
		fmt.Fprintf(&b, "M%s %s", f(m.X), f(m.Y))
		for i := 1; i <= n; i++ {
			p, m := pts[i%n], mid(i)
			fmt.Fprintf(&b, "Q%s %s %s %s", f(p.X), f(p.Y), f(m.X), f(m.Y))
		}
		b.WriteString("Z")
		return b.String()
	}
	fmt.Fprintf(&b, "M%s %s", f(pts[0].X), f(pts[0].Y))
	if len(pts) == 2 {
		fmt.Fprintf(&b, "L%s %s", f(pts[1].X), f(pts[1].Y))
		return b.String()
	}
	for i := 1; i < len(pts)-1; i++ {
		p, q := pts[i], pts[i+1]
		mx, my := (p.X+q.X)/2, (p.Y+q.Y)/2
		if i == len(pts)-2 {
			mx, my = q.X, q.Y
		}
		fmt.Fprintf(&b, "Q%s %s %s %s", f(p.X), f(p.Y), f(mx), f(my))
	}
	return b.String()
}

func isEdgy(s Shape) bool {
	switch s.(type) {
	case line, poly:
		return true
	}
	return false
}

// edgePath joins vertices with slightly bowed edges, keeping corners sharp.
func edgePath(r *rand.Rand, pts []Point, wobble float64) string {
	var b strings.Builder
	fmt.Fprintf(&b, "M%s %s", f(pts[0].X), f(pts[0].Y))
	for i := 1; i < len(pts); i++ {
		a, c := pts[i-1], pts[i]
		dx, dy := c.X-a.X, c.Y-a.Y
		length := math.Hypot(dx, dy)
		if length < 0.01 {
			continue
		}
		bow := jitter(r, wobble) * min(1, length/14)
		mx, my := (a.X+c.X)/2-dy/length*bow, (a.Y+c.Y)/2+dx/length*bow
		fmt.Fprintf(&b, "Q%s %s %s %s", f(mx), f(my), f(c.X), f(c.Y))
	}
	return b.String()
}

// Drawing is a hand-drawn picture: filled regions under ink strokes.
type Drawing struct {
	Fill []Shape
	Ink  []Shape
}

// Seed derives a stable seed from a name.
func Seed(name string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return h.Sum64()
}

// Options for rendering a drawing.
type Options struct {
	Size    int    // rendered width and height in px (0 = inherit 1em)
	Class   string // extra classes
	Label   string // accessible label; empty means decorative
	Wobble  float64
	ViewBox string // default "0 0 48 48"
	Stretch bool   // preserveAspectRatio="none", for underlines
}

// SVG renders d as an inline SVG element. Fill shapes are drawn with a small
// offset, like a print slightly out of register.
func (d Drawing) SVG(seed uint64, o Options) string {
	r := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	wob := o.Wobble
	if wob == 0 {
		wob = 0.9
	}
	vb := o.ViewBox
	if vb == "" {
		vb = "0 0 48 48"
	}
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="` + vb + `" class="wm-sketch`)
	if o.Class != "" {
		b.WriteString(" " + o.Class)
	}
	b.WriteString(`"`)
	if o.Size > 0 {
		fmt.Fprintf(&b, ` width="%d" height="%d"`, o.Size, o.Size)
	}
	if o.Stretch {
		b.WriteString(` preserveAspectRatio="none"`)
	}
	if o.Label != "" {
		b.WriteString(` role="img" aria-label="` + escapeAttr(o.Label) + `"`)
	} else {
		b.WriteString(` aria-hidden="true" focusable="false"`)
	}
	b.WriteString(`>`)
	if len(d.Fill) > 0 {
		b.WriteString(`<g class="wm-sketch__fill" transform="translate(1.3 1.1)">`)
		for _, s := range d.Fill {
			pts, _ := s.points(r, wob*0.6)
			path := smoothPath(pts, true)
			if isEdgy(s) {
				path = edgePath(r, pts, wob*0.6) + "Z"
			}
			b.WriteString(`<path d="` + path + `"/>`)
		}
		b.WriteString(`</g>`)
	}
	b.WriteString(`<g class="wm-sketch__ink">`)
	for _, s := range d.Ink {
		pts, closed := s.points(r, wob)
		path := smoothPath(pts, closed)
		if isEdgy(s) {
			path = edgePath(r, pts, wob)
		}
		b.WriteString(`<path pathLength="1" d="` + path + `"/>`)
	}
	b.WriteString(`</g></svg>`)
	return b.String()
}

func escapeAttr(s string) string {
	return strings.NewReplacer(`&`, "&amp;", `"`, "&quot;", `<`, "&lt;", `>`, "&gt;").Replace(s)
}
