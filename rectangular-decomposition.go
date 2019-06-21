package clip

import (
	"fmt"

	"github.com/biogo/store/interval"
	"github.com/unidoc/unipdf/v3/common"
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
			common.Log.Debug("j=%d k=%d\n\tv0=%v\n\tv1=%v", j, k, v0, v1)
			if v0.X == v1.X {
				vSides = append(vSides, NewSide(v0, v1))
			} else {
				hSides = append(hSides, NewSide(v0, v1))
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

	// Find the minimum set of splitting chords.
	splitters := findMinimalChords(hChords, vChords)
	common.Log.Info("**** splitters=%d", len(splitters))
	for i, c := range splitters {
		common.Log.Info("%6d: %v other=%v", i, c, c.OtherEnd())
	}

	var splittedContours [][]*Vertex
	// for _, contour := range contours {
	newContours := spiltContourOnChords(contours[0], splitters)
	splittedContours = append(splittedContours, newContours...)
	// }

	// Cut all the splitting chords
	// for _, splitter := range splitters {
	// 	splitSide(splitter)
	// }

	// Split all concave vertices
	// splitConcave(vertices)

	// Return regions
	// rectangles := findRegions(splitContours)
	rectangles := polygonToRectangles(splittedContours)

	common.Log.Info("*** %d rectangles", len(rectangles))
	for i, r := range rectangles {
		common.Log.Info("%6d: %v", i, r)
	}
	return rectangles
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
				upVSides = append(upVSides, NewSide(v, v.next))
			} else { //            v
				downVSides = append(downVSides, NewSide(v, v.next))
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
		tree.Insert(NewSide(closestSide.start, splitA))
		tree.Insert(NewSide(splitB, closestSide.end))

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
	common.Log.Info("getChords: vertices=%d %s", len(vertices), directionName(vertical))

	var concave []*Vertex
	for _, v := range vertices {
		v.Validate()
		if v.concave {
			concave = append(concave, v)
		}
	}

	// sort.Slice(concave, func(i, j int) bool {
	// 	vi, vj := concave[i], concave[j]
	// 	if vi.Cpt(vertical) != vj.Cpt(vertical) {
	// 		return vi.Cpt(vertical) < vj.Cpt(vertical)
	// 	}
	// 	return vi.Cpt(!vertical) < vj.Cpt(!vertical)
	// })

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
	sigMap := map[string]struct{}{}
	for i, v := range concave {
		xp, xv, xn := v.prev.Cpt(vertical), v.Cpt(vertical), v.next.Cpt(vertical)
		inwards := xp != xv
		var increasing bool
		if inwards {
			increasing = xv > xp
		} else {
			increasing = xv > xn
		}
		common.Log.Info("orientation i=%d %s inwards=%t increasing=%t (%g %g %g)",
			i, directionName(vertical), inwards, increasing, xp, xv, xn)
		c := findChord(v, tree, vertical, increasing)
		if c == nil {
			continue
		}
		sig := rectString(c)
		_, dup := sigMap[sig]
		common.Log.Info("candidate i=%d dup=%5t %s", i, dup, c)
		if dup {
			continue
		}
		sigMap[sig] = struct{}{}
		chords = append(chords, c)

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

		// if a.iContour == b.iContour {
		// 	n := len(contours[a.iContour])
		// 	d := (a.index - b.index + n) % n
		// 	common.Log.Info("i=%d: n=%d d=%d", i, n, d)
		// 	if d == 1 || d == n-1 {
		// 		// Adjacent points
		// 		continue
		// 	}
		// }
		// if !testSide(a, b, tree, vertical) {
		// 	// Check orientation of diagonal
		// 	// !@#$ Find the chords!
		// 	// chords = append(chords, NewSide(a, b, vertical))
		// }

	}
	common.Log.Info("chords=%d", len(chords))
	for i, c := range chords {
		common.Log.Info("%4d: %s", i, c)
	}
	if len(chords) == 0 {
		panic("no chords")
	}
	return chords
}

// findChord returns the closest chord from `vertex` to the sides in `tree`.
func findChord(vertex *Vertex, tree *IntervalTree, vertical, increasing bool) *Chord {
	xx, yy := vertex.Point.Cpt(vertical), vertex.Point.Cpt(!vertical)
	common.Log.Info("findChord: vertex=%v %s increasing=%t xx=%g yy=%g",
		vertex.Point, directionName(vertical), increasing, xx, yy)

	distance := infinity
	var closest *Side
	tree.QueryPoint(yy, func(r Rectilinear) bool {
		s := r.(*Side)
		x := s.start.Cpt(vertical)
		var dx float64
		if increasing {
			dx = x - xx
		} else {
			dx = xx - x
		}
		if 0 < dx && dx < distance {
			closest = s
			distance = dx
		}
		common.Log.Debug(" query: r=%s x=%g xx=%g dx=%g distance=%g closest=%v",
			rectString(s), x, xx, dx, distance, closest)
		return false
	})

	if closest == nil {
		panic("no chords")
		return nil
	}
	chord := Chord{v: vertex, s: closest}
	x0, x1, _, _ := chord.X0X1YVert()
	if x0 == x1 {
		panic(fmt.Errorf("Bad chord: %s", chord))
	}
	return &chord
}

// testSide returns true if segment [v0,v1] intersects an existing segment.
func testSide(v0, v1 *Vertex, tree *IntervalTree, vertical bool) bool {
	iv := newInterval(v0, v1, vertical)
	t := (*interval.Tree)(tree)
	matches := t.Get(iv)
	common.Log.Info("testSide: iv=%v %s tree=%v matches=%d\n\tv0=%v\n\tv1=%v",
		iv, directionName(vertical), tree, matches, v0, v1)
	return len(matches) > 0
}

// findMinimalChords returns that minimal set of chords that intersect (possibly by sharing vertices
// with) all the chords in `hChords` and `vChords`. hChords and vChords are horizontal and vertical
// respectively.
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
	common.Log.Info("***** hIndices=%d", len(hIndices))
	common.Log.Info("***** vIndices=%d", len(vIndices))

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
	common.Log.Info("================findCrossings")
	common.Log.Info("  hChords=%d", len(hChords))
	for i, c := range hChords {
		common.Log.Info("%4d: %s", i, c)
	}
	common.Log.Info("  vChords=%d", len(vChords))
	for i, c := range vChords {
		common.Log.Info("%4d: %s", i, c)
	}

	vTree := CreateIntervalTreeChords(vChords, "vChords")
	var crossings []Crossing
	common.Log.Info("-----------------------")
	for i, h := range hChords {
		common.Log.Info("%4d: %s", i, h)
		x0, x1, y, _ := h.X0X1YVert()
		vTree.QueryPoint(y, func(r Rectilinear) bool {
			_, _, x, _ := r.X0X1YVert()
			v := r.(*Chord)
			common.Log.Info("%8s %s %g<=%g<=%g=%t", "", v, x0, x, x1, x0 <= x && x <= x1)
			if x0 <= x && x <= x1 {

				crossings = append(crossings, Crossing{h: h, v: v})
			}
			return false
		})
	}
	common.Log.Info("  crossings=%d", len(crossings))
	for i, c := range crossings {
		common.Log.Info("%4d: h=%v v=%v", i, *c.h, *c.v)
	}
	return crossings
}

// spiltContourOnChords returns the contours that result from splitting `contour` along chords.
// 1) Find diagonals.
// 2a) Find chords.
// 2b) Split sides on chord intersections. findIntersection()
// 2c) Add the chord + splitting vertex to diagonals.
// 3) Split contour on diagonals.
// 4) Return the resulting contours.
func spiltContourOnChords(contour []*Vertex, chords []*Chord) [][]*Vertex {
	common.Log.Info("spiltContourOnChords: contour=%d chords=%d", len(contour), len(chords))

	var diagonals [][2]int
	matched := map[int]struct{}{}

	// 1) Find diagonals.
	for i, c := range chords {
		k := findOpposite(contour, c)
		common.Log.Info("  findOpposite(%s) -> %d", c, k)
		if k < 0 {
			continue
		}
		diagonals = append(diagonals, [2]int{i, k})
		matched[i] = struct{}{}
		matched[k] = struct{}{}
	}

	common.Log.Info("diagonals=%d", len(diagonals))
	for i, diag := range diagonals {
		common.Log.Info("%6d: %v", i, diag)
	}

	// 2a) Find chords.
	var intersections []Intersection
	for i, c := range chords {
		if _, ok := matched[i]; ok {
			continue
		}
		x, ok := findIntersection(contour, c)
		common.Log.Info("  findIntersection(%s) -> %v %t", c, x, ok)
		if !ok {
			continue
		}
		intersections = append(intersections, x)
		matched[i] = struct{}{}
	}
	common.Log.Info("intersections=%d", len(intersections))
	for i, x := range intersections {
		common.Log.Info("%6d: %v %v intersects %v-%v at %v",
			i, x.c, x.e0.Point, x.e1.Point, x.p)
	}

	newContour := splitSidesByChords(contour, intersections)

	common.Log.Info("***3 newContour: %d ", len(newContour))
	for i, v := range newContour {
		common.Log.Info("%6d: %v", i, v)
	}
	validateCountour(newContour)

	return [][]*Vertex{newContour}
}

func (x Intersection) validate() {
	e0, e1 := x.e0, x.e1
	if e0.next != e1 || e1.prev != e0 {
		panic("not a side")
	}
	if e0.Point.Equals(e1.Point) {
		panic("duplicate vertex")
	}
	v0 := x.c.v
	if v0.next == e0 || v0.prev == e1 {
		common.Log.Info("  adjacent 1")
		common.Log.Info("\tv0=%v", v0)
		common.Log.Info("\te0=%v", e0)
		common.Log.Info("\te1=%v", e1)
		panic("v==e 1")
	}
	if v0 == e0 || v0 == e1 {
		common.Log.Info("  adjacent 2")
		common.Log.Info("\tv0=%v", v0)
		common.Log.Info("\te0=%v", e0)
		common.Log.Info("\te1=%v", e1)
		panic("v==e 2")
	}
}

// splitSidesByChords splits all the sides of `contours` intersected by `chords` and returns the
// resulting contour
func splitSidesByChords(contour []*Vertex, intersections []Intersection) []*Vertex {
	common.Log.Info("splitSidesByChords: %d vertices %d intersections", len(contour), len(intersections))
	validateCountour(contour)

	bySide := map[string][]Intersection{}
	for _, x := range intersections {
		x.validate()
		side := fmt.Sprintf("%v#%v", x.e0.Point, x.e1.Point)
		bySide[side] = append(bySide[side], x)
	}

	var opposites []*Vertex
	for _, sideIntersections := range bySide {
		opps := doSideSplits(sideIntersections)
		opposites = append(opposites, opps...)
		validateCountour(append(contour, opposites...))
	}
	newContour := append(contour, opposites...)
	validateCountour(newContour)
	return rebuildCountour(newContour)
}

func doSideSplits(intersections []Intersection) []*Vertex {
	common.Log.Info("doSideSplits: %d intersections", len(intersections))

	var opposites []*Vertex
	e0 := intersections[0].e0
	e1 := intersections[0].e1
	v0 := e0

	for _, x := range intersections[1:] {
		if x.e0 != e0 || x.e1 != e1 {
			panic("Different side")
		}
		v := &Vertex{
			index: -1,
			Point: x.p,
			prev:  v0,
		}
		v0.next = v
		opposites = append(opposites, v)
		v0 = v
	}
	v0.next = e1
	e1.prev = v0
	return opposites
}

// func (x *Intersection) split() *Vertex {
// 	e0, e1 := x.e0, x.e1
// 	v0 := chords[ic].v
// 	v1 := &Vertex{
// 		prev: e0,
// 		next: e1,
// 	}
// 	e0.next = v1
// 	e1.prev = v1

// 	vertical := e0.X == e1.X
// 	if vertical {
// 		v1.X = e0.X
// 		v1.Y = v0.Y
// 	} else {
// 		v1.X = v0.X
// 		v1.Y = e0.Y
// 	}
// 	common.Log.Info(" splitOnChord\n\tv0=%v\n\te0=%v\n\te1=%v\n\tv1=%v", v0, e0, e1, v1)
// 	v1.Validate()
// 	return v1
// }

// // splitOnChord splits on chord from `v0` to edge `e0`,`e1`.
// func splitOnChord(v0, e0, e1 *Vertex) (a, b []*Vertex) {

// 	// Add vertex v1 that splits e0 and e1

// 	return splitOnDiagonal(contour, v0, v1)
// }

// func addChordVertex(contour[]*Vertex v0, e0, e1 *Vertex)  []*Vertex {
// 	// common.Log.Info("  splitOnChord")
// 	// common.Log.Info("\tv0=%v", v0)
// 	// common.Log.Info("\te0=%v", e0)
// 	// common.Log.Info("\te1=%v", e1)
// 	if v0 == e0 {
// 		panic("v0==e0")
// 	}
// 	if v0 == e1 {
// 		panic("v0==e1")
// 	}

// 	// Add vertex v1 that splits e0 and e1
// 	v1 := &Vertex{
// 		prev: e0,
// 		next: e1,
// 	}
// 	e0.next = v1
// 	e1.prev = v1

// 	vertical := e0.X == e1.X
// 	if vertical {
// 		v1.X = e0.X
// 		v1.Y = v0.Y
// 	} else {
// 		v1.X = v0.X
// 		v1.Y = e0.Y
// 	}
// 	v1.Validate()

// }

// splitOnDiagonal splits `contour` on diagonal from `v0` to `v1`.
func splitOnDiagonal(contour []*Vertex, v0, v1 *Vertex) (a, b []*Vertex) {
	common.Log.Info("splitOnDiagonal: %d vertices v0=%v v1=%v", len(contour), v0.Point, v1.Point)

	//        +---+
	//      v1| a |v0
	//    +-<-+···+-<-+
	//    |     b     |
	//    +-----------+
	a0, a1 := cv(v0), cv(v1)
	a1.next = a0
	a0.prev = a1
	a1.prev.next = a1
	a0.next.prev = a0

	b0, b1 := cv(v0), cv(v1)
	b0.next = b1
	b1.prev = b0
	b0.prev.next = b0
	b1.next.prev = b1

	for v := a0; v != a0; v = v.next {
		a = append(a, v)
	}
	for v := b0; v != b0; v = v.next {
		b = append(b, v)
	}

	return a, b
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

// findOpposite returns axis-parallel diagonals
// !@#$ replace with efficient version
func findOpposite(contour []*Vertex, c *Chord) int {
	o := c.OtherEnd()
	// common.Log.Info("findOpposite: c=%s other=%v", c, o)
	for i, v := range contour {
		m := v.Point.Equals(o)
		if m {
			return i
		}
	}
	// panic("no opposite")
	return -1
}

type Intersection struct {
	c      *Chord
	p      Point
	e0, e1 *Vertex
}

// findIntersection returns the side of `contour` intersected by `c`.
func findIntersection(contour []*Vertex, c *Chord) (Intersection, bool) {
	// common.Log.Info("findIntersection: c=%s", c)
	// n := len(contour)
	v := c.v

	for i0, e0 := range contour {
		i1 := (i0 + 1) % len(contour)
		e1 := contour[i1]
		// if adjacent(v0.index, c.v.index, n) || adjacent(v1.index, c.v.index, n) {
		// 	continue
		// }
		if e0.Point.Equals(c.v.Point) || e1.Equals(c.v.Point) {
			continue
		}
		s := NewSide(e0, e1)
		m := c.Intersects(s)
		if !m {
			continue
		}

		// p is the point that splits e0 and e1.
		var p Point
		vertical := e0.X == e1.X
		if vertical {
			p = Point{X: e0.X, Y: v.Y}
		} else {
			p = Point{Y: e0.Y, X: v.X}

		}
		return Intersection{c, p, e0, e1}, true

	}
	// panic("no intersection")
	return Intersection{}, false
}

func adjacent(i, j, n int) bool {
	k := (i - j + n) % n
	return k <= 0 || k == n-1
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

func polygonToRectangles(contours [][]*Vertex) []Rect {
	var allRects []Rect
	for _, countour := range contours {
		rect := contourToRectangle(countour)
		allRects = append(allRects, rect)
	}
	return allRects
}

func contourToRectangle(contour []*Vertex) Rect {
	isRectangle(contour)
	x, y := contour[0].X, contour[0].Y
	r := Rect{Llx: x, Lly: y, Urx: x, Ury: y}
	for _, v := range contour[1:] {
		x, y = v.X, v.Y
		if x < r.Llx {
			r.Llx = x
		} else if x > r.Urx {
			r.Urx = x
		}
		if y < r.Lly {
			r.Lly = y
		} else if y > r.Ury {
			r.Ury = y
		}
	}
	return r
}

func isRectangle(contour []*Vertex) bool {
	validateCountour(contour)
	if len(contour) != 4 {
		common.Log.Error("isRectangle: %s", showContour(contour))
		panic("contourToRectangle 1")
	}
	v := contour[0]
	counts := map[string]int{}
	for i := 0; i < 4; i++ {
		s := fmt.Sprintf("%v", v.Point)
		counts[s]++
		if counts[s] > 1 {
			common.Log.Error("isRectangle: %s", showContour(contour))
			panic("duplicate point")
		}
		if v.next.prev != v {
			common.Log.Error("isRectangle: %s", showContour(contour))
			panic("bad next link")
		}
		if v.prev.next != v {
			common.Log.Error("isRectangle: %s", showContour(contour))
			panic("bad prev link")
		}
		if v.next.X != v.next.X && v.next.Y != v.next.Y {
			common.Log.Error("isRectangle: %s", showContour(contour))
			panic("not rectangle")
		}
		v = v.next
	}
	if v != contour[0] {
		common.Log.Error("isRectangle: %s", showContour(contour))
		panic("not closed")
	}
	return true
}

// func findRegions(vertices []*Vertex) []Rect {
// 	n := len(vertices)
// 	common.Log.Info("findRegions: %d vertices", len(vertices))
// 	for i := 0; i < n; i++ {
// 		v := vertices[i]
// 		common.Log.Info("%4d: %p %v %v %v", i, v, *v, v.prev.Point, v.next.Point)
// 	}
// 	for _, v := range vertices {
// 		v.Validate()
// 	}
// 	common.Log.Info("~~~~~~~~~~~~~~~~~~~~~~~~~")
// 	//   0  1  2  3
// 	// 0 +--+  +--+
// 	//   |  |  |  |
// 	// 1 |  +--+  |
// 	//   |        |
// 	// 2 +--------+
// 	// Walk over vertex list
// 	var rectangles []Rect
// 	var count int
// 	for i, v := range vertices {
// 		common.Log.Info("i=%d: v=%s", i, v)
// 		if v.visited {
// 			continue
// 		}
// 		// Walk along loop
// 		lo := Point{infinity, infinity}
// 		hi := Point{-infinity, -infinity}
// 		for ; !v.visited; v = v.next {
// 			lo.X = math.Min(v.X, lo.X)
// 			hi.X = math.Max(v.X, hi.X)
// 			lo.Y = math.Min(v.Y, lo.Y)
// 			hi.Y = math.Max(v.Y, hi.Y)
// 			v.visited = true
// 			common.Log.Info(" visit %d:  %s %v %v", count, *v, lo, hi)
// 			count++
// 		}
// 		r := Rect{Llx: lo.X, Lly: lo.Y, Urx: hi.X, Ury: hi.Y}
// 		rectangles = append(rectangles, r)
// 		common.Log.Info("i=%d: %d rectangles: %+v", i, len(rectangles)-1, r)
// 	}
// 	return rectangles
// }
