package clip_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/peterwilliams97/clip"
	"github.com/unidoc/unipdf/v3/common"
)

func TestOverlap(t *testing.T) {
	for i, test := range overlapCases {
		common.Log.Debug("test %d -----------------------------\n%s", i, test)
		testOverlap(t, test)
	}
}

func testOverlap(t *testing.T, test overlapTest) {
	overlap := clip.BoxOverlap(test.boxes)
	if !sameOverlap(test.expected, overlap) {
		t.Fatalf("Wrong overlap\n\tboxes=%d %v\n\texpected=%d %v\n\t     got=%d %v",
			len(test.boxes), test.boxes, len(test.expected), test.expected, len(overlap), overlap)
	}
}

type overlapTest struct {
	boxes    []clip.Rect
	expected []clip.Overlap
}

func (test overlapTest) String() string {
	var sb strings.Builder
	for i, b := range test.boxes {
		fmt.Fprintf(&sb, "%6d: %.2f\n", i, b)
	}
	fmt.Fprintf(&sb, "boxes=%d %v\nexpected=%d %v",
		len(test.boxes), test.boxes, len(test.expected), test.expected)
	return sb.String()
}

func sameOverlap(overlap0, overlap1 []clip.Overlap) bool {
	if len(overlap0) != len(overlap1) {
		return false
	}
	for i, o0 := range overlap0 {
		o1 := overlap1[i]
		if !o0.Equals(o1) {
			return false
		}
	}
	return true
}

var overlapCases = []overlapTest{
	overlapTest{
		boxes: []clip.Rect{
			clip.Rect{0, 0, 1, 1},
		},
		expected: []clip.Overlap{},
	},
	overlapTest{
		boxes: []clip.Rect{
			clip.Rect{0, 0, 1, 1},
			clip.Rect{0.5, 0.5, 1.5, 1.5},
		},
		expected: []clip.Overlap{
			clip.Overlap{0, 1},
		},
	},
	overlapTest{
		boxes: []clip.Rect{
			clip.Rect{0, 0, 1, 1},
			clip.Rect{0.5, 0.5, 1.5, 1.5},
			clip.Rect{1.1, 1, 2, 2},
		},
		expected: []clip.Overlap{
			clip.Overlap{0, 1},
			clip.Overlap{1, 2},
		},
	},
	overlapTest{
		boxes: []clip.Rect{
			clip.Rect{0, 0, 1, 1},
			clip.Rect{0.5, 0.5, 1.5, 1.5},
			clip.Rect{1.6, 1, 2, 2},
		},
		expected: []clip.Overlap{
			clip.Overlap{0, 1},
		},
	},
	overlapTest{
		boxes: []clip.Rect{
			clip.Rect{0, 0, 1, 1},
			clip.Rect{0, 0.5, 1, 1.5},
			clip.Rect{0, 1.1, 1, 2},
		},
		expected: []clip.Overlap{
			clip.Overlap{0, 1},
			clip.Overlap{1, 2},
		},
	},
	overlapTest{
		boxes: []clip.Rect{
			clip.Rect{0, 0, 1, 1},
			clip.Rect{0.5, 0.5, 0.75, 0.75},
			clip.Rect{2, 0, 3, 1},
			clip.Rect{2.5, -10, 2.75, 10},
		},
		expected: []clip.Overlap{
			clip.Overlap{0, 1},
			clip.Overlap{2, 3},
		},
	},
	overlapTest{
		boxes: []clip.Rect{
			clip.Rect{0, 0, 1, 1},
			clip.Rect{0.5, 0.5, 0.75, 0.75},
			clip.Rect{2, 0, 3, 1},
			clip.Rect{2.5, -10, 2.75, 10},
			clip.Rect{5, 0, 4, 1},
			clip.Rect{4.5, 0.5, 5.5, 1.5},
			clip.Rect{0, 2, 1, 11},
			clip.Rect{3, 0, 4, 1},
		},
		expected: []clip.Overlap{
			clip.Overlap{0, 1},
			clip.Overlap{2, 3},
			clip.Overlap{2, 7},
			clip.Overlap{4, 5},
			clip.Overlap{4, 7},
		},
	},
}
