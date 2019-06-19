package clip_test

import (
	"testing"

	"github.com/peterwilliams97/clip"
	"github.com/unidoc/unipdf/common"
)

func TestRect(t *testing.T) {
	r := clip.Rect{Urx: 10, Ury: 20}
	area := r.Area()
	if area != 200 {
		t.Fatalf("Incorrect area: r=%+v area=%g", r, area)
	}
}

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

func TestChords(t *testing.T) {
	testOpposite(t, p(2, 2), p(2, 1), p(0, 2), p(3, 2))
	testOpposite(t, p(1, 2), p(1, 1), p(0, 2), p(3, 2))
	testOpposite(t, p(2, 2), p(1, 2), p(2, 0), p(2, 3))
	testOpposite(t, p(2, 1), p(1, 1), p(2, 0), p(2, 3))
	testIntersects(t, true, p(1, 1), p(2, 0), p(2, 3), p(2, 0), p(2, 3))
	testIntersects(t, true, p(1, 1), p(4, 0), p(4, 3), p(2, 0), p(2, 3))
	testIntersects(t, false, p(1, 1), p(4, 0), p(4, 3), p(5, 0), p(5, 3))
	testIntersects(t, false, p(1, 1), p(4, 0), p(4, 3), p(2, 4), p(2, 6))
}

func p(x, y float64) clip.Point {
	return clip.Point{x, y}
}

func testOpposite(t *testing.T, e, v, s0, s1 clip.Point) {
	c := clip.NewChord(v, clip.Line{s0, s1})
	o := c.OtherEnd()
	common.Log.Debug("c=%v -> opposite=%v", c, o)
	if !o.Equals(e) {
		t.Fatalf("got %v expected %v\n\tc=%v", o, e, c)
	}
}

func testIntersects(t *testing.T, e bool, v, s0, s1, l0, l1 clip.Point) {
	c := clip.NewChord(v, clip.Line{s0, s1})
	l := clip.NewSide(&clip.Vertex{Point: l0}, &clip.Vertex{Point: l1})
	i := c.Intersects(l)
	common.Log.Debug("c=%v l=%v-> intersects=%t", c, l, i)
	if i != e {
		t.Fatalf("got %t expected %t\n\tc=%v\n\tl=%v", i, e, c, l)
	}
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
	common.Log.Debug("m= %d x %d =\n%s\n%s", h, w, m, m.Show(2, 2))
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
