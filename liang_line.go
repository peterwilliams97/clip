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

func (p Point) project(d Point, t float64) Point {
	// fmt.Printf("project: p  =%+v d=%+v g=%g\n", p, d, g)
	// dg := d.mul(g)
	// pdg := p.add(dg)
	// fmt.Printf("project: dg =%+v\n", dg)
	// fmt.Printf("project: pdg=%+v\n", pdg)
	// fmt.Printf("Position: l=%+v\n", l)
	o := p.add(d.mul(t))
	// fmt.Printf("\tproject: p  =%+v d=%+v t=%g => %+v \n", p, d, t, o)
	// if t < 0.0 || t > 1.0 {
	// 	panic("t")
	// }
	return o
}

// func project(a, d Point, g float64) Point {
// 	return a.Add(d.Mul(g))
// 	return Point {
// 		a.X + g * d.X,
// 		a.Y + g * d.Y,
// 	}
//  }

func (p Point) isZero() bool {
	return isZero(p.X) && isZero(p.Y)
}

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
	// fmt.Printf("Position: l=%+v\n", l)
	return a.project(d, t)
	return a.add(d.mul(t))
}

func NewLine(ax, ay, bx, by float64) Line {
	return Line{Point{ax, ay}, Point{bx, by}}
}
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

// Parametrized point
// a ∙ (1 - t) + b ∙ t
type interval struct {
	tE float64 // Value of t where it enters the clipping window.
	tL float64 // Value of t where it leaves the clipping window.
}

// newInterval returns t range for new a clipping window.
// This must be 0-1 because clipped line is x = a ∙ (1 - t) + b ∙ t
// i.e. t=0 -> x=a
//      t=1 -> x=b
func newInterval() interval {
	return interval{0, 1}
}

// ClipLine clips the line between `a` and `b` to the rectangular window in `l`.
// Parametrized point
// a ∙ (1 - t) + b ∙ t
func (l liangBarsky) ClipLine(line Line) (Line, bool) {
	a, b := line.A, line.B
	d := b.sub(a)

	if d.isZero() && l.inside(a) {
		return Line{a, b}, true
	}
	// fmt.Printf("line=%+v\n", line)
	i := newInterval()
	if !(i.clipRange(l.Llx, l.Urx, a.X, d.X) && // horizonal
		i.clipRange(l.Lly, l.Ury, a.Y, d.Y)) { // vertical
		return Line{}, false
	}
	if i.tE > 0.0 {
		a = a.project(d, i.tE)
		a2 := line.Position(i.tE)
		if !a2.Equals(a) {
			panic("a-a2")
		}
		// line.A = a
		// fmt.Printf("line=%+v i=%+v\n", line, i)
	}
	if i.tL < 1.0 {
		// fmt.Println("======================")
		// fmt.Printf("a,b=%+v,%+v\n", a, b)
		// fmt.Printf("line=%+v i=%+v\n", line, i)
		b = a.project(d, i.tL)
		b2 := line.Position(i.tL + i.tE)
		if !b2.Equals(b) {
			fmt.Printf("b =%+v\n", b)
			fmt.Printf("b2=%+v\n", b2)
			panic("b-b2")
		}
	}
	return Line{a, b}, true
}

func (i *interval) clipRange(l, r, a, d float64) bool {
	return i.clipT(l-a, d) && i.clipT(a-r, -d)
}

// clipT tests t=`num`/`denom` for insideness in `tE`, `tL`
// tE <= t <= tL : inside
// Enter test: tE -> t
// Leave test:tL -> t
func (i *interval) clipT(num, denom float64) bool {
	if isZero(denom) {
		return num <= 0.0
	}

	t := num / denom

	if denom > 0.0 {
		// Enter test (lower x,y)
		if t > i.tL {
			return false
		}
		if t > i.tE {
			i.tE = t
		}
	} else {
		// Leave test (upper x,y)
		if t < i.tE {
			return false
		}
		if t < i.tL {
			i.tL = t
		}
	}
	return true
}

func (l liangBarsky) inside(a Point) bool {
	return l.Llx <= a.X && a.X <= l.Urx &&
		l.Lly <= a.Y && a.Y <= l.Ury
}

func isZero(a float64) bool {
	return math.Abs(a) < tol
}

const tol = 0.000001 * 0.000001

//  t=1.0064107158336175
