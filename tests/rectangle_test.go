package clip_test

import (
	"math"
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

// bmp tests  with a bitmap image specified by
// `h`: height in pixels
// `w`: width in pixels
// `img`: pixels in image
func bmp(t *testing.T, h, w int, img []float64, expected int) {
	common.Log.Info("bmp: h=%d w=%d img=%+v expected=%d", h, w, img, expected)
	m, err := clip.SliceToNDArray(h, w, img)
	if err != nil {
		t.Fatalf("err=%v", err)
	}

	common.Log.Info("bmp: m=%s", m)
	paths := clip.GetContours(m, false)
	common.Log.Info("bmp: paths=%+v", paths)
	test(t, paths, false, expected)
	panic("Done")
}

func test(t *testing.T, paths []clip.Path, clockwise bool, expected int) {
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

			var nclockwise bool
			if sx*sy < 0 {
				nclockwise = !clockwise
			} else {
				nclockwise = clockwise
			}
			verifyDecomp(t, npaths, nclockwise, expected)
		}
	}
}

// verifyDecomp tests DecomposeRegion on polygon `paths`.
// `clockwise` is true if
// `expected` is the expected number of overlaps.
func verifyDecomp(t *testing.T, paths []clip.Path, ccw bool, expected int) {
	clockwise := !ccw
	rectangles := clip.DecomposeRegion(paths, clockwise)
	common.Log.Info("verifyDecomp:\n\t paths=%d %+v\n\t clockwise=%t expected=%d\n\t rectangles=%d %+v",
		len(paths), paths, clockwise, expected, len(rectangles), rectangles)

	if len(rectangles) != expected {
		t.Fatalf("overlaps=%d expected=%d", len(rectangles), expected)
	}

	overlaps0 := clip.BoxOverlap(rectangles)
	var overlaps []clip.Overlap
	for _, o := range overlaps0 {
		a := rectangles[o.I1]
		b := rectangles[o.I2]
		if math.Min(a.Urx, b.Urx) > math.Max(a.Llx, b.Llx) && math.Min(a.Ury, b.Ury) > math.Max(a.Lly, b.Lly) {
			overlaps = append(overlaps, o)
		}
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

	// Compute area for polygon and check each path is covered by an edge of
	area := 0.0
	for _, pathSet := range paths {
		for j, a := range pathSet {
			b := pathSet[(j+1)%len(pathSet)]
			if a.Y == b.Y {
				area += (b.X - a.X) * a.Y
			}
		}
	}
	if !clockwise {
		area = -area
	}

	// Compute area for boxes
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

func sliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}
