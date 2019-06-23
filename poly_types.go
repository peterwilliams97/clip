package clip

import (
	"fmt"
	"strings"

	"github.com/unidoc/unipdf/v3/common"
)

/*
	Polygon internal types
	Coordinate origin is top-left

	Vertex: Vertex on a contour.
	Side:   Edge between 2 vertices. Either horizontal or vertical

*/

// Rectilinear is an interface for vertical and horizontal lines. These can be sides, chords or
// diagonals of rectilinear (orthogonal) polygons.
// https://en.wikipedia.org/wiki/Rectilinear_polygon
type Rectilinear interface {
	// X0X1YVert returns x0, x1, y, vertical where:
	// - vertical is true (false) for vertical (horizontal) chords
	// - x0, x1 = width in the direction of the chord
	// - y = coordinate in the other direction
	X0X1YVert() (float64, float64, float64, bool)
}

func rectString(r Rectilinear) string {
	x0, x1, y, vertical := r.X0X1YVert()
	return fmt.Sprintf("%v %g %s", Point{x0, x1}, y, directionName(vertical))
}

// directionName returns the name of the direction.
func directionName(vertical bool) string {
	if vertical {
		return "`vertical` "
	}
	return "`horizontal`"
}

// toLine returns the line along `r`.
func toLine(r Rectilinear) Line {
	x0, x1, y, vertical := r.X0X1YVert()
	if !vertical {
		return Line{Point{y, x0}, Point{y, x1}}
	}
	return Line{Point{x0, y}, Point{x1, y}}
}

// Vertex is a vertex on a rectilinear polygon.
type Vertex struct {
	Point
	iContour int     // Index of countour in polygon.
	index    int     // Index of vertex in contour.
	prev     *Vertex // Previous vertex in contour (or polygon?) !@#$
	next     *Vertex // Next vertex in contour (or polygon?) !@#$
	concave  bool    // True if vertex is concave.
	// visited  bool    // Does this belong here?
}

func NewVertex(point Point, index int, prev, next *Vertex) *Vertex {
	v := Vertex{
		Point: point,
		index: index,
		prev:  prev,
		next:  next,
	}
	v.Validate()
	return &v
}

func (v *Vertex) Validate() {
	if v == nil {
		return
	}
	if v.prev != nil && v.prev.Point.Equals(v.Point) {
		common.Log.Error("\n\tprev=%#v\n\t   v=%#v", *v.prev, *v)
		panic(fmt.Errorf("duplicate point: prev v=%v", *v))
	}
	if v.next != nil && v.next.Point.Equals(v.Point) {
		common.Log.Error("\n\t   v=%#v\n\tnext=%#v", *v, *v.next)
		panic(fmt.Errorf("duplicate point: next v=%v", *v))
	}
}

func (v Vertex) id() string {
	return fmt.Sprintf("%v", v.Point)
}

func (v Vertex) String() string {
	sp, sn := "(nil)", "(nil)"
	if v.prev != nil {
		sp = fmt.Sprintf("%g", v.prev.Point)
	}
	if v.next != nil {
		sn = fmt.Sprintf("%g", v.next.Point)
	}
	// return fmt.Sprintf("VERTEX{point:%+g index:%d prev:%p %v next:%p %v concave:%t visited:%t}",
	// 	v.Point, v.index, v.prev, sp, v.next, sn, v.concave, v.visited)
	concave := "convex "
	if v.concave {
		concave = "CONCAVE"
	}
	return fmt.Sprintf("VERTEX{index:%2d %v->{%g}->%v %s}",
		v.index, sp, v.Point, sn, concave)
}

func (v *Vertex) Join(prev, next *Vertex) {
	v.prev = prev
	v.next = next
}

// Side is a vertical or horizontal edge of a contour..
type Side struct { // A side? !@#$
	x0, x1       float64 // Start and end of the interval in the vertical or horizontal direction.
	y            float64 // Coordinate in the other directon to x0, x1
	start, end   *Vertex // Vertices at the start and end of the segment.
	vertical     bool    // Is this a vertical segment?
	number       int
	pStart, pEnd Point
}

func (s *Side) X0X1YVert() (x0, x1, y float64, vertical bool) {
	return s.x0, s.x1, s.y, s.vertical
}

func NewSide(start, end *Vertex) *Side {
	start.Validate()
	end.Validate()
	if start.X == end.X && start.Y == end.Y {
		panic("duplicate point")
	}
	if start.X != end.X && start.Y != end.Y {
		panic("diagonal side")
	}
	vertical := start.X == end.X

	var x0, x1, y float64
	if vertical {
		x0, x1, y = start.Y, end.Y, start.X
		// common.Log.Info("vertical=%t x0=%.1f x1=%.1f y=%.1f x0==x1=%t  x1-x0=%g",
		// 	vertical, x0, x1, y, x0 == x1, x1-x0)
		if x0 == x1 {
			common.Log.Error("\n\tstart=%v\n\t  end=%v", start, end)
			panic("not allowed")
		}
	} else {
		x0, x1, y = start.X, end.X, start.Y
		// common.Log.Info("vertical=%t x0=%.1f x1=%.1f y=%.1f x0==x1=%t  x1-x0=%g",
		// 	vertical, x0, x1, y, x0 == x1, x1-x0)
		if x0 == x1 {
			common.Log.Error("\n\tstart=%v\n\t  end=%v", start, end)
			panic("not allowed")
		}
	}
	// common.Log.Info("vertical=%t x0=%.1f x1=%.1f y=%.1f x0==x1=%t  x1-x0=%g",
	// 	vertical, x0, x1, y, x0 == x1, x1-x0)
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	// common.Log.Info("vertical=%t x0=%.1f x1=%.1f y=%.1f x0==x1=%t  x1-x0=%g",
	// 	vertical, x0, x1, y, x0 == x1, x1-x0)
	if x0 == x1 {
		common.Log.Error("\n\tstart=%v\n\t  end=%v", start, end)
		panic("not allowed")
	}

	return &Side{
		x0:       x0,
		x1:       x1,
		y:        y,
		start:    start,
		end:      end,
		vertical: vertical,
		number:   -1,
		pStart:   start.Point,
		pEnd:     end.Point,
	}
}

// Chord is a chord from `v` to `s`.
//   horizontal    vertical
//    x0   x1           y
//         |      x0    v
//   y >---+            |
//         |      x1 ---+----
type Chord struct {
	v *Vertex
	s *Side
}

func (c Chord) String() string {
	return fmt.Sprintf("CHORD{%s v=%v s=%v %v}",
		rectString(&c), c.v.Point, c.s.start.Point, c.s.end.Point)
}

// !@#$ For testing
func NewChord(v Point, s Line) *Chord {
	vert := s.A.X == s.B.X
	horz := s.A.Y == s.B.Y
	if vert == horz {
		panic("bad chord")
	}

	return &Chord{
		v: &Vertex{Point: v},
		s: NewSide(&Vertex{Point: s.A}, &Vertex{Point: s.B}),
	}
}

// X0X1YVert returns x0, x1, y, vertical where:
// - vertical is true (false) for vertical (horizontal) chords
// - x0, x1 = width in the direction of the chord
// - y = coordinate in the other direction
func (c *Chord) X0X1YVert() (x0, x1, y float64, vertical bool) {
	vertical = !c.s.vertical
	if vertical {
		x0, x1, y = c.v.Y, c.s.start.Y, c.v.X
	} else {
		x0, x1, y = c.v.X, c.s.start.X, c.v.Y
	}
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	return
}

// OtherEnd returns the coordinates of the intersection of the chord with the segment.
func (c *Chord) OtherEnd() Point {
	vertical := !c.s.vertical
	if vertical {
		return Point{c.v.X, c.s.start.Y}
	}
	return Point{c.s.start.X, c.v.Y}
}

// Intersects returns true if `c` intersects `s`.
func (c *Chord) Intersects(s *Side) bool {
	vertical := !c.s.vertical
	if vertical == s.vertical {
		return false
	}
	var x0, x1, y float64
	var sx, sy0, sy1 float64
	if vertical {
		x0, x1, y = c.v.Y, c.s.start.Y, c.v.X
		sx, sy0, sy1 = s.start.Y, s.start.X, s.end.X
	} else {
		x0, x1, y = c.v.X, c.s.start.X, c.v.Y
		sx, sy0, sy1 = s.start.X, s.start.Y, s.end.Y
	}
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if sy0 > sy1 {
		sy0, sy1 = sy1, sy0
	}
	return x0 <= sx && sx <= x1 && sy0 <= y && y <= sy1
}

// !@#$
func integerizePoly(poly []Path) []Path {
	for i, path := range poly {
		poly[i] = path.integerize()
	}
	return poly
}

func getDirection(x0, x1 float64) string {
	if x0 < x1 {
		return "<"
	}
	if x0 > x1 {
		return ">"
	}
	return "="
}

func showContour(contour []*Vertex) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d vertices\n", len(contour))
	for i, v := range contour[:len(contour)-1] {
		fmt.Fprintf(&b, "\t%2d %v %p->>%p>->%p\n", i, v, v.prev, v, v.next)
	}
	i := len(contour) - 1
	v := contour[i]
	fmt.Fprintf(&b, "\t%2d %v %p->{%p}->%p", i, v, v.prev, v, v.next)
	return b.String()
}

func validateCountour(contour []*Vertex) {
	v0 := contour[0]
	v := v0
	for i := 0; i < len(contour); i++ {
		if v.prev == nil || v.next == nil {
			common.Log.Error("Bad vertex: prev=%p next=%p \n\tv=%v\n\tcontour=%s",
				v.prev, v.next, v, showContour(contour))
			panic("bad vertex")
		}
		if v.prev.next != v || v.next.prev != v {
			common.Log.Error("Bad links: prev=%p next=%p \n\tv=%v\n\tcontour=%s",
				v.prev, v.next, v, showContour(contour))
			panic("bad links")
		}
		v = v.next

	}
	if v != v0 {
		common.Log.Error("Not closed: prev=%p next=%p \n\tv=%v\n\tcontour=%s",
			v.prev, v.next, v, showContour(contour))
		panic("not closed")
	}
}

func rebuildCountour(contour []*Vertex) []*Vertex {
	validateCountour(contour)
	newContour := make([]*Vertex, 0, len(contour))

	v := contour[0]
	prev := v.prev
	for i := 0; i < len(contour); i++ {
		next := v.next
		common.Log.Info("^^ i=%d index=%4d %v {%v} %v X=%5t Y=%5t",
			i, v.index, prev.Point, v.Point, next.Point,
			prev.X == v.X && v.X == next.X,
			prev.Y == v.Y && v.Y == next.Y)
		if !(prev.X == v.X && v.X == next.X) && !(prev.Y == v.Y && v.Y == next.Y) {
			newContour = append(newContour, cv(v))
			prev = v
		} else {
			common.Log.Info("$$ skip %v", v)
		}
		v = next
	}
	for i := 0; i < len(newContour); i++ {
		newContour[i].index = i
	}
	for i0 := 0; i0 < len(newContour); i0++ {
		i1 := (i0 + 1) % len(newContour)
		newContour[i0].next = newContour[i1]
		newContour[i1].prev = newContour[i0]
	}

	common.Log.Info("rebuildContour: %d vertices -> %d vertices", len(contour), len(newContour))
	// if v != contour[0] {
	// 	panic(">>not closed<<")
	// }

	common.Log.Info("counter=%s", showContour(contour))
	common.Log.Info("newContour=%s", showContour(newContour))
	validateCountour(newContour)
	if len(newContour) != 4 {
		panic("!=4 vertices")
	}

	return newContour
}

// cv returns a pointer to a copy of `v`.
func cv(v *Vertex) *Vertex {
	if v == nil {
		panic("nil pointer")
		return nil
	}
	w := *v
	return &w
}