package clip

import (
	"fmt"

	"github.com/biogo/store/interval"
	"github.com/unidoc/unipdf/v3/common"
)

// Generic intervals
type Int float64

func (c Int) Compare(b interval.Comparable) int {
	d := c - b.(Int)
	if isZero(float64(d)) {
		return 0
	}
	if d > 0 {
		return 1
	}
	return -1
}

// Interval is an interval over points in either the horizontal or vertical direction.
// Intervals can be made for any type that fullfills the `Rectilinear` interface.
// x0 and x1 are the lower and upper ends of the interval
// y is the coordinate in the other axis than x0, x1
// If vertical is false, x0, x1 are on the X-axis and yY is on the Y-axis.
// If vertical is true, x0, x1 are on the Y-axis and y is on the X-axis.
// r is the original object.
type Interval struct {
	x0, x1, y float64
	vertical  bool
	id        uintptr
	r         Rectilinear
}

// idCounter is used to give Interval.id unique values for all intervals. Interval trees won't work
// without it.
var idCounter uintptr

// IntervalTree allows us to add methods on interval.Tree. Could do without it.
type IntervalTree interval.Tree

// NewIntv is a hack for testing.
func NewIntv(x0, x1 float64) *Interval {
	s := &Side{x0: x0, x1: x1}
	idCounter = (idCounter + 10) % 10
	iv := rectilinearToInterval(s)
	return &iv

}

// Range returns the ends of the interval
func (i Interval) Range() (float64, float64) {
	return i.x0, i.x1
}

func newInterval(v0, v1 *Vertex, vertical bool) Interval {
	s := NewSide(v0, v1)
	return rectilinearToInterval(s)
}

func CreateIntervalTreeSides(sides []*Side, name string) *IntervalTree {
	rects := make([]Rectilinear, len(sides))
	for i, s := range sides {
		rects[i] = s
	}
	return createIntervalTree(rects, name)
}

func CreateIntervalTreeChords(chords []*Chord, name string) *IntervalTree {
	rects := make([]Rectilinear, len(chords))
	for i, c := range chords {
		rects[i] = c
	}
	return createIntervalTree(rects, name)
}

func CreateIntervalTreeInterval(intervals []*Interval, name string) *IntervalTree {
	rects := make([]Rectilinear, len(intervals))
	for i, iv := range intervals {
		rects[i] = iv
	}
	return createIntervalTree(rects, name)
}

func createIntervalTree(rects []Rectilinear, name string) *IntervalTree {
	// common.Log.Info("createIntervalTree: %d %q", len(rects), name)
	// for i, r := range rects {
	// 	common.Log.Info("%d: %s", i, rectilinearToInterval(r))
	// }
	if len(rects) == 0 {
		panic("createIntervalTree: empty")
	}
	tree := &IntervalTree{}
	for _, r := range rects {
		tree.Insert(r)
	}
	// This is critical!
	t := (*interval.Tree)(tree)
	t.AdjustRanges()
	tree.Validate()
	return tree
}

func (tree *IntervalTree) getIntervals() []*Interval {
	var intervals []*Interval
	t := (*interval.Tree)(tree)
	t.Do(func(e interval.Interface) bool {
		iv := e.(Interval)
		intervals = append(intervals, &iv)
		return false
	})
	return intervals
}

func (tree *IntervalTree) Validate() {
	intervals := tree.getIntervals()
	ValidateIntervals(intervals)
}

func ValidateIntervals(intervals []*Interval) {
	counts := map[string]int{}

	for i, v := range intervals {
		sig := rectString(v)
		counts[sig]++
		if counts[sig] > 1 {
			common.Log.Error("-------------counts---------------")
			for k, v := range counts {
				common.Log.Error("%s: %d", k, v)
			}
			common.Log.Error("-------------&&&---------------")
			for j, w := range intervals[:i+1] {
				common.Log.Error("%4d: %s", j, w)
			}
			panic(fmt.Errorf("Duplicate interval i=%d v=%v", i, v))
		}
	}
}

func rectilinearToInterval(r Rectilinear) Interval {
	idCounter++ // !@#$ Critical for passing tests
	x0, x1, y, vertical := r.X0X1YVert()
	return Interval{
		x0:       x0,
		x1:       x1,
		y:        y,
		vertical: vertical,
		r:        r,
		id:       idCounter, // counter has effect
	}
}

func (i *Interval) X0X1YVert() (x0, x1, y float64, vertical bool) {
	// common.Log.Info("Interval.X0X1YVert: %v", i)
	return i.x0, i.x1, i.y, i.vertical
}

func (tree *IntervalTree) Insert(r Rectilinear) {
	tree.Validate()
	v := rectilinearToInterval(r)
	t := (*interval.Tree)(tree)
	if err := t.Insert(v, true); err != nil {
		panic(fmt.Errorf("IntervalTree.Insert s=%v err=%v", r, err))
	}
	// common.Log.Info("Insert: v=%v", v)
	tree.Validate()
	common.Log.Debug("treeInsert: %v r=%#v", tree, r)
}

func (tree *IntervalTree) Delete(r Rectilinear) {
	i := rectilinearToInterval(r)
	t := (*interval.Tree)(tree)
	t.Delete(i, true)
	common.Log.Debug("treeDelete: %v r=%#v", tree, r)
}

func (tree *IntervalTree) QueryPoint(x float64, cb func(Rectilinear) bool) bool {
	var matched bool
	common.Log.Debug("QueryPoint: x=%g tree=%+v", x, tree)
	t := (*interval.Tree)(tree)
	q := query1d(x)
	ok := t.DoMatching(func(e interval.Interface) bool {
		iv := e.(Interval)
		matched := cb(iv.r)
		common.Log.Debug(" iv=%#v matched=%t", iv.r, matched)
		return matched
	}, q)
	if matched != ok {
		panic("QueryPoint")
	}
	common.Log.Debug("QueryPoint: matched=%t", matched)
	return matched
}

type query1d float64

func (q query1d) Overlap(b interval.Range) bool {
	var x0, x1 float64
	switch bc := b.(type) {
	case Interval:
		x0, x1 = bc.x0, bc.x1
	case *Mutable:
		x0, x1 = float64(bc.x0), float64(bc.x1)
	default:
		panic("unknown type")
	}
	x := float64(q)

	return x0 <= x && x <= x1
}

func (i Interval) Overlap(b interval.Range) bool {
	var x0, x1 float64
	switch bc := b.(type) {
	case Interval:
		x0, x1 = bc.x0, bc.x1
	case *Mutable:
		x0, x1 = float64(bc.x0), float64(bc.x1)
	default:
		panic("unknown type")
	}

	// Full-open interval indexing. !@#$
	return i.x1 >= x0 && i.x0 <= x1
}
func (i Interval) ID() uintptr                { return i.id }
func (i Interval) Start() interval.Comparable { return Int(i.x0) }
func (i Interval) End() interval.Comparable   { return Int(i.x1) }
func (i Interval) NewMutable() interval.Mutable {
	return &Mutable{
		x0: Int(i.x0),
		x1: Int(i.x1),
		y:  Int(i.y),
		r:  i.r,
		id: i.id,
	}
}
func (i Interval) String() string {
	direct := "horizontal"
	if i.vertical {
		direct = "vertical"
	}
	return fmt.Sprintf("INTERVAL{[%.1f,%.1f]%.1f(%s)\n\t%#v#%d}",
		i.x0, i.x1, i.y, direct, i.r, i.id)
}

type Mutable struct {
	x0, x1, y Int
	vertical  bool
	id        uintptr
	r         Rectilinear
}

func (m *Mutable) Start() interval.Comparable { return m.x0 }
func (m *Mutable) End() interval.Comparable   { return m.x1 }
func (m *Mutable) SetStart(c interval.Comparable) {
	// common.Log.Info("Mutable.SetStart %g->%g", m.x0, float64(c.(Int)))
	// if isZero(m.x0+3.55642) && isZero(float64(c.(Int))+8.3346) {
	// 	panic("SetStart")
	// }
	m.x0 = c.(Int)

}
func (m *Mutable) SetEnd(c interval.Comparable) {
	// common.Log.Info("Mutable.SetEnd %g->%g", m.x1, float64(c.(Int)))
	m.x1 = c.(Int)
}

// func (t *interval.Tree) queryPoint(x float64, f func(h *Side)) {
// }
// func (t *interval.Tree) insert(h *Side) {
// }
// func (t *interval.Tree) remove(h *Side) {
// }
