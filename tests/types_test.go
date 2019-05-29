package clip_test

import (
	"fmt"
	"testing"

	"github.com/peterwilliams97/clip"
)

func TestNDArray(t *testing.T) {
	testArray(t, 4, 20, 3.0)
	testArray(t, 5, 9, 10000.0/3.0)
	testArray(t, 11, 17, 0.001/17.0)
	testArray(t, 12, 12, 1.0/12.0)
	testArray(t, 11, 11, 1.0/11.0)
	testTranspose(t, 4, 20, 3.0)
	testTranspose(t, 5, 9, 10000.0/3.0)
	testTranspose(t, 11, 17, 0.001/17.0)
	testTranspose(t, 15, 15, 1.0e10)
	testTranspose(t, 1000, 10000, 1.0)
	testTranspose(t, 1000, 10000, 1.0e-10)
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
	fmt.Printf("m= %d x %d =\n%s\n%s\n", h, w, m, m.Show(2, 2))
}

func testTranspose(t *testing.T, h, w int, fac float64) {
	m := clip.CreateNDArray(h, w)
	a := clip.CreateNDArray(w, h)
	n := w * h
	count := 1
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := float64(count) / float64(n) * fac
			v = v * v
			m[y][x] = v
			a[x][y] = v
			count++
		}
	}
	mT := m.Transpose()
	if !a.Equals(mT) {
		d, err := a.Sub(mT)
		if err != nil {
			panic(err)
		}
		t.Fatalf("t!=m.Transpose\nm=%s\nm.T=%s\nt=%s\ndiff=%s", m, mT, a, d)
	}
}
