package clip_test

import (
	"fmt"
	"testing"

	"github.com/peterwilliams97/clip"
)

func TestContour(t *testing.T) {
	for _, test := range contourTests {
		testContour(t, test)
	}
}

var contourTests = []contourCase{
	contourCase{
		h: 3, w: 5,
		img: []float64{
			0, 0, 0, 0, 0,
			1, 0, 0, 0, 0,
			1, 1, 1, 1, 1,
		},
		expected: []clip.Path{
			clip.Path{
				clip.Point{X: 0, Y: 1},
				clip.Point{X: 0, Y: 3},
				clip.Point{X: 5, Y: 3},
				clip.Point{X: 5, Y: 2},
				clip.Point{X: 1, Y: 2},
				clip.Point{X: 1, Y: 1},
			},
		},
		/*
			   0   0   0   0   0
			0,1-1,1
			 | 1 | 0   0   0   0
			 |  1,2-------------5,2
			 | 1   1   1   1   1 |
			0,3-----------------5,3
		*/
	},
	contourCase{
		h: 3, w: 5,
		img: []float64{
			1, 1, 1, 0, 0,
			1, 0, 1, 1, 1,
			1, 1, 1, 1, 1,
		},
		expected: []clip.Path{
			clip.Path{
				clip.Point{0, 0},
				clip.Point{0, 3},
				clip.Point{5, 3},
				clip.Point{5, 1},
				clip.Point{3, 1},
				clip.Point{3, 0},
			},
			clip.Path{
				clip.Point{2, 1},
				clip.Point{2, 2},
				clip.Point{1, 2},
				clip.Point{1, 1},
			},
		},
		/*
		   0,0---------3,0
		    | 1   1   1 | 0   0
		    |          3,1-----5,1
		    | 1   0   1   1   1 |
		    |                   |
		    | 1   1   1   1   1 |
		   0,3-----------------5,3

		   	  1   1   1   0   0
		       1,1-2,1
		      1 | 0 | 1   1   1
		        1,2-2,2
		      1   1   1   1   1
		*/
	},
}

type contourCase struct {
	w, h     int
	img      []float64
	expected []clip.Path
}

func testContour(t *testing.T, test contourCase) {
	array, err := clip.SliceToNDArray(test.h, test.w, test.img)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	clockwise := false
	poly := clip.GetContours(array, clockwise)

	fmt.Printf("array=\n%s\n", array)
	fmt.Printf("expected=%+v\n", test.expected)
	fmt.Printf("     got=%+v\n", poly)

	if !samePoly(test.expected, poly) {
		t.Fatalf("Incorrect results:\n\tgot=%+v\n\texpected=%+v", poly, test.expected)
	}
}

func samePoly(poly0, poly []clip.Path) bool {
	if len(poly0) != len(poly) {
		return false
	}
	for i, path0 := range poly0 {
		path := poly[i]
		if !samePath(path0, path) {
			return false
		}
	}
	return true
}
func samePath(path0, path clip.Path) bool {
	if len(path0) != len(path) {
		return false
	}
	for i, p0 := range path0 {
		p := path[i]
		if !p.Equals(p0) {
			return false
		}
	}
	return true
}