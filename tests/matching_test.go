package clip_test

import (
	"testing"

	"github.com/peterwilliams97/clip"
	"github.com/unidoc/unidoc/common"
)

func init() {
	level := common.LogLevelInfo
	common.SetLogger(common.NewConsoleLogger(level))
	common.Log.Info("level=%#v", level)
	common.Log.Debug("level=%#v", level)
}

func TestMatching(t *testing.T) {
	for _, test := range matchingCases {
		test.run(t)
	}
}

type matchingTest struct {
	n, m  int
	edges [][2]int
	count int
}

var matchingCases = []matchingTest{
	matchingTest{1, 3, [][2]int{
		[2]int{0, 0},
		[2]int{0, 1},
		[2]int{0, 2},
	}, 1},
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
	common.Log.Info("verifyMatching: n=%d m=%d edges=%d", n, m, len(edges))
	for i, e := range edges {
		var mark string
		if containsEdge(matches, e) {
			mark = "***"
		}
		common.Log.Info("%4d: %v %s", i, e, mark)
	}
	common.Log.Info("verifyMatching:matches=%d", len(matches))
	for i, e := range matches {
		common.Log.Info("%4d: %v", i, e)
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
