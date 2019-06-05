package clip

import (
	"fmt"
	"math"
)

// liangBarsky is a Liang-Barsky clipper.
type liangBarsky struct {
	Rect
}

// NewLiangBarsky returns a liangBarsky with clip rectangle `window`.
func NewLiangBarsky(window Rect) liangBarsky {
	return liangBarsky{window}
}

// tInterval is an tInterval on a line (a, b) parametrized by p(t) = a ∙ (1 - t) + b ∙ t
// i.tE <= t <= i.tL for the tInterval
type tInterval struct {
	tE float64 // Value of t where it enters the clipping window.
	tL float64 // Value of t where it leaves the clipping window.
}

// newTInterval returns the t range for new a clipping tInterval on a line.
// This must be 0-1 because a clipped line is p(t) = a ∙ (1 - t) + b ∙ t
// i.e. t=0 -> p=a
//      t=1 -> p=b
func newTInterval() tInterval {
	return tInterval{0, 1}
}

// ClipLine clips the line between `a` and `b` to the rectangular window in `l`.
// Parametrized point  p = a ∙ (1 - t) + b ∙ t for i.tE <= t <= i.tL
func (l liangBarsky) ClipLine(line Line) (Line, bool) {
	a, b := line.A, line.B
	d := b.sub(a)

	if d.isZero() && l.inside(a) {
		return Line{a, b}, true
	}
	i := newTInterval()
	if !(i.clipRange(l.Llx, l.Urx, a.X, d.X) && // horizonal
		i.clipRange(l.Lly, l.Ury, a.Y, d.Y)) { // vertical
		return Line{}, false
	}
	a = line.Position(i.tE)
	b = line.Position(i.tL)
	if !l.inside(a) {
		panic(fmt.Errorf("a=%+v outside lb=%+v", a, l))
	}
	if !l.inside(b) {
		panic(fmt.Errorf("b=%+v outside lb=%+v", b, l))
	}
	return Line{a, b}, true
}

// clipT tests t=`a`/`d` for insideness in `tE` <= t*`d` <= `tL` betweem
// tE <= t <= tL : inside
// Enter test: tE -> t
// Leave test:tL -> t
func (i *tInterval) clipRange(ll, ur, a, d float64) bool {
	return i.clipT(ll-a, d) && i.clipT(a-ur, -d)
}

// clipT tests t=`a`/`d` for insideness in `tE`, `tL`
// tE <= t <= tL : inside
// Enter test: tE -> t
// Leave test:tL -> t
func (i *tInterval) clipT(a, d float64) bool {
	if isZero(d) {
		return a <= 0.0
	}

	t := a / d

	if d > 0.0 {
		// Enter test (lower left x,y)
		if t > i.tL {
			return false
		}
		if t > i.tE {
			i.tE = t
		}
	} else {
		// Leave test (upper right x,y)
		if t < i.tE {
			return false
		}
		if t < i.tL {
			i.tL = t
		}
	}
	return true
}

const infinity = math.MaxFloat64

// ClipPolygon returns `path` clipped to the `l` rectangular window,
func (l liangBarsky) ClipPolygon(path []Point) []Point {
	n := len(path)

	// Avoid special case for last edge
	path = append(path, path[0])

	clipped := make([]Point, 0, n)

	ll := Point{l.Llx, l.Lly}
	ur := Point{l.Urx, l.Ury}

	var in, out Point              // Coordinates of entry and exit points.
	var tOut1, tIn2, tOut2 float64 // Parameter values of same.
	var tIn, tOut Point            // Parameter values for intersection.

	var o Point // The next point to be added

	for i := 0; i < n; i++ { // for each edge
		p0 := path[i]
		p1 := path[i+1]
		delta := p1.sub(p0)

		// use this to determine which bounding lines for the clip region the
		// containing line hits first
		if delta.X > 0 || (isZero(delta.X) && p0.X > ur.X) {
			in.X, out.X = ll.X, ur.X
		} else {
			in.X, out.X = ur.X, ll.X
		}
		if delta.Y > 0 || (isZero(delta.Y) && p0.Y > ur.Y) {
			in.Y, out.Y = ll.Y, ur.Y
		} else {
			in.Y, out.Y = ur.Y, ll.Y
		}

		// find the t values for the x and y exit points
		if !isZero(delta.X) {
			tOut.X = (out.X - p0.X) / delta.X
		} else if ll.X <= p0.X && p0.X <= ur.X {
			tOut.X = infinity
		} else {
			tOut.X = -infinity
		}
		if !isZero(delta.Y) {
			tOut.Y = (out.Y - p0.Y) / delta.Y
		} else if ll.Y <= p0.Y && p0.Y <= ur.Y {
			tOut.Y = infinity
		} else {
			tOut.Y = -infinity
		}

		// Order the two exit points
		if tOut.X < tOut.Y {
			tOut1, tOut2 = tOut.X, tOut.Y
		} else {
			tOut1, tOut2 = tOut.Y, tOut.X
		}

		if tOut2 > 0 {
			if !isZero(delta.X) {
				tIn.X = (in.X - p0.X) / delta.X
			} else {
				tIn.X = -infinity
			}

			if !isZero(delta.Y) {
				tIn.Y = (in.Y - p0.Y) / delta.Y
			} else {
				tIn.Y = -infinity
			}
			if tIn.X < tIn.Y {
				tIn2 = tIn.Y
			} else {
				tIn2 = tIn.X
			}
			if tOut1 < tIn2 { // no visible segment
				if 0 < tOut1 && tOut1 <= 1 {
					// line crosses over intermediate corner region
					if tIn.X < tIn.Y {
						o = Point{out.X, in.Y}
					} else {
						o = Point{in.X, out.Y}
					}
					clipped = append(clipped, o)
				}
			} else {

				// line crosses though window
				if 0 < tOut1 && tIn2 <= 1 {
					if 0 <= tIn2 { // visible segment
						o = Point{in.X, p0.Y + tIn.X*delta.Y}
						if tIn.X > tIn.Y {
							o = Point{in.X, p0.Y + tIn.X*delta.Y}
						} else {
							o = Point{p0.X + tIn.Y*delta.X, in.Y}
						}
						clipped = append(clipped, o)
					}

					if tOut1 <= 1 {
						if tOut.X < tOut.Y {
							o = Point{out.X, p0.Y + tOut.X*delta.Y}
						} else {
							o = Point{p0.X + tOut.Y*delta.X, out.Y}
						}
						clipped = append(clipped, o)
					} else {
						clipped = append(clipped, p1)
					}
				}
			}

			if 0 < tOut2 && tOut2 <= 1 {
				o = Point{out.X, out.Y}
				clipped = append(clipped, o)
			}
		}
	}

	return trim(removeRepeats(clipped))
}

// removeRepeats returns `path` with repeated points removed.
func removeRepeats(path []Point) []Point {
	if len(path) == 0 {
		return []Point{}
	}
	out := make([]Point, 1, len(path))
	out[0] = path[0]
	for _, p := range path[1:] {
		if !p.Equals(out[len(out)-1]) {
			out = append(out, p)
		}
	}
	return out
}

// trim returns a copy of `path` with minimal backing memory.
func trim(path []Point) []Point {
	out := make([]Point, len(path))
	copy(out, path)
	return out
}

// inside returns true if `a` is inside window `l`.
func (l liangBarsky) inside(a Point) bool {
	return l.Llx-tol <= a.X && a.X <= l.Urx+tol &&
		l.Lly-tol <= a.Y && a.Y <= l.Ury+tol
}

// inside returns true if all points on `line` are inside window `l`.
func (l liangBarsky) LineInside(line Line) bool {
	return l.inside(line.A) && l.inside(line.B)
}
