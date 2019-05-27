package liang_test

import (
	"math/rand"
	"testing"

	"github.com/peterwilliams97/clip"
)

func TestLiangeLine(t *testing.T) {
	window := clip.Rect{5, 5, 10, 10}

	for i := 0.0; i <= 5; i++ {
		test := lineTest{
			window:         window,
			line:           clip.NewLine(i, 0, 20+i, 20),
			expectedInside: true,
			expectedLine:   clip.NewLine(5+i, 5, 15, 15-i),
		}
		testLine(t, test)
	}
	for i := 6.0; i <= 10; i++ {
		test := lineTest{
			window:         window,
			line:           clip.NewLine(i, 0, 20+i, 20),
			expectedInside: false,
			expectedLine:   clip.NewLine(0, 0, 0, 0),
		}
		testLine(t, test)
	}

	for i := 0; i < 1000; i++ {
		x0, y0 := rnd(-20, 20), rnd(-20, 20)
		w, h := rnd(0, 20), rnd(0, 20)
		window := clip.Rect{x0, y0, x0 + w, y0 + h}
		lb := clip.NewLiangBarsky(window)
		for i := 0; i < 1000; i++ {
			llx := rnd(-20, 20)
			lly := rnd(-20, 20)
			urx := rnd(-20, 20)
			ury := rnd(-20, 20)
			line := clip.NewLine(llx, lly, urx, ury)
			lb.ClipLine(line)
		}
	}
}

// rnd returns a random float in the range x0..x1
func rnd(x0, x1 float64) float64 {
	return x0 + (x1-x0)*rand.Float64()
}

type lineTest struct {
	window         clip.Rect
	line           clip.Line
	expectedInside bool
	expectedLine   clip.Line
}

func testLine(t *testing.T, test lineTest) {
	l := clip.NewLiangBarsky(test.window)
	actualLine, actualInside := l.ClipLine(test.line)
	// fmt.Printf("expected=%+v actualLine=%+v\n", test.expectedLine, actualLine)
	if !actualInside {
		if test.expectedInside {
			t.Fatalf("Insidedness. test=%+v inside=%t", test, actualInside)
		}
	} else {
		if !actualLine.Equals(test.expectedLine) {
			t.Fatalf("clip. test=%+v\n\texpectedLine=%+v\n\t  actualLine=%+v",
				test, test.expectedLine, actualLine)
		}
	}
}
