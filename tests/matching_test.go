package clip_test

import (
	"testing"

	"github.com/peterwilliams97/clip"
	"github.com/unidoc/unipdf/v3/common"
)

var showAllMatches bool

func init() {
	showAllMatches = false
}

func TestVertexCover(t *testing.T) {
	n := 2
	m := 3
	edges := [][2]int{
		[2]int{0, 0},
		[2]int{0, 2},
		[2]int{1, 1},
		[2]int{0, 1},
	}
	expL, expR := []int{0, 1}, []int{}
	coverL, coverR := clip.BipartiteVertexCover(n, m, edges)
	if !eqSl(coverL, expL) {
		t.Fatalf("coverL != expL\n\t  expL=%v\n\tcoverL=%v\n\t  expL=%v\n\tcoverL=%v",
			expL, coverL, expR, coverR)
	}
	if !eqSl(coverR, expR) {
		t.Fatalf("coverR != expR\n\t  expL=%v\n\tcoverL=%v\n\t  expL=%v\n\tcoverL=%v",
			expL, coverL, expR, coverR)
	}
	for _, test := range coverCases {
		test.run(t)
	}
}

type coverTest struct {
	n, m       int
	edges      [][2]int
	expL, expR []int
}

var coverCases = []coverTest{
	coverTest{0, 0, [][2]int{}, []int{}, []int{}},
	coverTest{2, 3, [][2]int{
		[2]int{0, 0},
		[2]int{0, 2},
		[2]int{1, 1},
		[2]int{0, 1},
	}, []int{0, 1}, []int{},
	},
	coverTest{1, 3, [][2]int{
		[2]int{0, 0},
		[2]int{0, 1},
		[2]int{0, 2},
	}, []int{0}, []int{}},
	coverTest{2, 2, [][2]int{
		[2]int{0, 1},
		[2]int{1, 0},
	}, []int{0, 1}, []int{}},
	coverTest{3, 3, [][2]int{
		[2]int{2, 0},
		[2]int{1, 1},
		[2]int{0, 2},
	}, []int{0, 1, 2}, []int{}},
	coverTest{3, 2, [][2]int{
		[2]int{2, 0},
		[2]int{1, 1},
		[2]int{0, 0},
	}, []int{1}, []int{0}},
	coverTest{3, 3, [][2]int{
		[2]int{2, 0},
		[2]int{2, 2},
		[2]int{1, 1},
		[2]int{0, 0},
	}, []int{0, 1, 2}, []int{}},
	coverTest{3, 2, [][2]int{
		[2]int{2, 0},
		[2]int{2, 1},
		[2]int{1, 1},
		[2]int{0, 0},
	}, []int{0}, []int{1}},
	coverTest{3, 2, [][2]int{
		[2]int{2, 0},
		[2]int{2, 1},
		[2]int{1, 1},
		[2]int{0, 1},
	}, []int{2}, []int{1}},
	coverTest{4, 2, [][2]int{
		[2]int{0, 0},
		[2]int{1, 0},
		[2]int{2, 1},
		[2]int{3, 1},
	}, []int{}, []int{0, 1}},
	// matchingTest{3, 3, [][2]int{
	// 	[2]int{0, 0},
	// 	[2]int{0, 1},
	// 	[2]int{0, 2},
	// 	[2]int{1, 1},
	// 	[2]int{2, 0},
	// 	[2]int{2, 2},
	// }, 3},
	// matchingTest{3, 3, [][2]int{
	// 	[2]int{0, 1},
	// 	[2]int{0, 2},
	// 	[2]int{1, 0},
	// 	[2]int{1, 2},
	// 	[2]int{2, 0},
	// 	[2]int{2, 1},
	// }, 3},
	// matchingTest{3, 3, [][2]int{
	// 	[2]int{0, 1},
	// 	[2]int{0, 2},
	// 	[2]int{1, 0},
	// 	[2]int{1, 2},
	// 	[2]int{2, 0},
	// 	[2]int{2, 1},
	// 	[2]int{2, 2},
	// }, 3},
	// matchingTest{5, 5, [][2]int{
	// 	[2]int{0, 0},
	// 	[2]int{0, 1},
	// 	[2]int{1, 0},
	// 	[2]int{2, 1},
	// 	[2]int{2, 2},
	// 	[2]int{3, 2},
	// 	[2]int{3, 3},
	// 	[2]int{3, 4},
	// 	[2]int{4, 4},
	// }, 5},
	coverTest{4, 3, [][2]int{
		[2]int{0, 0},
		[2]int{0, 1},
		[2]int{1, 0},
		[2]int{1, 1},
		[2]int{1, 2},
		[2]int{2, 2},
		[2]int{2, 1},
		[2]int{3, 0},
		[2]int{3, 2},
	}, []int{}, []int{0, 1, 2}},
}

func (test coverTest) run(t *testing.T) {
	test.runOne(t)
	// rev := coverTest{
	// 	n:     test.m,
	// 	m:     test.n,
	// 	edges: make([][2]int, len(test.edges)),
	// 	expL:  test.expR,
	// 	expR:  test.expL,
	// }
	// for i, e := range test.edges {
	// 	rev.edges[i][0] = e[1]
	// 	rev.edges[i][1] = e[0]
	// }
	// rev.runOne(t)
}
func (test coverTest) runOne(t *testing.T) {
	coverL, coverR := clip.BipartiteVertexCover(test.n, test.m, test.edges)
	if !eqSl(coverL, test.expL) {
		t.Fatalf("coverL != expL\n"+
			"\tn=%d m=%d\n"+
			"\tedges=%d %v\n"+
			"\t  expL=%v  expR=%v\n"+
			"\tcoverL=%v coverR=%v",
			test.n, test.m, len(test.edges), test.edges,
			test.expL, test.expR, coverL, coverR)
	}
	if !eqSl(coverR, test.expR) {
		t.Fatalf("coverR != expR\n"+
			"\tn=%d m=%d\n"+
			"\tedges=%d %v\n"+
			"\t  expL=%v  expR=%v\n"+
			"\tcoverL=%v coverR=%v",
			test.n, test.m, len(test.edges), test.edges,
			test.expL, test.expR, coverL, coverR)
	}
}

// eqSl returns true if `a` and `b` have the same elements in the same order.
func eqSl(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, x := range a {
		y := b[i]
		if x != y {
			return false
		}
	}
	return true
}

func _TestMatching(t *testing.T) {
	for i, test := range matchingCases {
		if showAllMatches {
			common.Log.Info("test %d ===================", i)
		}
		test.run(t)
	}
}

// matchingTest is a bipartite matching on a bipartite graph.
// The graph has `n` vertices on the first side and `m` vertices on the second size.
// `edges`[i] is the ith edge. edges[i][0] (edges[i][10]) is an index into the first (second) side
// vertices.
// `count` is the expected number of edges in the matching.
type matchingTest struct {
	n, m  int
	edges [][2]int
	count int
}

var matchingCases = []matchingTest{
	matchingTest{0, 0, [][2]int{}, 0},
	matchingTest{1, 3, [][2]int{
		[2]int{0, 0},
		[2]int{0, 1},
		[2]int{0, 2},
	}, 1},
	matchingTest{2, 2, [][2]int{
		[2]int{0, 1},
		[2]int{1, 0},
	}, 2},
	matchingTest{3, 3, [][2]int{
		[2]int{2, 0},
		[2]int{1, 1},
		[2]int{0, 2},
	}, 3},
	matchingTest{3, 2, [][2]int{
		[2]int{2, 0},
		[2]int{1, 1},
		[2]int{0, 0},
	}, 2},
	matchingTest{3, 3, [][2]int{
		[2]int{2, 0},
		[2]int{2, 2},
		[2]int{1, 1},
		[2]int{0, 0},
	}, 3},
	matchingTest{3, 3, [][2]int{
		[2]int{0, 0},
		[2]int{0, 1},
		[2]int{0, 2},
		[2]int{1, 1},
		[2]int{2, 0},
		[2]int{2, 2},
	}, 3},
	matchingTest{3, 3, [][2]int{
		[2]int{0, 1},
		[2]int{0, 2},
		[2]int{1, 0},
		[2]int{1, 2},
		[2]int{2, 0},
		[2]int{2, 1},
	}, 3},
	matchingTest{3, 3, [][2]int{
		[2]int{0, 1},
		[2]int{0, 2},
		[2]int{1, 0},
		[2]int{1, 2},
		[2]int{2, 0},
		[2]int{2, 1},
		[2]int{2, 2},
	}, 3},
	matchingTest{5, 5, [][2]int{
		[2]int{0, 0},
		[2]int{0, 1},
		[2]int{1, 0},
		[2]int{2, 1},
		[2]int{2, 2},
		[2]int{3, 2},
		[2]int{3, 3},
		[2]int{3, 4},
		[2]int{4, 4},
	}, 5},
	matchingTest{4, 4, [][2]int{
		[2]int{0, 0},
		[2]int{0, 1},
		[2]int{0, 3},
		[2]int{1, 0},
		[2]int{1, 1},
		[2]int{1, 2},
		[2]int{2, 2},
		[2]int{2, 1},
		[2]int{2, 3},
		[2]int{3, 3},
		[2]int{3, 0},
		[2]int{3, 2},
	}, 4},
}

func (test matchingTest) run(t *testing.T) {
	verifyMatching(t, test.n, test.m, test.edges, test.count)
	flipped := make([][2]int, len(test.edges))
	for i, e := range test.edges {
		flipped[i] = [2]int{e[1], e[0]}
	}
	verifyMatching(t, test.m, test.n, flipped, test.count)
}

func verifyMatching(t *testing.T, n, m int, edges [][2]int, expectedCount int) {
	matches := clip.BipartiteMatching(n, m, edges)
	if showAllMatches {
		common.Log.Info("verifyMatching: n=%d m=%d edges=%d matches=%d", n, m, len(edges), len(matches))
		for i, e := range edges {
			var mark string
			if containsEdge(matches, e) {
				mark = "***"
			}
			common.Log.Info("%4d: %v %s", i, e, mark)
		}
	}

	for i := 0; i < n; i++ {
		count := 0
		for _, v := range matches {
			if v[0] == i {
				count++
			}
		}
		if count > 1 {
			t.Fatalf("i=%d count=%d. First vertex occurred more than once.", i, count)
		}
	}
	for i := 0; i < m; i++ {
		count := 0
		for _, v := range matches {
			if v[1] == i {
				count++
			}
		}
		if count > 1 {
			t.Fatalf("i=%d count=%d. Second vertex occurred more than once.", i, count)
		}
	}
	for i, v := range matches {
		if !containsEdge(edges, v) {
			t.Fatalf("i=%d v=%v is not a valid edge", i, v)
		}
	}
	if len(matches) != expectedCount {
		t.Fatalf("match count=%d expected=%d", len(matches), expectedCount)
	}
}

func containsEdge(edges [][2]int, e [2]int) bool {
	for _, ee := range edges {
		if ee[0] == e[0] && ee[1] == e[1] {
			return true
		}
	}
	return false
}
