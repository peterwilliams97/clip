package clip_test

import (
	"fmt"
	"testing"

	"github.com/peterwilliams97/clip"
)

func TestNDArray(t *testing.T) {
	testArray(t, 4, 2, 1.0)
	testArray(t, 10, 10, 100.0/3.0)
	testArray(t, 10, 10, 100000.0/3.0)
	testArray(t, 12, 13, 1.0/13.0)
}

func testArray(t *testing.T, h, w int, fac float64) {
	m := clip.CreateNDArray(h, w)
	count := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m[y][x] = float64(count) * fac
			count++
		}
	}
	fmt.Printf("m= %d x %d =\n%s\n", h, w, m)
}
