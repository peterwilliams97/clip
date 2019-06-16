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

// Vertex is a vertex on a rectilinear polygon.
type Vertex struct {
	Point
	iPath   int     // Index of countour in polygon.
	index   int     // Index of vertex in contour.
	prev    *Vertex // Previous vertex in contour (or polygon?) !@#$
	next    *Vertex // Next vertex in contour (or polygon?) !@#$
	concave bool    // True if vertex is concave.
	visited bool    // Does this belong here?
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
	start, end   *Vertex // Vertices at the start and end of the segment.
	vertical     bool    // Is this a vertical segment?
	number       int
	pStart, pEnd Point
}

// func NewSeg(x0, x1 float64) *Side {
// 	return &Side{x0: x0, x1: x1}
// }

func newSide(start, end *Vertex, vertical bool) *Side {
	start.Validate()
	end.Validate()

	var x0, x1 float64
	if vertical { // Why vertical -> X  ? !@#$ Seems to be consistently inverted.
		x0 = start.Point.X
		x1 = end.Point.X
	} else {
		x0 = start.Point.Y
		x1 = end.Point.Y
	}
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if x0 == x1 {
		panic("not allowed")
	}

	return &Side{
		x0:       x0,
		x1:       x1,
		start:    start,
		end:      end,
		vertical: vertical,
		number:   -1,
		pStart:   start.Point,
		pEnd:     end.Point,
	}
}

func vertexIndex(vertices []*Vertex, vtx *Vertex) int {
	if len(vertices) == 0 {
		return -1
	}
	for i, v := range vertices {
		if v == vtx {
			return i
		}
	}
	return -1
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
