package clip

import (
	"fmt"

	"github.com/biogo/store/interval"
	"github.com/unidoc/unidoc/common"
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
	*Segment
	id uintptr
	// Sub        []Interval
	// Payload interface{}
}

type IntervalTree interval.Tree

func newInterval(v0, v1 *Vertex, vertical bool) Interval {
	return Interval{Segment: newSegment(v0, v1, vertical)}
}

func testSegment(v0, v1 *Vertex, tree *IntervalTree, vertical bool) bool {
	i := newInterval(v0, v1, vertical)
	t := (*interval.Tree)(tree)
	matches := t.Get(i)
	return len(matches) > 0
}

func CreateIntervalTree(segments []*Segment) *IntervalTree {
	common.Log.Debug("CreateIntervalTree: %d", len(segments))
	tree := &IntervalTree{}
	for _, s := range segments {
		tree.Insert(s)
	}
	return tree
}

func (tree *IntervalTree) Insert(s *Segment) {
	i := Interval{Segment: s}
	t := (*interval.Tree)(tree)
	if err := t.Insert(i, false); err != nil {
		panic(err)
	}
	common.Log.Debug("treeInsert: %v %v", tree, *s)
}

func (tree *IntervalTree) Delete(s *Segment) {
	i := Interval{Segment: s}
	t := (*interval.Tree)(tree)
	t.Delete(i, false)
	common.Log.Debug("treeDelete: %v %v", tree, *s)
}

func (tree *IntervalTree) QueryPoint(x float64, cb func(s *Segment) bool) bool {
	var matched bool
	common.Log.Debug("queryPoint: x=%g", x)
	t := (*interval.Tree)(tree)
	ok := t.Do(func(e interval.Interface) bool {
		i := e.(Interval)
		matched := cb(i.Segment)
		common.Log.Debug(" -- i=%#v matched=%t", i, matched)
		return matched
	})
	if matched != ok {
		panic("queryPoint")
	}
	return matched
}

// function testSegment(a, b, tree, direction) {
//   var ax = a.point[direction^1]
//   var bx = b.point[direction^1]
//   return !!tree.queryPoint(a.point[direction], function(s) {
//     var x = s.start.point[direction^1]
//     if(ax < x && x < bx) {
//       return true
//     }
//     return false
//   })
// }

func (i Interval) Overlap(b interval.Range) bool {
	var x0, x1 float64
	switch bc := b.(type) {
	case Interval:
		x0 = bc.x0
		x1 = bc.x1
	case *Mutable:
		x0, x1 = bc.x0, bc.x1
	default:
		panic("unknown type")
	}

	// Half-open interval indexing.
	return i.x1 > x0 && i.x0 < x1
}
func (i Interval) ID() uintptr                  { return i.id }
func (i Interval) Start() interval.Comparable   { return Int(i.x0) }
func (i Interval) End() interval.Comparable     { return Int(i.x1) }
func (i Interval) NewMutable() interval.Mutable { return &Mutable{Segment: i.Segment, id: i.id} }
func (i Interval) String() string {
	return fmt.Sprintf("[%g,%g)#%d",
		i.x0, i.x1, i.id)
}

type Mutable struct {
	*Segment
	id uintptr
}

func (m *Mutable) Start() interval.Comparable     { return Int(m.x0) }
func (m *Mutable) End() interval.Comparable       { return Int(m.x1) }
func (m *Mutable) SetStart(c interval.Comparable) { m.x0 = float64(c.(Int)) }
func (m *Mutable) SetEnd(c interval.Comparable)   { m.x1 = float64(c.(Int)) }

// func (t *interval.Tree) queryPoint(x float64, f func(h *Segment)) {
// }
// func (t *interval.Tree) insert(h *Segment) {
// }
// func (t *interval.Tree) remove(h *Segment) {
// }
