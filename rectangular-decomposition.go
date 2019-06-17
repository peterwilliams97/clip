package clip

import (
	"math"
	"sort"

	"github.com/biogo/store/interval"
	"github.com/unidoc/unipdf/common"
)

// DecomposeRegion breaks rectilinear polygon `polygon` into non-overlapping rectangles.
// `polygon`: an array of contours representing the boundary of the region.  Each contour must
//    be a simple rectilinear polygon (i.e. no self-intersections), and the line sides of any two
//    contours must only meet at vertices.
// `clockwise`: a boolean flag which if set flips the orientation of the loops.  Default is
//    `true`, ie all loops follow the right-hand rule (counter clockwise orientation) !@#$
//  Returns: A list of rectangles that decompose the region bounded by loops into the smallest
//  number of non-overlapping rectangles
func DecomposeRegion(polygon []Path, clockwise bool) []Rect {
	polygon = integerizePoly(polygon)
	common.Log.Info("DecomposeRegion:====================================-")
	common.Log.Info("DecomposeRegion: polygon=%d clockwise=%t", len(polygon), clockwise)
	for i, contour := range polygon {
		common.Log.Info("\t%3d:%+v", i, contour)
	}
	common.Log.Info("DecomposeRegion:====================================+")

	vertices, contours := asVertices(polygon, clockwise)

	// Next build interval trees for sides.
	// vSides: vertical edges. hSides: horizontal edges
	var vSides, hSides []*Side
	for _, contour := range contours {
		for j, v0 := range contour {
			k := (j + 1) % len(contour)
			v1 := contour[k]
			common.Log.Info("j=%d k=%d\n\tv0=%v\n\tv1=%v", j, k, v0, v1)
			if v0.X == v1.X {
				vSides = append(vSides, newSide(v0, v1))
			} else {
				hSides = append(hSides, newSide(v0, v1))
			}
			if clockwise {
				v0.prev, v1.next = v1, v0
			} else {
				v0.next, v1.prev = v1, v0
			}
			common.Log.Debug("clockwise=%t len(p)=%d\n\tp[%d]=%v\n\tp[%d]=%v",
				clockwise, len(contour), j, v0, k, v1)
			v0.Validate()
			v1.Validate()
		}
	}
	vTree := CreateIntervalTreeSides(vSides, "vSides")
	hTree := CreateIntervalTreeSides(hSides, "hSides")

	// Find horizontal and vertical chords.
	// Are these supposed to be cogrid chords? !@#$%
	hChords := getChords(vertices, contours, false, vTree)
	vChords := getChords(vertices, contours, true, hTree)

	// Find all splitting edges.
	splitters := findMinimalChords(hChords, vChords)

	// Cut all the splitting chords
	for _, splitter := range splitters {
		splitSide(splitter)
	}

	// Split all concave vertices
	splitConcave(vertices)

	// Return regions
	return findRegions(vertices)
}

// asVertices returns `polygon` in our internal format, a slice of vertices.
// - vertices is a vertex for every point in `polygon`.
// - contours is a vertex for every point in every contour in `polygon`.
func asVertices(polygon []Path, clockwise bool) (vertices []*Vertex, contours [][]*Vertex) {
	contours = make([][]*Vertex, len(polygon))

	for i, path := range polygon {
		n := len(path)
		for j, cur := range path {
			prev := path[(j-1+n)%n]
			next := path[(j+1+n)%n]
			common.Log.Debug("---------------------------------------------")
			common.Log.Debug("j=%d\n\t prev=%+v\n\t  cur=%+v\n\t next=%+v", j, prev, cur, next)
			concave := false

			if prev.X == cur.X {
				if next.X == cur.X {
					panic("xx1")
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
					panic("xx1")
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
				Point:    cur,
				iContour: i,
				index:    j,
				concave:  concave,
			}
			common.Log.Debug("vtx=%+v", vtx)
			contours[i] = append(contours[i], vtx)
			vertices = append(vertices, vtx)
		}
	}
	return vertices, contours
}

func splitConcave(vertices []*Vertex) {
	common.Log.Info("splitConcave: vertices=%d", len(vertices))
	for i, v := range vertices {
		common.Log.Info("\t%3d: %p %s", i, v, v)
		v.Validate()
	}
	common.Log.Info("================^^^================")

	// First step: build trees from vertical sides.
	var upVSides, downVSides []*Side
	for i, v := range vertices {
		common.Log.Debug("\t%3d: %+v", i, v)
		if v.next.X == v.X {
			if v.next.Y < v.Y { // ^
				upVSides = append(upVSides, newSide(v, v.next))
			} else { //            v
				downVSides = append(downVSides, newSide(v, v.next))
			}
		}
	}
	common.Log.Info("splitConcave: upVSides=%d", len(upVSides))
	for i, s := range upVSides {
		common.Log.Info("\t%3d: %+v", i, *s)
	}
	common.Log.Info("splitConcave: downVSides=%d", len(downVSides))
	for i, s := range downVSides {
		common.Log.Info("\t%3d: %+v", i, *s)
	}
	common.Log.Info("================~~~================")

	upVTree := CreateIntervalTreeSides(upVSides, "upVSides")
	downVTree := CreateIntervalTreeSides(downVSides, "downVSides")
	common.Log.Debug("splitConcave: upVTree=%v", upVTree)
	common.Log.Debug("splitConcave: downVTree=%v", downVTree)

	for i, v := range vertices {
		if !v.concave {
			continue
		}
		common.Log.Info("@@i=%d v=%#v", i, v)
		v.Validate()

		// Compute orientation of concavity.
		// "Concave up" is   v shaped. a),b),e),f).
		// "Concave down" is ^ shaped. c),d),g),h).
		// http://mathsfirst.massey.ac.nz/Calculus/Sign2ndDer/Sign2DerPOI.htm
		//
		//         ^          ^      >--+      +---<
		//     a)  |     b)   |    c)   |   d) |
		//         +--<    >--+         v      v
		//
		//         v          v      <--+      +--->
		//     e)  |     f)   |    g)   |   h) |
		//         +-->    <--+         ^      ^

		y0 := v.Y
		var concaveUp bool   //          a),b),e),f)
		if v.prev.X == v.X { //          e)-h)
			concaveUp = v.prev.Y < y0 // e),f)
		} else { //                      a)-d)
			concaveUp = v.next.Y < y0 // a), b)
		}
		common.Log.Info("splitConcave: i=%d concaveUp=%t y0=%g", i, concaveUp, y0)
		common.Log.Info("prev=%v point=%v next=%v", v.prev.Point, v.Point, v.next.Point)
		common.Log.Info("X:prev->Point: %s | Y:prev->Point: %s | Y:next->Point: %s",
			getDirection(v.prev.X, v.X),
			getDirection(v.prev.Y, y0),
			getDirection(v.next.Y, y0))

		// Scan a horizontal ray
		var closestDistance float64
		var closestSide *Side
		common.Log.Info("----scan: v=%+v  concaveUp=%t", v.Point, concaveUp)
		if concaveUp {
			closestDistance = infinity
			upVTree.QueryPoint(v.X, func(r Rectilinear) bool {
				h := r.(*Side)
				y := h.start.Y
				match := y0 < y && y < closestDistance
				common.Log.Info("cb: upVTree y=%g  y0=%g closestDistance=%g\n\th=%+v",
					y, y0, closestDistance, *h)
				if match {
					closestDistance = y
					closestSide = h
				}
				common.Log.Info("cb: match=%t\n\tclosest=%g %+v", match, closestDistance, closestSide)
				return false
			})
		} else {
			closestDistance = -infinity
			downVTree.QueryPoint(v.X, func(r Rectilinear) bool {
				h := r.(*Side)
				y := h.start.Y
				match := y0 > y && y > closestDistance
				common.Log.Info("cb: downVTree y=%g  y0=%g closestDistance=%g\n\th=%+v",
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
		if concaveUp {
			tree = downVTree
		} else {
			tree = upVTree
		}
		tree.Delete(closestSide)
		tree.Insert(newSide(closestSide.start, splitA))
		tree.Insert(newSide(splitB, closestSide.end))

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

// getChords returns the chords off concave vertices in `vertices`.
func getChords(vertices []*Vertex, contours [][]*Vertex, vertical bool, tree *IntervalTree) []*Chord {
	common.Log.Info("getChords: vertices=%d vertical=%t", len(vertices), vertical)

	var concave []*Vertex
	for _, v := range vertices {
		v.Validate()
		if v.concave {
			concave = append(concave, v)
		}
	}

	sort.Slice(concave, func(i, j int) bool {
		vi, vj := concave[i], concave[j]
		if vi.Cpt(vertical) != vj.Cpt(vertical) {
			return vi.Cpt(vertical) < vj.Cpt(vertical)
		}
		return vi.Cpt(!vertical) < vj.Cpt(!vertical)
	})

	common.Log.Info("concave=%d", len(concave))
	for i, v := range concave {
		common.Log.Info("%4d: %v", i, v)
	}

	// Assume counter clockwise
	//
	//         ^          ^      >--+      +---<
	//     a)  |     b)   |    c)   |   d) |
	//         +--<    >--+         v      v
	//
	//         v          v      <--+      +--->
	//     e)  |     f)   |    g)   |   h) |
	//         +-->    <--+         ^      ^

	//        +-<-+
	//      f |   | a
	//    +---+   +---+
	//    |           |
	//    +---+   +---+
	//      c |   | h
	//        +->-+

	var chords []*Chord
	for i, a := range concave[:len(concave)-1] {
		b := concave[i+1]
		// x0, x1 := minMax(a.Cpt(vertical), b.Cpt(vertical))
		// search lower (higher) from x0 (x1)
		// if vertical {
		// 	sameDirection = a.Y == b.Y
		// } else {
		// 	sameDirection = a.X == b.X
		// }
		// common.Log.Info("i=%d: sameDirection=%t\n\ta=%v\n\tb=%v", i, sameDirection, a, b)
		// if !sameDirection {
		// 	continue
		// }

		if a.iContour == b.iContour {
			n := len(contours[a.iContour])
			d := (a.index - b.index + n) % n
			common.Log.Info("i=%d: n=%d d=%d", i, n, d)
			if d == 1 || d == n-1 {
				// Adjacent points
				continue
			}
		}
		if !testSide(a, b, tree, vertical) {
			// Check orientation of diagonal
			// !@#$ Find the chords!
			// chords = append(chords, newSide(a, b, vertical))
		}

	}
	common.Log.Info("chords=%d", len(chords))
	for i, s := range chords {
		common.Log.Info("%4d: %v", i, s)
	}
	return chords
}

func minMax(x0, x1 float64) (float64, float64) {
	if x0 < x1 {
		return x0, x1
	}
	return x1, x0
}

// findChord returns the chord from the vertex
func findChord(vertex *Vertex, tree IntervalTree, vertical, increasing bool) *Chord {
	xx, yy := vertex.Point.Cpt(!vertical), vertex.Point.Cpt(vertical)
	var distance float64
	if increasing {
		distance = infinity
	} else {
		distance = -infinity
	}
	var closest *Side
	tree.QueryPoint(yy, func(r Rectilinear) bool {
		s := r.(*Side)
		x := s.start.Cpt(!vertical)
		if increasing {
			if x > xx && x-xx < distance {
				closest = s
				distance = x - xx
			}
		} else {
			if x > xx && x-xx < distance {
				closest = s
				distance = x - xx
			}
		}
		return false
	})

	if closest == nil {
		return nil
	}
	return &Chord{v: vertex, s: closest}
}

// testSide returns true if segment [v0,v1] intersects an existing segment.
func testSide(v0, v1 *Vertex, tree *IntervalTree, vertical bool) bool {
	iv := newInterval(v0, v1, vertical)
	t := (*interval.Tree)(tree)
	matches := t.Get(iv)
	common.Log.Info("testSide: iv=%v vertical=%t tree=%v matches=%d\n\tv0=%v\n\tv1=%v",
		iv, vertical, tree, matches, v0, v1)
	return len(matches) > 0
}

func findMinimalChords(hChords, vChords []*Chord) []*Chord {
	common.Log.Info("findMinimalChords: hChords=%d vChords=%d", len(hChords), len(vChords))

	// First find crossings
	crossings := findCrossings(hChords, vChords)
	common.Log.Info("findMinimalChords: crossings=%d", len(crossings))

	chordIdx := make(map[*Chord]int, len(hChords)+len(vChords))
	// Then tag and convert edge format
	for i, chord := range hChords {
		chordIdx[chord] = i
	}
	for i, chord := range vChords {
		chordIdx[chord] = i
	}

	//   var edges = crossings.map(function(c) {
	//     return [ c[0].number, c[1].number ]
	//   })
	edges := make([][2]int, len(crossings))
	for i, c := range crossings {
		edges[i] = [2]int{chordIdx[c.h], chordIdx[c.v]}
	}

	// Find independent set
	hIndices, vIndices := bipartiteIndependentSet(len(hChords), len(vChords), edges)
	common.Log.Info("hIndices=%d", len(hIndices))
	common.Log.Info("vIndices=%d", len(vIndices))

	// Convert into result format
	result := make([]*Chord, len(hIndices)+len(vIndices))
	for i, idx := range hIndices {
		result[i] = hChords[idx]
	}
	for i, idx := range vIndices {
		result[i+len(hIndices)] = vChords[idx]
	}
	common.Log.Info("result=%d", len(result))
	// panic("done bipartite")

	return result
}

type Crossing struct {
	h, v *Chord
}

// findCrossings returns the all intersections of horizontal and vertical chords.
func findCrossings(hChords, vChords []*Chord) []Crossing {
	// hTree := CreateIntervalTreeChords(hChords, "hChords")
	var crossings []Crossing
	// for _, v := range vChords {
	// 	// x := v.start.X
	// 	// !@#$ hChords has to be verticals for this query to work!.
	// 	// !@#$ Do the query!
	// 	// hTree.QueryPoint(v.start.Y, func(h *Side) bool {
	// 	// 	x := h.start.X
	// 	// 	if v.x0 <= x && x <= v.x1 {
	// 	// 		crossings = append(crossings, Crossing{h: h, v: v})
	// 	// 	}
	// 	// 	return false
	// 	// })
	// }
	return crossings
}

func splitSide(chord *Chord) {
	common.Log.Info("splitSide: %v", chord)
	// panic("splitSide")
	// //Store references
	// a := segment.start
	// b := segment.end
	// pa := a.prev
	// na := a.next
	// pb := b.prev
	// nb := b.next

	// // Fix concavity
	// a.concave = false
	// b.concave = false

	// // Compute orientation
	// ao := pa.Point.Cpt(segment.vertical) == a.Point.Cpt(segment.vertical)
	// bo := pb.Point.Cpt(segment.vertical) == b.Point.Cpt(segment.vertical)

	// if ao && bo {
	// 	//Case 1:
	// 	//            ^
	// 	//            |
	// 	//  --->A+++++B<---
	// 	//      |
	// 	//      V
	// 	a.prev = pb
	// 	pb.next = a
	// 	b.prev = pa
	// 	pa.next = b
	// } else if ao && !bo {
	// 	//Case 2:
	// 	//      ^     |
	// 	//      |     V
	// 	//  --->A+++++B--->
	// 	//
	// 	//
	// 	a.prev = b
	// 	b.next = a
	// 	pa.next = nb
	// 	nb.prev = pa
	// } else if !ao && bo {
	// 	//Case 3:
	// 	//
	// 	//
	// 	//  <---A+++++B<---
	// 	//      ^     |
	// 	//      |     V
	// 	a.next = b
	// 	b.prev = a
	// 	na.prev = pb
	// 	pb.next = na

	// } else if !ao && !bo {
	// 	//Case 3:
	// 	//            |
	// 	//            V
	// 	//  <---A+++++B--->
	// 	//      ^
	// 	//      |
	// 	a.next = nb
	// 	nb.prev = a
	// 	b.next = na
	// 	na.prev = b
	// }
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
