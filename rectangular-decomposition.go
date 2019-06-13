package clip

import (
	"math"
	"sort"

	"github.com/unidoc/unipdf/common"
)

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
	common.Log.Info("DecomposeRegion:====================================-")
	common.Log.Info("DecomposeRegion: paths=%d clockwise=%t", len(paths), clockwise)
	for i, path := range paths {
		common.Log.Info("\t%3d:%+v", i, path)
	}
	common.Log.Info("DecomposeRegion:====================================+")

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
			a.Validate()
			b.Validate()
		}
	}
	htree := CreateIntervalTree(hsegments, "hsegments")
	vtree := CreateIntervalTree(vsegments, "vsegments")

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

//    Concave vertices
//    ================
//         ^          ^      >--+      +---<
//     a)  |     b)   |    c)   |   d) |
//         +--<    >--+         v      v
//
//         v          v      <--+      +--->
//     e)  |     f)   |    g)   |   h) |
//         +-->    <--+         ^      ^
//
//     anti-clockwise  a),c),f),h)
//     ---------------------------
//          +-<-+              +--<---+    +------+
//         f|   |a             |      |    |      |
//      +---+   +---+          |      |a  f|      |
//      |           |          +---+  +----+  +---+
//      +---+   +---+             c|          |h
//         c|   |h                f|          |a
//          +->-+              +---+  +----+  +---+
//                             |      |h  c|      |
//                             |      |    |      |
//                             +-->---+    +------+
//
//     clockwise  b),d),e),g)
//     ----------------------
//          +->-+              +-->---+    +------+
//         b|   |e             |      |    |      |
//      +---+   +---+          |      |e  b|      |
//      |           |          +---+  +----+  +---+
//      +---+   +---+             g|          |d
//         g|   |d                b|          |e
//          +-<-+              +---+  +----+  +---+
//                             |      |d  g|      |
//                             |      |    |      |
//                             +--<---+    +------+

func splitConcave(vertices []*Vertex) {
	common.Log.Info("splitConcave: vertices=%d", len(vertices))
	for i, v := range vertices {
		common.Log.Info("\t%3d: % %p", i, v, v)
		v.Validate()
	}
	common.Log.Info("================^^^================")
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
	common.Log.Info("splitConcave: leftsegments=%d", len(leftsegments))
	for i, s := range leftsegments {
		common.Log.Info("\t%3d: %+v", i, *s)
	}
	common.Log.Info("splitConcave: rightsegments=%d", len(rightsegments))
	for i, s := range rightsegments {
		common.Log.Info("\t%3d: %+v", i, *s)
	}
	common.Log.Info("================~~~================")

	lefttree := CreateIntervalTree(leftsegments, "leftsegments")
	righttree := CreateIntervalTree(rightsegments, "rightsegments")
	common.Log.Debug("splitConcave: lefttree=%v", lefttree)
	common.Log.Debug("splitConcave: righttree=%v", righttree)

	for i, v := range vertices {
		if !v.concave {
			continue
		}
		common.Log.Info("@@i=%d v=%#v", i, v)

		// Compute orientation
		y := v.point.Y
		var toRight bool                 // a),b),e),f)
		if v.prev.point.X == v.point.X { // cases e)-h)
			toRight = v.prev.point.Y < y // e),f)
		} else { //                         cases a)-d)
			toRight = v.next.point.Y < y // a), b)
		}
		common.Log.Info("splitConcave: i=%d toRight=%t y=%g", i, toRight, y)
		common.Log.Info("prev=%v point=%v next=%v", v.prev.point, v.point, v.next.point)
		common.Log.Info("X:prev->point: %s", getDirection(v.prev.point.X, v.point.X))
		common.Log.Info("Y:prev->point: %s", getDirection(v.prev.point.Y, v.point.Y))
		common.Log.Info("Y:next->point: %s", getDirection(v.next.point.Y, v.point.Y))

		v.Validate()

		// Scan a horizontal ray
		var closestDistance float64
		var closestSegment *Segment
		if !toRight {
			closestDistance = -infinity
			righttree.QueryPoint(v.point.X, func(h *Segment) bool {
				x := h.start.point.Y
				common.Log.Info("cb: closestDistance=%g x=%g y=%g h=%+v", closestDistance, x, y, h)
				match := x > closestDistance
				if match {
					closestDistance = x
					closestSegment = h
				}
				common.Log.Info("cb: match=%t closest=%g %v", match, closestDistance,
					closestSegment)
				return false
			})
		} else {
			closestDistance = infinity
			lefttree.QueryPoint(v.point.X, func(h *Segment) bool {
				x := h.start.point.Y
				common.Log.Debug("cb: y=%g x=%g closestDistance=%g ", y, x, closestDistance)
				match := x < closestDistance
				if match {
					closestDistance = x
					closestSegment = h
				}
				common.Log.Debug("cb: x=%g h=%v match=%t\n\tclosest=%g %v", x, *h, match, closestDistance,
					closestSegment)
				return false
			})
		}

		common.Log.Info("closestDistance=%g closestSegment=%+v", closestDistance, closestSegment)
		common.Log.Info("closestSegment\n\tstart=%+v\n\t  end=%+v\n\t    v=%+v",
			*closestSegment.start, *closestSegment.end, *v)

		panic("Done")

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

		common.Log.Info("splitA=%+v", *splitA)
		common.Log.Info("splitB=%+v", *splitB)
		splitA.Validate()
		splitB.Validate()

		// Update segment tree
		var tree *IntervalTree
		if toRight {
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
	common.Log.Debug("getDiagonals: vertices=%d vertical=%t", len(vertices), vertical)

	var concave []*Vertex
	for _, v := range vertices {
		v.Validate()
		if v.concave {
			concave = append(concave, v)
		}
	}
	common.Log.Debug("concave=%d", len(concave))

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
	common.Log.Debug("diagonals=%d", len(diagonals))
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
	htree := CreateIntervalTree(hdiagonals, "hdiagonals")
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
	common.Log.Info("findRegions: %d vertices", len(vertices))
	for i := 0; i < n; i++ {
		vertices[i].visited = false
		v := vertices[i]
		common.Log.Info("%4d: %p %v %v %v", i, v, *v, v.prev.point, v.next.point)
	}
	for _, v := range vertices {
		v.Validate()
	}
	common.Log.Info("~~~~~~~~~~~~~~~~~~~~~~~~~")
	//   0  1  2  3
	// 0 +--+  +--+
	//   |  |  |  |
	// 1 |  +--+  |
	//   |        |
	// 2 +--------+
	// Walk over vertex list
	var rectangles []Rect
	var count int
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
			lo.Y = math.Min(p.Y, lo.Y)
			hi.Y = math.Max(p.Y, hi.Y)
			v.visited = true
			common.Log.Info("visit %d %p %v %v %v", count, v, *v, lo, hi)
			count++
		}
		r := Rect{Llx: lo.X, Lly: lo.Y, Urx: hi.X, Ury: hi.Y}
		rectangles = append(rectangles, r)
		common.Log.Info("%4d %d: %+v", i, len(rectangles)-1, r)
	}
	return rectangles
}
