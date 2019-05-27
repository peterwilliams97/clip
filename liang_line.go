package clip

import (
	"fmt"
	"math"
)

// Point is a 2-d point.
type Point struct {
	X, Y float64
}

func (p Point) add(d Point) Point {
	return Point{p.X + d.X, p.Y + d.Y}
}

func (p Point) sub(d Point) Point {
	return Point{p.X - d.X, p.Y - d.Y}
}

func (p Point) mul(g float64) Point {
	return Point{p.X * g, p.Y * g}
}

func (p Point) isZero() bool {
	return isZero(p.X) && isZero(p.Y)
}

// Equals returns true if `p` and `d` are in the same location.
func (p Point) Equals(d Point) bool {
	return p.sub(d).isZero()
}

type Line struct {
	A, B Point
}

// Position returns the parametrized point.
// p = a ∙ (1 - t) + b ∙ t = a + (b - a) ∙ t
func (l Line) Position(t float64) Point {
	a, b := l.A, l.B
	d := b.sub(a)
	return a.add(d.mul(t))
}

// NewLine returns the line from (ax, ay) to (bx, by).
func NewLine(ax, ay, bx, by float64) Line {
	return Line{Point{ax, ay}, Point{bx, by}}
}

// Equals returns true if `l` and `d` are in the same location.
func (l Line) Equals(d Line) bool {
	return l.A.Equals(d.A) && l.B.Equals(d.B)
}

// Rect is a rectangle.
type Rect struct {
	Llx, Lly, Urx, Ury float64
}

type liangBarsky struct {
	Rect
}

// NewLiangBarsky returns a liangBarsky with clip rectangle `window`.
func NewLiangBarsky(window Rect) liangBarsky {
	return liangBarsky{window}
}

// interval is an interval on a line (a, b) parametrized by p(t) = a ∙ (1 - t) + b ∙ t
// i.tE <= t <= i.tL for the interval
type interval struct {
	tE float64 // Value of t where it enters the clipping window.
	tL float64 // Value of t where it leaves the clipping window.
}

// newInterval returns the t range for new a clipping interval on a line.
// This must be 0-1 because a clipped line is p(t) = a ∙ (1 - t) + b ∙ t
// i.e. t=0 -> p=a
//      t=1 -> p=b
func newInterval() interval {
	return interval{0, 1}
}

// ClipLine clips the line between `a` and `b` to the rectangular window in `l`.
// Parametrized point  p = a ∙ (1 - t) + b ∙ t for i.tE <= t <= i.tL
func (l liangBarsky) ClipLine(line Line) (Line, bool) {
	a, b := line.A, line.B
	d := b.sub(a)

	if d.isZero() && l.inside(a) {
		return Line{a, b}, true
	}
	i := newInterval()
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
func (i *interval) clipRange(ll, ur, a, d float64) bool {
	return i.clipT(ll-a, d) && i.clipT(a-ur, -d)
}

// clipT tests t=`a`/`d` for insideness in `tE`, `tL`
// tE <= t <= tL : inside
// Enter test: tE -> t
// Leave test:tL -> t
func (i *interval) clipT(a, d float64) bool {
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

// inside returns true if `a` is inside window `l`.
func (l liangBarsky) inside(a Point) bool {
	return l.Llx-tol <= a.X && a.X <= l.Urx+tol &&
		l.Lly-tol <= a.Y && a.Y <= l.Ury+tol
}

// inside returns true if all points on `line` are inside window `l`.
func (l liangBarsky) LineInside(line Line) bool {
	return l.inside(line.A) && l.inside(line.B)
}

func isZero(a float64) bool {
	return math.Abs(a) < tol
}

// tol is the tolerance on all measurements
const tol = 0.000001 * 0.000001
