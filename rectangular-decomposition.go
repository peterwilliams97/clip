package clip

import (
	"fmt"
	"math"
	"sort"

	"github.com/biogo/store/interval"
	"github.com/unidoc/unidoc/common"
)

/*
	Coordinate origin is top-left

*/

const INF = 2 ^ 32

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

// Segment is a vertical or horizontal segment.
type Segment struct { // A chord?
	x0, x1     float64 // start and end of the interval in the vertical or horizontal direction.
	start, end *Vertex
	vertical   bool // Is this a vertical segment?
	number     int
}

func newSegment(start, end *Vertex, vertical bool) *Segment {
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
	}
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
	// clockwise = !clockwise

	common.Log.Debug("DecomposeRegion: paths=%d clockwise=%t", len(paths), clockwise)
	for i, path := range paths {
		common.Log.Debug("\t%3d:%+v", i, path)
	}
	common.Log.Debug("====================================")

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
				if clockwise {
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
				if clockwise {
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
			a := p[j]
			b := p[(j+1)%len(p)]
			if a.point.X == b.point.X {
				// hsegments are vertical !@#$
				hsegments = append(hsegments, newSegment(a, b, false))
			} else {
				// vsegments are horizontal !@#$
				vsegments = append(vsegments, newSegment(a, b, true))
			}
			if clockwise {
				b.next = a
			} else {
				a.prev, a.next = a, b
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
				leftsegments = append(leftsegments, newSegment(v, v.next, true))
			} else {
				rightsegments = append(rightsegments, newSegment(v, v.next, true))
			}
		}
	}

	lefttree := createIntervalTree(leftsegments)
	righttree := createIntervalTree(rightsegments)
	for i, v := range vertices {
		common.Log.Debug("i=%d v=%#v", i, v)
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
			queryPoint(righttree, v.point.X, func(h *Segment) bool {
				x := h.start.point.Y
				if closestDistance < x && x < y {
					closestDistance = x
					closestSegment = h
				}
				return false
			})
		} else {
			queryPoint(lefttree, v.point.X, func(h *Segment) bool {
				x := h.start.point.Y
				if y < x && x < closestDistance {
					closestDistance = x
					closestSegment = h
				}
				return false
			})
		}

		common.Log.Debug("closestSegment=%#v\n", closestSegment)

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
		var tree interval.Tree
		if direction < 0 {
			tree = righttree
		} else {
			tree = lefttree
		}
		treeDelete(tree, closestSegment)
		treeInsert(tree, newSegment(closestSegment.start, splitA, true))
		treeInsert(tree, newSegment(splitB, closestSegment.end, true))

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

// Stub
func createIntervalTree(segments []*Segment) interval.Tree {
	var tree interval.Tree
	return tree
}

func getDiagonals(vertices []*Vertex, npaths [][]*Vertex, vertical bool, tree interval.Tree) []*Segment {
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
	// First find crossings
	crossings := findCrossings(hdiagonals, vdiagonals)

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
	htree := createIntervalTree(hdiagonals)
	var crossings []Crossing
	for _, v := range vdiagonals {
		// x := v.start.point.X
		queryPoint(htree, v.start.point.Y, func(h *Segment) bool {
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

// bipartiteIndependentSet computes a maximum independent set for a bipartite graph.
// It takes O(sqrt(V) * E) time.
// * `n` is a the number of vertices in the first component
// * `m` is the number of vertices in the second component
// * `edges` is a list of edges in the bipartite graph represented by pairs of integers
// **Returns** A pair of lists representing the maximum independent set for the graph
// http://en.wikipedia.org/wiki/Maximum_independent_set
// http://en.wikipedia.org/wiki/Bipartite_graph
// A set is independent if and only if its complement is a vertex cover.
// PROOF: A set V of vertices is an independent set
//    IFF every edge in the graph is adjacent to at most one member of V
//    IFF every edge in the graph is adjacent to at least one member not in V
//    IFF the complement of V is a vertex cover.
func bipartiteIndependentSet(n, m int, edges [][2]int) ([]int, []int) {
	coverL, coverR := bipartiteVertexCover(n, m, edges)
	return complement(coverL, n), complement(coverR, m)
}

// func compareInt(a, b int) {
//   return a - b
// }
// complement returns [0:`n`) \ `list`
func complement(list []int, n int) []int {
	sort.Ints(list)

	result := make([]int, n-len(list))
	a, b := 0, 0
	for i := 0; i < n; i++ {
		if list[a] == i {
			a++
		} else {
			result[b] = i
			b++
		}
	}
	return result
}

// bipartiteVertexCover computes a minimum vertex cover of a bipartite graph.
//  `n`: number of vertices in the left component
//  `m`: number of vertices in the right component
//  `edges`:  list of edges from the left component connecting to the right component represented
//      by pairs of integers between 0 and n-1,m-1 respectively
// Returns A pair of lists representing the vertices in the left component and the right component
//   respectively which are in the cover.
// Internally, this implementation uses the Hopcroft-Karp algorithm and König's theorem to compute
// the minimal vertex cover of a bipartite graph in O(sqrt(V) * E) time.
// bipartiteMatching uses Hopscroft-Karp, this function uses König's theorem as in
//    http://tryalgo.org/en/matching/2016/08/05/konig/
// https://en.wikipedia.org/wiki/Hopcroft%E2%80%93Karp_algorithm
// https://en.wikipedia.org/wiki/K%C5%91nig%27s_theorem_(graph_theory)
func bipartiteVertexCover(n, m int, edges [][2]int) ([]int, []int) {
	match := bipartiteMatching(n, m, edges)

	// Initialize adjacency lists
	adjL := make([][]int, n)
	matchL := make([]int, n)
	matchCount := make([]int, n)
	coverL := make([]int, n)
	for i := 0; i < n; i++ {
		// adjL[i] = nil
		matchL[i] = -1
		// matchCount[i] = 0
		// coverL[i] = 0
	}
	adjR := make([][]int, m)
	matchR := make([]int, m)
	coverR := make([]int, m)
	for i := 0; i < m; i++ {
		// adjR[i] = nil
		matchR[i] = -1
		// coverR[i] = 0
	}

	// Unpack matching.
	for _, m := range match {
		s, t := m[0], m[1]
		matchL[s] = t
		matchR[t] = s
	}

	// Loop over edges.
	for _, e := range edges {
		s, t := e[0], e[1]
		if matchL[s] == t {
			cnt := matchCount[s]
			matchCount[s]++
			if cnt == 0 {
				continue
			}
		}
		adjL[s] = append(adjL[s], t)
		adjR[t] = append(adjR[t], s)
	}

	// Construct cover
	var left []int
	var right []int
	for i := 0; i < n; i++ {
		bpWalk(right, i, adjL, matchL, coverL, matchR, coverR)
	}
	for i := 0; i < m; i++ {
		bpWalk(left, i, adjR, matchR, coverR, matchL, coverL)
	}

	// Clean up any left over edges
	for i := 0; i < n; i++ {
		if coverL[i] == 0 && matchL[i] >= 0 {
			coverR[matchL[i]] = 1
			coverL[i] = 1 // !@#$ Does this have any effect?
			left = append(left, i)
		}
	}

	return left, right
}

// bipartite walk
func bpWalk(list []int, v int, adjL [][]int, matchL, coverL, matchR, coverR []int) {
	if coverL[v] != 0 || matchL[v] >= 0 {
		return
	}
	for v >= 0 {
		coverL[v] = 1
		adj := adjL[v]
		next := -1
		// !@#$ Seems like an inefficient way to find max u: !coverR[u]
		for _, u := range adj {
			if coverR[u] != 0 {
				continue
			}
			next = u
		}
		if next < 0 {
			break
		}
		coverR[next] = 1
		list = append(list, next)
		v = matchR[next]
	}
}

// bipartiteMatching finds a maximum bipartite matching in an unweighted graph.
//  The current implementation uses the Hopcroft-Karp algorithm and runs in O(sqrt(V) * E + V) time.
// * `n` is the number of vertices in the first component
// * `m` is the number of vertices in the second component
// * `edges` is the list of edges, represented by pairs of integers between 0 and n-1,m-1 respectively.
// **Returns** A list of edges representing the matching
// https://en.wikipedia.org/wiki/Matching_(graph_theory)
// https://en.wikipedia.org/wiki/Hopcroft%E2%80%93Karp_algorithm#Pseudocode
func bipartiteMatching(n, m int, edges [][2]int) [][2]int {
	// Initalize adjacency list, visit flag, distance.
	adjN := make([][]int, n)
	g1 := make([]int, n)
	dist := make([]int, n)
	for i := 0; i < n; i++ {
		g1[i] = -1
		adjN[i] = nil
		dist[i] = INF
	}
	adjM := make([][]int, m)
	g2 := make([]int, m)
	for i := 0; i < m; i++ {
		g2[i] = -1
		adjM[i] = nil
	}

	// Build adjacency matrix
	E := len(edges)
	for i := 0; i < E; i++ {
		e := edges[i]
		adjN[e[0]] = append(adjN[e[0]], e[1])
		adjM[e[1]] = append(adjM[e[1]], e[0])
	}

	// Why isn't adjM used any more? !@#$
	dmax := INF

	// Depth-first search?
	var dfs func(v int) bool
	dfs = func(v int) bool {
		if v < 0 {
			return true
		}
		adj := adjN[v]
		for _, u := range adj {
			pu := g2[u]
			dpu := dmax
			if pu >= 0 {
				dpu = dist[pu]
			}
			if dpu == dist[v]+1 {
				if dfs(pu) {
					g1[v] = u
					g2[u] = v
					return true
				}
			}
		}
		dist[v] = INF
		return false
	}

	// Run search
	toVisit := make([]int, n)
	matching := 0
	for {
		// Initialize queue
		count := 0
		for i := 0; i < n; i++ {
			if g1[i] < 0 {
				dist[i] = 0
				toVisit[count] = i
				count++
			} else {
				dist[i] = INF
			}
		}

		// Run BFS
		ptr := 0
		dmax = INF
		for ptr < count {
			v := toVisit[ptr]
			ptr++
			dv := dist[v]
			if dv < dmax {
				adj := adjN[v]
				l := len(adj)
				for j := 0; j < l; j++ {
					u := adj[j]
					pu := g2[u]
					if pu < 0 {
						if dmax == INF {
							dmax = dv + 1
						}
					} else if dist[pu] == INF {
						dist[pu] = dv + 1
						toVisit[count] = pu
						count++
					}
				}
			}
		}

		// Check for termination
		if dmax == INF {
			break
		}

		// Run DFS on each vertex in N
		for i := 0; i < n; i++ {
			if g1[i] < 0 {
				if dfs(i) {
					matching += 1
				}
			}
		}
	}

	// Construct result
	count := 0
	result := make([][2]int, matching)
	for i := 0; i < n; i++ {
		if g1[i] < 0 {
			continue
		}
		result[count] = [2]int{i, g1[i]}
		count++
	}

	if count != matching {
		panic("Didn't expect this.")
	}

	return result
}

// Generic intervals
type Int float64

func (c Int) Compare(b interval.Comparable) int {
	d := c - b.(Int)
	if isZero(float64(d)) {
		return 0
	}
	if d > 0 {
		return 1
	}
	return -1

}

// Interval is an interval over points in either the horizontal or vertical direction.
type Interval struct {
	*Segment
	id uintptr
	// Sub        []Interval
	// Payload interface{}
}

func newInterval(v0, v1 *Vertex, vertical bool) Interval {
	return Interval{Segment: newSegment(v0, v1, vertical)}
	// return Interval{
	// 	x0: Int(v0.point.Cpt(vertical)),
	// 	x1: Int(v1.point.Cpt(vertical)),
	// 	v0: v0,
	// 	v1: v1,
	// }
}

func testSegment(v0, v1 *Vertex, tree interval.Tree, vertical bool) bool {
	i := newInterval(v0, v1, vertical)
	matches := tree.Get(i)
	return len(matches) > 0
}

func treeDelete(tree interval.Tree, s *Segment) {
	i := Interval{Segment: s}
	tree.Delete(i, false)
}

func treeInsert(tree interval.Tree, s *Segment) {
	i := Interval{Segment: s}
	if err := tree.Insert(i, false); err != nil {
		panic(err)
	}
}

func queryPoint(tree interval.Tree, x float64, cb func(s *Segment) bool) bool {
	var matched bool
	ok := tree.Do(func(e interval.Interface) bool {
		// s := e.(Int)
		// matched = s == x
		i := e.(Interval)
		matched := cb(i.Segment)
		return matched
	})
	if matched != ok {
		panic("queryPoint")
	}
	return matched
}

// function testSegment(a, b, tree, direction) {
//   var ax = a.point[direction^1]
//   var bx = b.point[direction^1]
//   return !!tree.queryPoint(a.point[direction], function(s) {
//     var x = s.start.point[direction^1]
//     if(ax < x && x < bx) {
//       return true
//     }
//     return false
//   })
// }

func (i Interval) Overlap(b interval.Range) bool {
	var x0, x1 float64
	switch bc := b.(type) {
	case Interval:
		x0 = bc.x0
		x1 = bc.x1
	case *Mutable:
		x0, x1 = bc.x0, bc.x1
	default:
		panic("unknown type")
	}

	// Half-open interval indexing.
	return i.x1 > x0 && i.x0 < x1
}
func (i Interval) ID() uintptr                  { return i.id }
func (i Interval) Start() interval.Comparable   { return Int(i.x0) }
func (i Interval) End() interval.Comparable     { return Int(i.x1) }
func (i Interval) NewMutable() interval.Mutable { return &Mutable{Segment: i.Segment, id: i.id} }
func (i Interval) String() string {
	return fmt.Sprintf("[%g,%g)#%d",
		i.x0, i.x1, i.id)
}

type Mutable struct {
	*Segment
	id uintptr
}

func (m *Mutable) Start() interval.Comparable     { return Int(m.x0) }
func (m *Mutable) End() interval.Comparable       { return Int(m.x1) }
func (m *Mutable) SetStart(c interval.Comparable) { m.x0 = float64(c.(Int)) }
func (m *Mutable) SetEnd(c interval.Comparable)   { m.x1 = float64(c.(Int)) }

// func (t *interval.Tree) queryPoint(x float64, f func(h *Segment)) {
// }
// func (t *interval.Tree) insert(h *Segment) {
// }
// func (t *interval.Tree) remove(h *Segment) {
// }
