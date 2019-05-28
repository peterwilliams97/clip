package clip

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

// Line is a straight line.
type Line struct {
	A, B Point
}

// Path is a path that is not necessarily closed.
type Path []Point

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
