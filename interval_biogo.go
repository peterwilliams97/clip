package clip

import (
	"fmt"
	"math"

	"github.com/biogo/store/interval"
	"github.com/unidoc/unipdf/common"
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
type Interval struct {
	*Side
	id uintptr
	// Sub        []Interval
	// Payload interface{}
}

// idCounter is used to give Interval.id unique values for all intervals. Interval trees won't work
// without it.
var idCounter uintptr

// IntervalTree allows us to add methods on interval.Tree. Could do without it.
type IntervalTree interval.Tree

// NewIntv is a hack for testing.
func NewIntv(x0, x1 float64) Interval {
	s := Side{x0: x0, x1: x1}
	idCounter = (idCounter + 10) % 10
	i := Interval{Side: &s, id: idCounter}
	return i
}

// Range returns
func (i Interval) Range() (float64, float64) {
	return i.x0, i.x1
}

func newInterval(v0, v1 *Vertex, vertical bool) Interval {
	s := newSide(v0, v1, vertical)
	idCounter++
	return Interval{Side: s, id: idCounter}
}

func CreateIntervalTree(segments []*Side, name string) *IntervalTree {
	common.Log.Debug("CreateIntervalTree: %d %q", len(segments), name)
	tree := &IntervalTree{}
	for i, s := range segments {
		tree.Insert(s)
		sStart, sEnd := "(nil)", "(nil)"
		if s.start != nil {
			sStart = fmt.Sprintf("%+g", s.start.Point)
		}
		if s.end != nil {
			sEnd = fmt.Sprintf("%+g", s.end.Point)
		}
		common.Log.Debug("-- %d: %v=%s-%s tree=%v", i, s, sStart, sEnd, tree)
	}
	// This is critical!
	t := (*interval.Tree)(tree)
	t.AdjustRanges()
	tree.Validate()
	return tree
}

func (tree *IntervalTree) getIntervals() []Interval {
	var intervals []Interval
	t := (*interval.Tree)(tree)
	t.Do(func(e interval.Interface) bool {
		iv := e.(Interval)
		intervals = append(intervals, iv)
		return false
	})
	return intervals
}

func (tree *IntervalTree) Validate() {
	intervals := tree.getIntervals()
	ValidateIntervals(intervals)
}

func ValidateIntervals(intervals []Interval) {
	return
	x0Counts := map[float64]int{}
	x1Counts := map[float64]int{}
	facX := 1e8
	for i, iv := range intervals {
		x0, x1 := iv.Range()
		x0 = math.Round(x0*facX) / facX
		x1 = math.Round(x1*facX) / facX
		x0Counts[x0]++
		x1Counts[x1]++
		if x0Counts[x0] > 1 || x1Counts[x1] > 1 {
			common.Log.Error("-------------&&&---------------")
			for j, jv := range intervals[:i+1] {
				common.Log.Error("%4d: %s", j, jv)
			}
			// panic(fmt.Errorf("Duplicate interval i=%d iv=%v", i, iv))
		}
	}
}

func (tree *IntervalTree) Insert(s *Side) {
	tree.Validate()
	idCounter++                              // !@#$ Critical for passing tests
	i := Interval{Side: s, id: idCounter} // counter has effect

	t := (*interval.Tree)(tree)
	// d := *s
	if err := t.Insert(i, true); err != nil {
		panic(fmt.Errorf("IntervalTree.Insert s=%v err=%v", *s, err))
	}
	// common.Log.Info("Insert: s=%v->%v i=%v", d, *s, i)
	tree.Validate()
	common.Log.Debug("treeInsert: %v s=%+v", tree, *s)
}

func (tree *IntervalTree) Delete(s *Side) {
	i := Interval{Side: s}
	t := (*interval.Tree)(tree)
	t.Delete(i, true)
	common.Log.Debug("treeDelete: %v s=%+v", tree, *s)
}

func (tree *IntervalTree) QueryPoint(x float64, cb func(s *Side) bool) bool {
	var matched bool
	common.Log.Debug("QueryPoint: x=%g tree=%+v", x, tree)
	t := (*interval.Tree)(tree)
	q := query1d(x)
	ok := t.DoMatching(func(e interval.Interface) bool {
		iv := e.(Interval)
		matched := cb(iv.Side)
		common.Log.Debug(" iv=%#v matched=%t", *iv.Side, matched)
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
		x0, x1 = float64(bc._x0), float64(bc._x1)
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
		x0, x1 = bc.x0, bc.x1
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
		_x0:     i.Start().(Int),
		_x1:     i.End().(Int),
		Side: i.Side,
		id:      i.id}
}
func (i Interval) String() string {
	seg := "   (nil)    "
	if i.Side != nil {
		seg = fmt.Sprintf("%p[%g,%g)", i.Side, i.x0, i.x1)
	}
	return fmt.Sprintf("%15s#%d", seg, i.id)
}

type Mutable struct {
	_x0, _x1 Int
	*Side
	id uintptr
}

func (m *Mutable) Start() interval.Comparable { return m._x0 }
func (m *Mutable) End() interval.Comparable   { return m._x1 }
func (m *Mutable) SetStart(c interval.Comparable) {
	// common.Log.Info("Mutable.SetStart %g->%g", m.x0, float64(c.(Int)))
	// if isZero(m.x0+3.55642) && isZero(float64(c.(Int))+8.3346) {
	// 	panic("SetStart")
	// }
	m._x0 = c.(Int)

}
func (m *Mutable) SetEnd(c interval.Comparable) {
	// common.Log.Info("Mutable.SetEnd %g->%g", m.x1, float64(c.(Int)))
	m._x1 = c.(Int)
}

// func (t *interval.Tree) queryPoint(x float64, f func(h *Side)) {
// }
// func (t *interval.Tree) insert(h *Side) {
// }
// func (t *interval.Tree) remove(h *Side) {
// }
