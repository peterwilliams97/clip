package clip

import (
	"math"
	"sort"

	"github.com/biogo/store/interval"
	"github.com/unidoc/unidoc/common"
)

type Vertex struct {
	point   Point
	iPath   int
	index   int
	concave bool
	next    *Vertex
	prev    *Vertex
	visited bool
}

type Segment struct {
	point      Point
	start, end *Vertex
	number     int
}

func newSegment(start, end *Vertex, direction bool) *Segment {
	var a, b float64
	if direction {
		a = start.point.X
		b = end.point.X
	} else {
		a = start.point.Y
		b = end.point.Y
	}
	var point Point
	if a < b {
		point = Point{a, b}
	} else {
		point = Point{b, a}
	}
	return &Segment{
		point:     point,
		start:     start,
		end:       end,
		direction: direction,
		number:    -1,
	}
}

// DecomposeRegion breaks `paths` into non-overlapping rectangles.
func DecomposeRegion(paths []Path, clockwise bool) []Rect {

	common.Log.Info("DecomposeRegion: paths=%d", len(paths))
	for i, path := range paths {
		common.Log.Info("\t%3d:%+v", i, path)
	}
	common.Log.Info("====================================")

	// First step: unpack all vertices into internal format.
	var vertices []Vertex

	npaths := make([][]Vertex, len(paths))
	for i, path := range paths {
		n := len(path)
		prev := path[n-3]
		cur := path[n-2]
		next := path[n-1]
		common.Log.Info("DecomposeRegion: %3d: path=%+v n=%d\n\t prev=%#v\n\t  cur=%#v\n\t next=%#v",
			i, path, n, prev, cur, next)
		for j := 0; j < n; j++ {
			prev = cur
			cur = next
			next = path[j]
			common.Log.Info("j=%d\n\t prev=%+v\n\t  cur=%+v\n\t next=%+v", j, prev, cur, next)
			concave := false
			if prev.X == cur.X {
				if next.X == cur.X {
					continue
				}
				dir0 := prev.Y < cur.Y
				dir1 := cur.X < next.X
				concave = dir0 == dir1
				common.Log.Info("  @1 dir0=%t dir1=%t concave=%t", dir0, dir1, concave)
			} else {
				if next.Y == cur.Y {
					continue
				}
				dir0 := prev.X < cur.X
				dir1 := cur.Y < next.Y
				concave = dir0 != dir1
				common.Log.Info("  @1 dir0=%t dir1=%t concave=%t", dir0, dir1, concave)
			}
			if clockwise {
				concave = !concave
			}
			vtx := Vertex{
				point:   cur,
				iPath:   i,
				index:   (j + n - 1) % n,
				concave: concave,
			}
			common.Log.Info("vtx=%+v", vtx)
			npaths[i] = append(npaths[i], vtx)
			vertices = append(vertices, vtx)
		}
	}

	// Next build interval trees for segments, link vertices into a list
	var hsegments []*Segment
	var vsegments []*Segment

	for i := 0; i < len(npaths); i++ {
		p := npaths[i]
		for j := 0; j < len(p); j++ {
			a := &p[j]
			b := &p[(j+1)%len(p)]
			if a.point.X == b.point.Y {
				hsegments = append(hsegments, newSegment(a, b, false))
			} else {
				vsegments = append(vsegments, newSegment(a, b, true))
			}
			if clockwise {
				a.prev = b
				b.next = a
			} else {
				a.next = b
				b.prev = a
			}
		}
	}
	htree := createIntervalTree(hsegments)
	vtree := createIntervalTree(vsegments)

	// Find horizontal and vertical diagonals.
	hdiagonals := getDiagonals(vertices, npaths, false, vtree)
	vdiagonals := getDiagonals(vertices, npaths, true, htree)

	// Find all splitting edges
	splitters := findSplitters(hdiagonals, vdiagonals)

	// Cut all the splitting diagonals
	for _, splitter := range splitters {
		splitSegment(&splitter)
	}

	// Split all concave vertices
	splitConcave(vertices)

	// Return regions
	return findRegions(vertices)
}

func splitConcave(vertices []Vertex) {
	common.Log.Info("splitConcave: vertices=%d", len(vertices))
	for i, v := range vertices {
		common.Log.Info("\t%3d: %+v", i, v)
	}
	common.Log.Info("=============================")
	// First step: build segment tree from vertical segments
	var leftsegments []*Segment
	var rightsegments []*Segment

	for i, v := range vertices {
		common.Log.Info("\t%3d: %+v", i, v)
		if v.next.point.Y == v.point.Y {
			if v.next.point.X < v.point.X {
				leftsegments = append(leftsegments, newSegment(&v, v.next, true))
			} else {
				rightsegments = append(rightsegments, newSegment(&v, v.next, true))
			}
		}
	}

	lefttree := createIntervalTree(leftsegments)
	righttree := createIntervalTree(rightsegments)
	for _, v := range vertices {
		if !v.concave {
			continue
		}

		// Compute orientation
		y := v.point.Y
		var direct bool
		if v.prev.point.X == v.point.X {
			direct = v.prev.point.Y < y
		} else {
			direct = v.next.point.Y < y
		}
		direction := 1
		if direct {
			direction = -1
		}

		// Scan a horizontal ray
		var closestSegment *Segment
		closestDistance := infinity * float64(direction)
		if direction < 0 {
			righttree.queryPoint(v.point.X, func(h *Segment) {
				x := h.start.point.Y
				if x < y && x > closestDistance {
					closestDistance = x
					closestSegment = h
				}
			})
		} else {
			lefttree.queryPoint(v.point.X, func(h *Segment) {
				x := h.start.point.Y
				if x > y && x < closestDistance {
					closestDistance = x
					closestSegment = h
				}
			})
		}

		// Create two splitting vertices
		point := Point{v.point.X, closestDistance}
		splitA := Vertex{point: point}
		splitB := Vertex{point: point}

		// Clear concavity flag
		v.concave = false

		// Split vertices
		splitA.prev = closestSegment.start
		closestSegment.start.next = &splitA
		splitB.next = closestSegment.end
		closestSegment.end.prev = &splitB

		// Update segment tree
		var tree IntervalTree
		if direction < 0 {
			tree = righttree
		} else {
			tree = lefttree
		}
		tree.remove(closestSegment)
		tree.insert(newSegment(closestSegment.start, &splitA, true))
		tree.insert(newSegment(&splitB, closestSegment.end, true))

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
			splitA.next = &v
			splitB.prev = v.prev
		} else {
			// Case 2
			//     |       ^
			//     V       |
			// <---*+++++++X
			//             |
			//             |
			splitA.next = v.next
			splitB.prev = &v
		}

		// Fix up links
		splitA.next.prev = &splitA
		splitB.prev.next = &splitB
	}
}

type IntervalTree struct{}

// type Diagonal struct{}
type Splitter struct{}

// Stub
func createIntervalTree(segments []*Segment) *interval.IntTree {
	tree := interval.IntTree{}
	return &tree
}

func getDiagonals(vertices []Vertex, npaths [][]Vertex, direction bool, tree *interval.IntTree) []*Segment {
	var concave []Vertex
	for _, v := range vertices {
		if v.concave {
			concave = append(concave, v)
		}
	}
	if direction {
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
		if direction {
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
			if !testSegment(a, b, tree, direction) {
				// Check orientation of diagonal
				diagonals = append(diagonals, newSegment(&a, &b, direction))
			}
		}
	}
	return diagonals
}

func testSegment(a, b Vertex, tree IntervalTree, direction bool) bool {
	ax := a.point.Cpt(direction)
	bx := b.point.Cpt(direction)

	return false
}
function testSegment(a, b, tree, direction) {
  var ax = a.point[direction^1]
  var bx = b.point[direction^1]
  return !!tree.queryPoint(a.point[direction], function(s) {
    var x = s.start.point[direction^1]
    if(ax < x && x < bx) {
      return true
    }
    return false
  })
}

// SegInterval is an interval of segments
type SegInterval struct {
	x float64
	UID        uintptr
	// Payload    interface{}
}

func (i SegInterval) Overlap(b interval.IntRange) bool {
	// Half-open interval indexing.
	return i.End > b.Start && i.start < b.End
}
func (i SegInterval) ID() uintptr              { return i.UID }
func (i SegInterval) Range() interval.IntRange { return interval.IntRange{i.start, i.End} }
func (i SegInterval) String() string {
	return fmt.Sprintf("[%d,%d)#%d", i.start, i.End, i.UID)
}

func findSplitters(hdiagonals, vdiagonals []*Segment) []Splitter {
	return nil
}

func splitSegment(splitter *Splitter) {
}

func (t *IntervalTree) queryPoint(x float64, f func(h *Segment)) {
}
func (t *IntervalTree) insert(h *Segment) {
}
func (t *IntervalTree) remove(h *Segment) {
}

func findRegions(vertices []Vertex) []Rect {
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
		for ; !v.visited; v = *v.next {
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
