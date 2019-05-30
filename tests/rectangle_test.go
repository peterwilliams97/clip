package clip_test

import (
	"math"
	"sort"
	"testing"

	"github.com/peterwilliams97/clip"
	"github.com/unidoc/unidoc/common"
)

func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelInfo))
}

func TestDecomposition(t *testing.T) {

	bmp(t, 1, 1, []float64{1}, 1)

	bmp(t, 2, 3, []float64{
		1, 0, 1,
		1, 1, 1,
	}, 3)

	bmp(t, 2, 4, []float64{
		1, 1, 0, 1,
		0, 1, 1, 1,
	}, 3)

	bmp(t, 3, 5, []float64{
		1, 1, 0, 1, 1,
		0, 1, 1, 1, 0,
		1, 1, 0, 1, 1,
	}, 5)

	bmp(t, 3, 5, []float64{
		1, 1, 1, 1, 1,
		1, 0, 1, 0, 1,
		1, 1, 1, 1, 1,
	}, 5)

	bmp(t, 4, 4, []float64{
		0, 1, 0, 0,
		0, 1, 1, 1,
		0, 1, 0, 1,
		1, 1, 1, 1,
	}, 4)

	bmp(t, 4, 4, []float64{
		1, 1, 0, 0,
		0, 1, 1, 1,
		0, 1, 0, 1,
		0, 1, 1, 1,
	}, 5)

	bmp(t, 6, 6, []float64{
		1, 1, 0, 0, 0, 1,
		0, 1, 1, 1, 0, 1,
		0, 1, 0, 1, 0, 1,
		0, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 1, 1,
		1, 1, 1, 0, 0, 1,
	}, 0)

	bmp(t, 3, 6, []float64{
		0, 1, 0, 0, 0, 1,
		1, 1, 1, 1, 0, 0,
		1, 0, 0, 1, 0, 0,
	}, 5)

	bmp(t, 2, 5, []float64{
		0, 1, 0, 1, 0,
		1, 1, 0, 1, 1,
	}, 4)

	bmp(t, 20, 20, []float64{
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 1, 1, 1, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0,
		0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 0,
		0, 0, 1, 0, 0, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1, 0, 0, 1, 1, 0,
		0, 0, 1, 0, 0, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1, 0, 0, 1, 1, 0,
		0, 1, 1, 0, 0, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1, 0, 0, 1, 1, 0,
		0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0,
		1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0,
		1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0,
		1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0,
		0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 1, 1, 1, 1, 0,
		0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0,
		0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0,
		0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0,
		0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0,
		1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0, 0, 0,
		1, 1, 1, 1, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0,
		1, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}, 29)

	//   // *-*
	//   // | |
	//   // * |
	//   // | |
	//   // *-*
	//   test([
	//       [[0,0], [0,1], [0,2], [1,2], [1,0]]
	//     ], true, 1)

	//   //   *-*
	//   //   | |
	//   // *-*-*
	//   // | |
	//   // *-*
	//   test([
	//       [[0,0], [0,1], [1,1], [1,0]],
	//       [[1,1], [1,2], [2,2], [2,1]]
	//     ], true, 2)

	//   //   *---*
	//   //   |   |
	//   // *-*-* |
	//   // | | | |
	//   // | *-* |
	//   // |     |
	//   // *-----*
	//   test([
	//     [[0,0], [3,0], [3,3], [1,3], [1,2], [2,2], [2,1], [1,1], [1,2], [0,2]]
	//     ], false, 4)

	//   //   *-----*
	//   //   |     |
	//   // *-*     |
	//   // |       |
	//   // | *-*   |
	//   // | | |   |
	//   // | *-*   |
	//   // |       |
	//   // *-------*
	//   test([
	//       [[1,1], [1,2], [2,2], [2,1]],
	//       [[0,0], [4,0], [4,4], [1,4], [1,3], [0,3]]
	//     ], false, 4)

	//   //   *-*
	//   //   | |
	//   // *-* *-*
	//   // |     |
	//   // *-* *-*
	//   //   | |
	//   //   *-*
	//   var plus = [
	//     [1, 1],
	//     [0, 1],
	//     [0, 2],
	//     [1, 2],
	//     [1, 3],
	//     [2, 3],
	//     [2, 2],
	//     [3, 2],
	//     [3, 1],
	//     [2, 1],
	//     [2, 0],
	//     [1, 0]
	//   ]
	//   test([plus], true, 3)

	//   // *---*
	//   // |   |
	//   // *-* *-*
	//   //   |   |
	//   //   *---*
	//   var zigZag = [
	//     [1,1],
	//     [0,1],
	//     [0,2],
	//     [2,2],
	//     [2,1],
	//     [3,1],
	//     [3,0],
	//     [1,0]
	//   ]
	//   test([zigZag], true, 2)

	//   //    *-*
	//   //    | |
	//   //  *-* *-*
	//   //  |     |
	//   //  *-----*
	//   var bump = [
	//     [0,0],
	//     [0,1],
	//     [1,1],
	//     [1,2],
	//     [2,2],
	//     [2,1],
	//     [3,1],
	//     [3,0]
	//   ]
	//   test([bump], true, 2)

	//   //   *-*
	//   //   | |
	//   // *-* |
	//   // |   |
	//   // *---*
	//   var bracket = [
	//     [0,0],
	//     [0,1],
	//     [1,1],
	//     [1,2],
	//     [2,2],
	//     [2,0]
	//   ]
	//   test([bracket], true, 2)

}

func verifyDecomp(t *testing.T, paths []clip.Path, ccw bool, expected int) {
	rectangles := clip.DecomposeRegion(paths, ccw)

	overlaps0 := boxOverlap(rectangles)
	var overlaps []Overlap
	for _, o := range overlaps0 {
		a := rectangles[o.i1]
		b := rectangles[o.i2]
		if math.Min(a.Urx, b.Urx) > math.Max(a.Llx, b.Llx) && math.Min(a.Ury, b.Ury) > math.Max(a.Lly, b.Lly) {
			overlaps = append(overlaps, o)
		}
	}
	if len(overlaps) != expected {
		t.Fatalf("overlaps=%d expected=%d", len(overlaps), expected)
	}

	// t.same(boxOverlap(rectangles).filter(function(x) {
	//   var a = rectangles[x[0]]
	//   var b = rectangles[x[1]]
	//   var x = Math.math.Min(a[1][0], b[1][0]) - Math.math.Max(a[0][0], b[0][0])
	//   if(x <= 0) {
	//     return false
	//   }
	//   var y = Math.math.Min(a[1][1], b[1][1]) - Math.math.Max(a[0][1], b[0][1])
	//   if(y <= 0) {
	//     return false
	//   }
	//   return true
	// }), [], "non-overlap")

	//Compute area for polygon and check each path is covered by an edge of
	area := 0.0
	for _, pathSet := range paths {
		for j, a := range pathSet {
			b := pathSet[(j+1)%len(pathSet)]
			if a.Y == b.Y {
				area += (b.X - a.X) * a.Y
			}
		}
	}
	if !ccw {
		area = -area
	}

	//Compute area for boxes
	boxarea := 0.0
	for _, r := range rectangles {
		boxarea += r.Area()
		if !r.Valid() {
			t.Fatalf("Bad rectangle %+v", r)
		}
	}
	if boxarea != area {
		t.Fatalf("box area wrong %g expected %g", boxarea, area)
	}

	//TODO: Add more tests here?
}

func test(t *testing.T, paths []clip.Path, ccw bool, expected int) {
	// Check all 4 orientations
	for sx := 1; sx >= -1; sx -= 2 {
		for sy := 1; sy >= -1; sy -= 2 {
			npaths := make([]clip.Path, len(paths))
			for i, path := range paths {
				out := make(clip.Path, len(path))
				for j, p := range path {
					out[j] = clip.Point{X: float64(sx) * p.X, Y: float64(sy) * p.Y}
				}
				npaths[i] = out
			}

			var nccw bool
			if sx*sy < 0 {
				nccw = !ccw
			} else {
				nccw = ccw
			}
			verifyDecomp(t, npaths, nccw, expected)
		}
	}
}

// Test with a bitmap image
func bmp(t *testing.T, h, w int, img []float64, expected int) {
	m, err := clip.SliceToNDArray(h, w, img)
	if err != nil {
		panic(err)
	}

	paths := clip.GetContours(m, true)
	test(t, paths, false, expected)
}

// Michael Doescher
// October 10, 2013
// This program reports overlapping boxes
// Input = an array of box coordinates.  Each box is defined as an array of points.  The points
//         represent the lower left and upper right corner.
// Output = A two dimensional array.  Each row contains two values indicating the index value of the
//          boxes from the input that overlap
func boxOverlap(boxes []clip.Rect) []Overlap {
	events := generateEvents(boxes)
	sort.Slice(events, func(i, j int) bool {
		a, b := events[i], events[j]
		if a.x < b.x {
			return false
		}
		if a.x > b.x {
			return true
		}
		if a.x == b.x && a.typ == "add" && b.typ == "remove" {
			return false // adding before removing allows for boxes that overlap
		}
		if a.x == b.x && a.typ == "remove" && b.typ == "add" {
			return true // only on the edge to count as overlapping.
		}
		return false
	})
	return generateOvelapList(boxes, events)
}

// module.exports = function(boxes) {
// 	// if (!isInputOk(boxes)) {return null;}
// 	var events = generateEvents(boxes);
// 	events.sort(compare);
// 	overlaps = new Array();
// 	var overlaps = generateOvelapList(boxes, events, overlaps);
// 	return overlaps;
// }

/*
[
	[[0, 0], [1, 1]], 			//box 1
	[[0.5, 0.5], [10, 10]]		//box 2
]
*/

type Overlap struct {
	i1, i2 int // Indexes in boxes array of overlapping pair of boxes
}

type Event struct {
	x     float64
	typ   string
	index int
}

func generateEvents(boxes []clip.Rect) []Event {
	var leftEvents, rightEvents []Event

	for i, b := range boxes { // traverse the list of boxes
		leftx := math.Min(b.Llx, b.Urx)
		rightx := math.Max(b.Llx, b.Urx)

		eventl := Event{
			x:     leftx,
			typ:   "add",
			index: i,
		}
		leftEvents = append(leftEvents, eventl)
		eventr := Event{
			x:     rightx,
			typ:   "remove",
			index: i,
		}
		rightEvents = append(rightEvents, eventr)
	}
	events := make([]Event, len(leftEvents)+len(rightEvents))
	for i, e := range leftEvents {
		events[len(leftEvents)-1-i] = e
	}
	for i, e := range rightEvents {
		events[len(leftEvents)+i] = e
	}
	return events
}

// func compare(a,b) {
//   if (a.x < b.x)
//      return -1;
//   if (a.x > b.x)
//     return 1;
//   if (a.x == b.x && a.type == "add" && b.type == "remove") return -1;	// adding before removing allows for boxes that overlap
//   if (a.x == b.x && a.type == "remove" && b.type == "add") return 1; 	// only on the edge to count as overlapping.
//   return 0;
// }

func generateOvelapList(boxes []clip.Rect, events []Event) []Overlap {
	var Q []int            // a list of indices into the boxes array of boxes that intersect the sweeping plane
	var overlaps []Overlap // pairs of boxes that overlap (indices into the boxes array

	for _, e := range events {
		if e.typ == "add" {
			overlaps = findOverlap(Q, e.index, overlaps, boxes)
			Q = append(Q, e.index)
		}
		if e.typ == "remove" {
			ind := -1
			for i, ee := range events {
				if Q[i] == ee.index {
					ind = i
					break
				}
			}
			if ind < 0 {
				panic("cant happen")
			}
			// ind := Q.indexOf(events[i].index)
			// Q.splice(ind, 1);
			Q = append(Q[:ind], Q[ind:]...)
		}
	}
	return overlaps
}

func sliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func findOverlap(Q []int, box int, overlaps []Overlap, boxes []clip.Rect) []Overlap {
	if len(Q) == 0 {
		return overlaps
	}
	for _, q := range Q {
		b := boxes[q]
		eb := boxes[box]
		y1 := math.Min(b.Lly, b.Ury)
		y2 := math.Max(b.Lly, b.Ury)
		ey1 := math.Min(eb.Lly, eb.Ury)
		ey2 := math.Max(eb.Lly, eb.Ury)

		add := (ey1 >= y1 && ey1 <= y2) ||
			(ey2 >= y1 && ey2 <= y2) ||
			(ey1 < y1 && ey2 > y2)

		if add {
			o := createOverlap(q, box)
			overlaps = append(overlaps, o)
		}
	}
	return overlaps
}

func createOverlap(a, b int) Overlap {
	if a > b {
		a, b = b, a
	}
	return Overlap{a, b}
}
