package clip

import (
	"math"
	"sort"

	"github.com/biogo/store/interval"
	"github.com/unidoc/unipdf/common"
)

// DecomposeRegion breaks rectilinear polygon `polygon` into non-overlapping rectangles.
// `polygon`: an array of contours representing the boundary of the region.  Each contour must
//    be a simple rectilinear polygon (i.e. no self-intersections), and the line segments of any two
//    contours must only meet at vertices.
// `clockwise`: a boolean flag which if set flips the orientation of the loops.  Default is
//    `true`, ie all loops follow the right-hand rule (counter clockwise orientation) !@#$
//  Returns: A list of rectangles that decompose the region bounded by loops into the smallest
//  number of non-overlapping rectangles
func DecomposeRegion(polygon []Path, clockwise bool) []Rect {
	polygon = integerizePoly(polygon)
	// clockwise = !clockwise
	common.Log.Info("DecomposeRegion:====================================-")
	common.Log.Info("DecomposeRegion: polygon=%d clockwise=%t", len(polygon), clockwise)
	for i, path := range polygon {
		common.Log.Info("\t%3d:%+v", i, path)
	}
	common.Log.Info("DecomposeRegion:====================================+")

	// First step: unpack all vertices into internal format.
	var vertices []*Vertex

	contours := make([][]*Vertex, len(polygon))
	for i, path := range polygon {
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
				Point:   cur,
				iPath:   i,
				index:   (j + n - 1) % n,
				concave: concave,
			}
			common.Log.Debug("vtx=%+v", vtx)
			contours[i] = append(contours[i], vtx)
			vertices = append(vertices, vtx)
		}
	}

	// Next build interval trees for segments, link vertices into a list.
	// vSides: vertical edges
	// hSides: horizontal edges
	var vSides, hSides []*Side

	for _, contour := range contours {
		for j, p0 := range contour {
			k := (j + 1) % len(contour)
			p1 := contour[k]
			if p0.X == p1.X {
				vSides = append(vSides, newSide(p0, p1, false))
				if p0.Y == p1.Y {
					panic("duplicate point")
				}
			} else {
				// hSides are horizontal !@#$
				hSides = append(hSides, newSide(p0, p1, true))
				if p0.Y != p1.Y {
					panic("diagonal")
				}
			}
			// if clockwise {
			// 	p1.next = p0
			// } else {
			// 	p0.prev, p0.next = p0, p1
			// }
			if clockwise {
				p0.prev, p1.next = p1, p0
			} else {
				p0.next, p1.prev = p1, p0
			}
			common.Log.Debug("clockwise=%t len(p)=%d\n\tp[%d]=%v\n\tp[%d]=%v",
				clockwise, len(contour), j, p0, k, p1)
			p0.Validate()
			p1.Validate()
		}
	}
	htree := CreateIntervalTree(vSides, "vSides")
	vtree := CreateIntervalTree(hSides, "hSides")

	// Find horizontal and vertical diagonals.
	// Are these supposed to be cogrid chords? !@#$%
	hdiagonals := getDiagonals(vertices, contours, false, vtree)
	vdiagonals := getDiagonals(vertices, contours, true, htree)

	// Find all splitting edges
	splitters := findSplitters(hdiagonals, vdiagonals)

	// Cut all the splitting diagonals
	for _, splitter := range splitters {
		splitSide(splitter)
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
		common.Log.Info("\t%3d: %p %s", i, v, v)
		v.Validate()
	}
	common.Log.Info("================^^^================")

	// First step: build segment tree from vertical segments.
	var leftsegments, rightsegments []*Side
	for i, v := range vertices {
		common.Log.Debug("\t%3d: %+v", i, v)
		if v.next.Y == v.Y {
			if v.next.X < v.X { // <--
				leftsegments = append(leftsegments, newSide(v, v.next, true))
			} else { //                        -->
				rightsegments = append(rightsegments, newSide(v, v.next, true))
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
		//         ^          ^      >--+      +---<
		//     a)  |     b)   |    c)   |   d) |
		//         +--<    >--+         v      v
		//
		//         v          v      <--+      +--->
		//     e)  |     f)   |    g)   |   h) |
		//         +-->    <--+         ^      ^

		y0 := v.Y
		var toLeft bool      // a),b),e),f)
		if v.prev.X == v.X { // cases e)-h)
			toLeft = v.prev.Y < y0 // e),f)
		} else { //                         cases a)-d)
			toLeft = v.next.Y < y0 // a), b)
		}
		common.Log.Info("splitConcave: i=%d toLeft=%t y0=%g", i, toLeft, y0)
		common.Log.Info("prev=%v point=%v next=%v", v.prev.Point, v.Point, v.next.Point)
		common.Log.Info("X:prev->Point: %s | Y:prev->Point: %s | Y:next->Point: %s",
			getDirection(v.prev.X, v.X),
			getDirection(v.prev.Y, y0),
			getDirection(v.next.Y, y0))

		v.Validate()

		// Scan a horizontal ray
		var closestDistance float64
		var closestSide *Side
		common.Log.Info("----scan: v=%+v  toLeft=%t", v.Point, toLeft)
		if !toLeft {
			closestDistance = -infinity
			righttree.QueryPoint(v.X, func(h *Side) bool {
				y := h.start.Y
				match := y0 > y && y > closestDistance
				common.Log.Info("cb: righttree y=%g  y0=%g closestDistance=%g\n\th=%+v",
					y, y0, closestDistance, *h)
				if match {
					closestDistance = y
					closestSide = h
				}
				common.Log.Info("cb: match=%t\n\tclosest=%g %+v", match, closestDistance, closestSide)
				return false
			})
		} else {
			closestDistance = infinity
			lefttree.QueryPoint(v.X, func(h *Side) bool {
				y := h.start.Y
				match := y0 < y && y < closestDistance
				common.Log.Info("cb: lefttree y=%g  y0=%g closestDistance=%g\n\th=%+v",
					y, y0, closestDistance, *h)
				if match {
					closestDistance = y
					closestSide = h
				}
				common.Log.Info("cb: match=%t\n\tclosest=%g %+v", match, closestDistance, closestSide)
				return false
			})
		}

		common.Log.Info("closestDistance=%g closestSide=%+v", closestDistance, closestSide)
		common.Log.Info("closestSide\n\tstart=%+v\n\t  end=%+v\n\t    v=%+v",
			*closestSide.start, *closestSide.end, *v)

		// panic("Done")

		// Create two splitting vertices.
		point := Point{v.X, closestDistance}
		splitA := &Vertex{Point: point}
		splitB := &Vertex{Point: point}

		// Clear concavity flag
		v.concave = false

		// Split vertices
		splitA.prev = closestSide.start
		closestSide.start.next = splitA
		splitB.next = closestSide.end
		closestSide.end.prev = splitB

		common.Log.Info("splitA=%+v", *splitA)
		common.Log.Info("splitB=%+v", *splitB)
		splitA.Validate()
		splitB.Validate()

		// Update segment tree
		var tree *IntervalTree
		if toLeft {
			tree = righttree
		} else {
			tree = lefttree
		}
		tree.Delete(closestSide)
		tree.Insert(newSide(closestSide.start, splitA, true))
		tree.Insert(newSide(splitB, closestSide.end, true))

		// Append vertices
		vertices = append(vertices, splitA, splitB)

		// Cut v, 2 different cases
		if v.prev.X == v.X {
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

// getDiagonals returns the chords off concave vertices in `vertices`.
func getDiagonals(vertices []*Vertex, contours [][]*Vertex, vertical bool, tree *IntervalTree) []*Side {
	common.Log.Info("getDiagonals: vertices=%d vertical=%t", len(vertices), vertical)

	var concave []*Vertex
	for _, v := range vertices {
		v.Validate()
		if v.concave {
			concave = append(concave, v)
		}
	}

	if vertical {
		sort.Slice(concave, func(i, j int) bool {
			a, b := concave[i], concave[j]
			d := a.Y - b.Y
			if d != 0 {
				return d > 0
			}
			return a.X > b.X
		})
	} else {
		sort.Slice(concave, func(i, j int) bool {
			a, b := concave[i], concave[j]
			d := a.X - b.X
			if d != 0 {
				return d > 0
			}
			return a.Y > b.Y
		})
	}
	common.Log.Info("concave=%d", len(concave))
	for i, s := range concave {
		common.Log.Info("%4d: %v", i, s)
	}

	var diagonals []*Side
	for i := 1; i < len(concave); i++ {
		a := concave[i-1]
		b := concave[i]
		var sameDirection bool
		if vertical {
			sameDirection = a.Y == b.Y
		} else {
			sameDirection = a.X == b.X
		}
		common.Log.Info("i=%d: sameDirection=%t\n\ta=%v\n\tb=%v", i, sameDirection, a, b)
		if !sameDirection {
			continue
		}

		if a.iPath == b.iPath {
			n := len(contours[a.iPath])
			d := (a.index - b.index + n) % n
			common.Log.Info("i=%d: n=%d d=%d", i, n, d)
			if d == 1 || d == n-1 {
				// Adjacent points
				continue
			}
		}
		if !testSide(a, b, tree, vertical) {
			// Check orientation of diagonal
			diagonals = append(diagonals, newSide(a, b, vertical))
		}

	}
	common.Log.Info("diagonals=%d", len(diagonals))
	for i, s := range diagonals {
		common.Log.Info("%4d: %v", i, s)
	}
	return diagonals
}

// testSide returns true if segment [v0,v1] intersects an existing segment.
func testSide(v0, v1 *Vertex, tree *IntervalTree, vertical bool) bool {
	i := newInterval(v0, v1, vertical)
	t := (*interval.Tree)(tree)
	matches := t.Get(i)
	common.Log.Info("testSide: i=%v vertical=%t tree=%v matches=%d\n\tv0=%v\n\tv1=%v",
		i, vertical, tree, matches, v0, v1)
	return len(matches) > 0
}

func findSplitters(hdiagonals, vdiagonals []*Side) []*Side {
	common.Log.Info("findSplitters: hdiagonals=%d vdiagonals=%d", len(hdiagonals), len(vdiagonals))

	// First find crossings
	crossings := findCrossings(hdiagonals, vdiagonals)
	common.Log.Info("findSplitters: crossings=%d", len(crossings))

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
	common.Log.Info("selectedL=%d", len(selectedL))
	common.Log.Info("selectedR=%d", len(selectedR))

	// Convert into result format
	result := make([]*Side, len(selectedL)+len(selectedR))
	for i, v := range selectedL {
		result[i] = hdiagonals[v]
	}
	for i, v := range selectedR {
		result[i+len(selectedL)] = vdiagonals[v]
	}
	common.Log.Info("result=%d", len(result))
	// panic("done bipartite")

	return result
}

type Crossing struct {
	h, v *Side
}

// findCrossings returns the all intersections of horizontal and vertical chords.
func findCrossings(hdiagonals, vdiagonals []*Side) []Crossing {
	htree := CreateIntervalTree(hdiagonals, "hdiagonals")
	var crossings []Crossing
	for _, v := range vdiagonals {
		// x := v.start.X
		// !@#$ hdiagonals has to be verticals for this query to work!.
		htree.QueryPoint(v.start.Y, func(h *Side) bool {
			x := h.start.X
			if v.x0 <= x && x <= v.x1 {
				crossings = append(crossings, Crossing{h: h, v: v})
			}
			return false
		})
	}
	return crossings
}

func splitSide(segment *Side) {
	common.Log.Info("splitSide: %v", segment)
	panic("splitSide")
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
	ao := pa.Point.Cpt(segment.vertical) == a.Point.Cpt(segment.vertical)
	bo := pb.Point.Cpt(segment.vertical) == b.Point.Cpt(segment.vertical)

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
		common.Log.Info("%4d: %p %v %v %v", i, v, *v, v.prev.Point, v.next.Point)
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
			p := v.Point
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
