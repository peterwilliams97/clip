package clip

import (
	"fmt"
	"strings"
)

// Point is a 2-d point.
type Point struct {
	X, Y float64
}

// Cpt returns p.Y if `vertical` is true else p.X.
func (p Point) Cpt(vertical bool) float64 {
	if vertical {
		return p.Y
	}
	return p.X
}

func (p Point) add(d Point) Point {
	return Point{p.X + d.X, p.Y + d.Y}
}

func (p Point) sub(d Point) Point {
	return Point{p.X - d.X, p.Y - d.Y}
}

func (p Point) mul(g float64) Point {
	return Point{p.X * g, p.Y * g}
}

func (p Point) isZero() bool {
	return isZero(p.X) && isZero(p.Y)
}

// Equals returns true if `p` and `d` are in the same location.
func (p Point) Equals(d Point) bool {
	return p.sub(d).isZero()
}

// Line is a straight line.
type Line struct {
	A, B Point
}

// Path is a path that is not necessarily closed.
type Path []Point

// Position returns the parametrized point.
// p = a ∙ (1 - t) + b ∙ t = a + (b - a) ∙ t
func (l Line) Position(t float64) Point {
	a, b := l.A, l.B
	d := b.sub(a)
	return a.add(d.mul(t))
}

// NewLine returns the line from (ax, ay) to (bx, by).
func NewLine(ax, ay, bx, by float64) Line {
	return Line{Point{ax, ay}, Point{bx, by}}
}

// Equals returns true if `l` and `d` are in the same location.
func (l Line) Equals(d Line) bool {
	return l.A.Equals(d.A) && l.B.Equals(d.B)
}

// Rect is a rectangle.
type Rect struct {
	Llx, Lly, Urx, Ury float64
}

func (r Rect) Area() float64 {
	if !r.Valid() {
		panic(fmt.Errorf("invalid rectangle r=%+v", r))
	}
	return (r.Urx - r.Llx) * (r.Ury - r.Lly)
}

func (r Rect) Valid() bool {
	return r.Urx >= r.Llx && r.Ury >= r.Lly
}

// NDArray is like a Python 2-d ndarray
type NDArray [][]float64

func CreateNDArray(h, w int) NDArray {
	backing := make([]float64, h*w)
	m := make([][]float64, h)
	for y := 0; y < h; y++ {
		m[y] = backing[y*w : (y+1)*w]
	}
	return m
}

func SliceToNDArray(h, w int, a []float64) (NDArray, error) {
	if len(a) != w*h {
		return nil, fmt.Errorf("len(a)=%d h=%d w=%d", len(a), w, h)
	}
	backing := make([]float64, h*w)
	copy(backing, a)
	m := make([][]float64, h)
	for y := 0; y < h; y++ {
		m[y] = backing[y*w : (y+1)*w]
	}
	return m, nil
}

const (
	// Maximum number of columns and rows to display
	maxRows    = 10
	maxColumns = 10
)

func (m NDArray) String() string {
	return m.Show(10, 10)
}

// Show returns a string representation of `m`.
// `maxRows`: Maximum number of rows to display.
// `maxColumns`: Maximum number of columns to display.
func (m NDArray) Show(maxRows, maxColumns int) string {
	h := len(m)
	if h == 0 {
		return "[]"
	}
	w := len(m[0])
	if w == 0 {
		return "[]"
	}

	vals := make([]string, h*w)
	n := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			vals[x+w*y] = fmt.Sprintf("%.3g", m[y][x])
			if len(vals[x+w*y]) > n {
				n = len(vals[x+w*y])
			}
		}
	}
	format := fmt.Sprintf("%%%ds", n)

	var sb strings.Builder
	fmt.Fprint(&sb, "[")
	skippedY := false
	for y := 0; y < h; y++ {
		if h > maxRows && maxRows <= y*2 && y*2 < h*2-maxRows {
			if !skippedY {
				fmt.Fprintf(&sb, "  "+format+"\n", "...")
				skippedY = true
			}
			continue
		}
		if y > 0 {
			fmt.Fprint(&sb, " ")
		}
		fmt.Fprint(&sb, "[")
		skippedX := false
		for x := 0; x < w-1; x++ {
			if w > maxColumns && maxColumns <= x*2 && x*2 < w*2-maxColumns {
				if !skippedX {
					fmt.Fprintf(&sb, format+" ", "...")
					skippedX = true
				}
				continue
			}
			fmt.Fprintf(&sb, format+" ", vals[x+w*y])
		}
		fmt.Fprintf(&sb, format+"]", vals[w*y+w-1])
		if y < h-1 {
			fmt.Fprintln(&sb, "")
		}

	}
	fmt.Fprint(&sb, "]")
	return sb.String()
}

func (m NDArray) Shape() (h, w int) {
	if len(m) == 0 {
		return 0, 0
	}
	return len(m), len(m[0])
}

func (m NDArray) Transpose() NDArray {
	h, w := m.Shape()
	t := CreateNDArray(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			t[x][y] = m[y][x]
		}
	}
	return t
}

func (m NDArray) Equals(d NDArray) bool {
	h, w := m.Shape()
	hd, wd := d.Shape()
	if hd != h || wd != w {
		return false
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if d[y][x] != m[y][x] {
				return false
			}
		}
	}
	return true
}

func (m NDArray) Sub(d NDArray) (NDArray, error) {
	h, w := m.Shape()
	hd, wd := d.Shape()
	if hd != h || wd != w {
		return nil, fmt.Errorf("Dimension mismatch %dx%d - %dx%d", h, w, hd, wd)
	}
	t := CreateNDArray(h, w)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			t[y][x] = m[y][x] - d[y][x]
		}
	}
	return t, nil
}
