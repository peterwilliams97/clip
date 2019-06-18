package clip

import (
	"fmt"

	"github.com/unidoc/unipdf/common"
)

/*
	Polygon internal types
	Coordinate origin is top-left

	Vertex: Vertex on a contour.
	Side:   Edge between 2 vertices. Either horizontal or vertical

*/

type Rectilinear interface {
	X0X1YVert() (float64, float64, float64, bool)
}

func rectString(r Rectilinear) string {
	x0, x1, y, vertical := r.X0X1YVert()
	direct := "horizontal"
	if vertical {
		direct = "vertical"
	}
	return fmt.Sprintf("[%g,%g] %g %#q", x0, x1, y, direct)
}

// Vertex is a vertex on a rectilinear polygon.
type Vertex struct {
	Point
	iContour int     // Index of countour in polygon.
	index    int     // Index of vertex in contour.
	prev     *Vertex // Previous vertex in contour (or polygon?) !@#$
	next     *Vertex // Next vertex in contour (or polygon?) !@#$
	concave  bool    // True if vertex is concave.
	visited  bool    // Does this belong here?
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
	return fmt.Sprintf("VERTEX{index:%d prev:%v point:%g next:%v concave:%t visited:%t}",
		v.index, sp, v.Point, sn, v.concave, v.visited)
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

func newSide(start, end *Vertex) *Side {
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

//   horizontal    vertical
//    x0   x1           y
//         |      x0    v
//   y >---+            |
//         |      x1 ---+----
type Chord struct {
	v *Vertex
	s *Side
}

func (c *Chord) X0X1YVert() (x0, x1, y float64, vertical bool) {
	vertical = !c.s.vertical
	x0, y = c.v.X, c.v.Y
	x1 = c.s.y
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	return
}

func (c Chord) String() string {
	return fmt.Sprintf("CHORD{%s v=%v s=%v %v}",
		rectString(&c), c.v.Point, c.s.start.Point, c.s.end.Point)
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
