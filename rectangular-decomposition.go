package clip

import (
	"math"
	"sort"

	"github.com/unidoc/unidoc/common"
)

/*
	Coordinate origin is top-left

*/

// Vertex is a vertex on a rectilinear polygon.
type Vertex struct {
	point   Point
	iPath   int     // Index of countour (path) in paths (polygon).
	index   int     // Index of point in contour.
	concave bool    // True if vertex is concave.
	next    *Vertex // Next vertex in contour (or polygon?) !@#$
	prev    *Vertex // Previous vertex in contour (or polygon?) !@#$
	visited bool
}

func (v *Vertex) validate() {
	if v.prev != nil && v.prev.point.Equals(v.point) {
		common.Log.Error("\n\tprev=%#v\n\t   v=%#v\n\tnext=%#v", *v.prev, *v, *v.next)
		panic("duplicate point: prev")
	}
	if v.next != nil && v.next.point.Equals(v.point) {
		common.Log.Error("\n\tprev=%#v\n\t   v=%#v\n\tnext=%#v", *v.prev, *v, *v.next)
		panic("duplicate point: next")
	}
}

// Segment is a vertical or horizontal segment.
type Segment struct { // A chord?
	x0, x1       float64 // Start and end of the interval in the vertical or horizontal direction.
	start, end   *Vertex // Vertices at the start and end of the segment.
	vertical     bool    // Is this a vertical segment?
	number       int
	iStart, iEnd int
}

// func NewSeg(x0, x1 float64) *Segment {
// 	return &Segment{x0: x0, x1: x1}
// }

func newSegment(start, end *Vertex, vertical bool) *Segment {
	return newSegmentVertices(start, end, vertical, nil)
}

func newSegmentVertices(start, end *Vertex, vertical bool, vertices []*Vertex) *Segment {
	var x0, x1 float64
	if vertical { // Why vertical -> X  ? !@#$ Seems to be consistently inverted.
		x0 = start.point.X
		x1 = end.point.X
	} else {
		x0 = start.point.Y
		x1 = end.point.Y
	}
	if x0 > x1 {
		x0, x1 = x1, x0
	}

	return &Segment{
		x0:       x0,
		x1:       x1,
		start:    start,
		end:      end,
		vertical: vertical,
		number:   -1,
		iStart:   vertexIndex(vertices, start),
		iEnd:     vertexIndex(vertices, end),
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

func integerizePoly(poly []Path) []Path {
	for i, path := range poly {
		poly[i] = path.integerize()
	}
	return poly
}

// DecomposeRegion breaks rectilinear polygon `paths` into non-overlapping rectangles.
// * `paths` is an array of loops vertices representing the boundary of the region.  Each loop must
//    be a simple rectilinear polygon (ie no self intersections), and the line segments of any two
//    loops must only meet at vertices.  The collection of loops must also be bounded.
// * `clockwise` is a boolean flag which if set flips the orientation of the loops.  Default is
//    `true`, ie all loops follow the right-hand rule (counter clockwise orientation)
// **Returns** A list of rectangles that decompose the region bounded by loops into the smallest
//  number of non-overlapping rectangles
func DecomposeRegion(paths []Path, clockwise bool) []Rect {
	paths = integerizePoly(paths)
	// clockwise = !clockwise
	common.Log.Debug("DecomposeRegion:====================================-")
	common.Log.Debug("DecomposeRegion: paths=%d clockwise=%t", len(paths), clockwise)
	for i, path := range paths {
		common.Log.Debug("\t%3d:%+v", i, path)
	}
	common.Log.Debug("DecomposeRegion:====================================+")

	// First step: unpack all vertices into internal format.
	var vertices []*Vertex

	npaths := make([][]*Vertex, len(paths))
	for i, path := range paths {
		n := len(path)
		prev := path[n-3]
		cur := path[n-2]
		next := path[n-1]
		common.Log.Debug("DecomposeRegion: i=%d\n\t path=%+v n=%d\n\t prev=%#v\n\t  cur=%#v\n\t next=%#v",
			i, path, n, prev, cur, next)
		for j := 0; j < n; j++ {
			prev = cur
			cur = next
			next = path[j]
			common.Log.Debug("---------------------------------------------")
			common.Log.Debug("j=%d\n\t prev=%+v\n\t  cur=%+v\n\t next=%+v", j, prev, cur, next)
			concave := false

			// prev.X == cur.X && next.X != cur.X
			// prev.Y == cur.Y && next.Y != cur.Y
			if prev.X == cur.X {
				if next.X == cur.X {
					continue
				}
				// a) v    b)   v   c) +-->   d) <--+
				//    |         |      |            |
				//    +-->   <--+      ^            ^

				dir0 := prev.Y < cur.Y // c) d)
				dir1 := cur.X < next.X // a) c)
				concave = dir0 == dir1 // b) c) !@#$ Not concave for anti-clockwise
				if !clockwise {
					concave = !concave
				}
				common.Log.Debug("  @1 dir0=%t dir1=%t concave=%t", dir0, dir1, concave)
			} else {
				if next.Y == cur.Y {
					continue
				}
				dir0 := prev.X < cur.X
				dir1 := cur.Y < next.Y
				concave = dir0 != dir1
				if !clockwise {
					concave = !concave
				}
				common.Log.Debug("  @2 dir0=%t dir1=%t concave=%t", dir0, dir1, concave)
			}

			vtx := &Vertex{
				point:   cur,
				iPath:   i,
				index:   (j + n - 1) % n,
				concave: concave,
			}
			common.Log.Debug("vtx=%+v", vtx)
			npaths[i] = append(npaths[i], vtx)
			vertices = append(vertices, vtx)
		}
	}

	// Next build interval trees for segments, link vertices into a list.
	var hsegments []*Segment
	var vsegments []*Segment

	for _, p := range npaths {
		for j := 0; j < len(p); j++ {
			k := (j + 1) % len(p)
			a := p[j]
			b := p[k]
			if a.point.X == b.point.X {
				// hsegments are vertical !@#$
				hsegments = append(hsegments, newSegment(a, b, false))
			} else {
				// vsegments are horizontal !@#$
				vsegments = append(vsegments, newSegment(a, b, true))
			}
			// if clockwise {
			// 	b.next = a
			// } else {
			// 	a.prev, a.next = a, b
			// }
			if clockwise {
				a.prev, b.next = b, a
			} else {
				a.next, b.prev = b, a
			}
			common.Log.Debug("clockwise=%t len(p)=%d\n\tp[%d]=%v\n\tp[%d]=%v",
				clockwise, len(p), j, a, k, b)
			a.validate()
			b.validate()
		}
	}
	htree := CreateIntervalTree(hsegments)
	vtree := CreateIntervalTree(vsegments)

	// Find horizontal and vertical diagonals.
	hdiagonals := getDiagonals(vertices, npaths, false, vtree)
	vdiagonals := getDiagonals(vertices, npaths, true, htree)

	// Find all splitting edges
	splitters := findSplitters(hdiagonals, vdiagonals)

	// Cut all the splitting diagonals
	for _, splitter := range splitters {
		splitSegment(splitter)
	}

	// Split all concave vertices
	splitConcave(vertices)

	// Return regions
	return findRegions(vertices)
}

func splitConcave(vertices []*Vertex) {
	common.Log.Debug("splitConcave: vertices=%d", len(vertices))
	for i, v := range vertices {
		common.Log.Debug("\t%3d: %+v", i, v)
	}
	common.Log.Debug("=============================")
	// First step: build segment tree from vertical segments.
	var leftsegments []*Segment
	var rightsegments []*Segment

	for i, v := range vertices {
		common.Log.Debug("\t%3d: %+v", i, v)
		if v.next.point.Y == v.point.Y {
			if v.next.point.X < v.point.X {
				// <--
				leftsegments = append(leftsegments, newSegmentVertices(v, v.next, true, vertices))
			} else {
				// -->
				rightsegments = append(rightsegments, newSegmentVertices(v, v.next, true, vertices))
			}
		}
	}
	common.Log.Debug("splitConcave: leftsegments=%d", len(leftsegments))
	for i, s := range leftsegments {
		common.Log.Debug("\t%3d: %+v", i, *s)
	}
	common.Log.Debug("splitConcave: rightsegments=%d", len(rightsegments))
	for i, s := range rightsegments {
		common.Log.Debug("\t%3d: %+v", i, *s)
	}

	lefttree := CreateIntervalTree(leftsegments)
	righttree := CreateIntervalTree(rightsegments)
	common.Log.Debug("splitConcave: lefttree=%v", lefttree)
	common.Log.Debug("splitConcave: righttree=%v", righttree)

	for i, v := range vertices {
		common.Log.Debug("i=%d v=%#v", i, v)
		if !v.concave {
			continue
		}

		// Compute orientation
		y := v.point.Y
		var direct bool
		if v.prev.point.X == v.point.X {
			// |                         ^
			// |                         |
			// v  direct = true          | direct = false
			direct = v.prev.point.Y < y
		} else {
			//    ^                   ---+
			//    |                      |
			// ---+  direct = true       v  direct = false
			direct = v.next.point.Y < y
		}

		common.Log.Debug("queryPoint: direction=%t y=%g", direct, y)
		common.Log.Debug("prev=%v point=%v next=%v", v.prev.point, v.point, v.next.point)
		v.validate()

		// Scan a horizontal ray
		var closestDistance float64
		var closestSegment *Segment
		if direct {
			closestDistance = -infinity
			righttree.QueryPoint(v.point.X, func(h *Segment) bool {
				x := h.start.point.Y
				match := closestDistance < x && x < y
				if match {
					closestDistance = x
					closestSegment = h
				}
				common.Log.Debug("x=%g h=%v match=%t closest=%g %v", x, *h, match, closestDistance,
					closestSegment)
				return false
			})
		} else {
			closestDistance = infinity
			lefttree.QueryPoint(v.point.X, func(h *Segment) bool {
				x := h.start.point.Y
				match := y < x && x < closestDistance
				if match {
					closestDistance = x
					closestSegment = h
				}
				common.Log.Debug("x=%g h=%v match=%t closest=%g %v", x, *h, match, closestDistance,
					closestSegment)
				return false
			})
		}

		common.Log.Debug("closestSegment=%#v closestDistance=%g\n", closestSegment, closestDistance)

		// Create two splitting vertices
		point := Point{v.point.X, closestDistance}
		splitA := &Vertex{point: point}
		splitB := &Vertex{point: point}

		// Clear concavity flag
		v.concave = false

		// Split vertices
		splitA.prev = closestSegment.start
		closestSegment.start.next = splitA
		splitB.next = closestSegment.end
		closestSegment.end.prev = splitB

		// Update segment tree
		var tree *IntervalTree
		if direct {
			tree = righttree
		} else {
			tree = lefttree
		}
		tree.Delete(closestSegment)
		tree.Insert(newSegment(closestSegment.start, splitA, true))
		tree.Insert(newSegment(splitB, closestSegment.end, true))

		// Append vertices
		vertices = append(vertices, splitA, splitB)

		// Cut v, 2 different cases
		if v.prev.point.X == v.point.X {
			// Case 1
			//             ^
			//             |
			// --->*+++++++X
			//     |       |
			//     V       |
			splitA.next = v
			splitB.prev = v.prev
		} else {
			// Case 2
			//     |       ^
			//     V       |
			// <---*+++++++X
			//             |
			//             |
			splitA.next = v.next
			splitB.prev = v
		}

		// Fix up links
		splitA.next.prev = splitA
		splitB.prev.next = splitB
	}
}

// type interval.Tree struct{}

// type Diagonal struct{}
// type Splitter struct{}

func getDiagonals(vertices []*Vertex, npaths [][]*Vertex, vertical bool, tree *IntervalTree) []*Segment {
	var concave []*Vertex
	for _, v := range vertices {
		if v.concave {
			concave = append(concave, v)
		}
	}
	if vertical {
		sort.Slice(concave, func(i, j int) bool {
			a, b := concave[i], concave[j]
			d := a.point.Y - b.point.Y
			if d != 0 {
				return d > 0
			}
			return a.point.X > b.point.X
		})
	} else {
		sort.Slice(concave, func(i, j int) bool {
			a, b := concave[i], concave[j]
			d := a.point.X - b.point.X
			if d != 0 {
				return d > 0
			}
			return a.point.Y > b.point.Y
		})
	}

	var diagonals []*Segment
	for i := 1; i < len(concave); i++ {
		a := concave[i-1]
		b := concave[i]
		var sameDirection bool
		if vertical {
			sameDirection = a.point.Y == b.point.Y
		} else {
			sameDirection = a.point.X == b.point.X
		}

		if sameDirection {
			if a.iPath == b.iPath {
				n := len(npaths[a.iPath])
				d := (a.index - b.index + n) % n
				if d == 1 || d == n-1 {
					continue
				}
			}
			if !testSegment(a, b, tree, vertical) {
				// Check orientation of diagonal
				diagonals = append(diagonals, newSegment(a, b, vertical))
			}
		}
	}
	return diagonals
}

func findSplitters(hdiagonals, vdiagonals []*Segment) []*Segment {
	common.Log.Debug("findSplitters: hdiagonals=%d vdiagonals=%d", len(hdiagonals), len(vdiagonals))

	// First find crossings
	crossings := findCrossings(hdiagonals, vdiagonals)
	common.Log.Debug("findSplitters: crossings=%d", len(crossings))

	// Then tag and convert edge format
	for i := 0; i < len(hdiagonals); i++ {
		hdiagonals[i].number = i
	}
	for i := 0; i < len(vdiagonals); i++ {
		vdiagonals[i].number = i
	}

	//   var edges = crossings.map(function(c) {
	//     return [ c[0].number, c[1].number ]
	//   })
	edges := make([][2]int, len(crossings))
	for i, c := range crossings {
		edges[i] = [2]int{c.h.number, c.v.number}
	}

	// Find independent set
	selectedL, selectedR := bipartiteIndependentSet(len(hdiagonals), len(vdiagonals), edges)

	// Convert into result format
	result := make([]*Segment, len(selectedL)+len(selectedR))
	for i, v := range selectedL {
		result[i] = hdiagonals[v]
	}
	for i, v := range selectedR {
		result[i+len(selectedR)] = hdiagonals[v]
	}

	return result
}

type Crossing struct {
	h, v *Segment
}

// Find all crossings between diagonals.
func findCrossings(hdiagonals, vdiagonals []*Segment) []Crossing {
	htree := CreateIntervalTree(hdiagonals)
	var crossings []Crossing
	for _, v := range vdiagonals {
		// x := v.start.point.X
		htree.QueryPoint(v.start.point.Y, func(h *Segment) bool {
			x := h.start.point.X
			if v.x0 <= x && x <= v.x1 {
				crossings = append(crossings, Crossing{h: h, v: v})
			}
			return false
		})
	}
	return crossings
}

func splitSegment(segment *Segment) {
	//Store references
	a := segment.start
	b := segment.end
	pa := a.prev
	na := a.next
	pb := b.prev
	nb := b.next

	// Fix concavity
	a.concave = false
	b.concave = false

	// Compute orientation
	ao := pa.point.Cpt(segment.vertical) == a.point.Cpt(segment.vertical)
	bo := pb.point.Cpt(segment.vertical) == b.point.Cpt(segment.vertical)

	if ao && bo {
		//Case 1:
		//            ^
		//            |
		//  --->A+++++B<---
		//      |
		//      V
		a.prev = pb
		pb.next = a
		b.prev = pa
		pa.next = b
	} else if ao && !bo {
		//Case 2:
		//      ^     |
		//      |     V
		//  --->A+++++B--->
		//
		//
		a.prev = b
		b.next = a
		pa.next = nb
		nb.prev = pa
	} else if !ao && bo {
		//Case 3:
		//
		//
		//  <---A+++++B<---
		//      ^     |
		//      |     V
		a.next = b
		b.prev = a
		na.prev = pb
		pb.next = na

	} else if !ao && !bo {
		//Case 3:
		//            |
		//            V
		//  <---A+++++B--->
		//      ^
		//      |
		a.next = nb
		nb.prev = a
		b.next = na
		na.prev = b
	}
}

func findRegions(vertices []*Vertex) []Rect {
	n := len(vertices)
	for i := 0; i < n; i++ {
		vertices[i].visited = false
	}
	// Walk over vertex list
	var rectangles []Rect
	for i := 0; i < n; i++ {
		v := vertices[i]
		if v.visited {
			continue
		}
		// Walk along loop
		lo := Point{infinity, infinity}
		hi := Point{-infinity, -infinity}
		for ; !v.visited; v = v.next {
			p := v.point
			lo.X = math.Min(p.X, lo.X)
			hi.X = math.Max(p.X, hi.X)
			lo.Y = math.Min(p.Y, lo.X)
			hi.Y = math.Max(p.Y, hi.X)
			v.visited = true
		}
		r := Rect{Llx: lo.X, Lly: lo.Y, Urx: hi.X, Ury: hi.Y}
		rectangles = append(rectangles, r)
	}
	return rectangles
}
