package clip

import (
	"math"
	"sort"

	"github.com/unidoc/unipdf/common"
)

// Overlap describes the overlap of 2 Rects in a slice.
type Overlap struct {
	I1, I2 int // Indexes of overlapping pairs of Rect's in Rect slice.
}

func (o Overlap) Equals(d Overlap) bool {
	return o.I1 == d.I1 && o.I2 == d.I2
}

// Michael Doescher
// October 10, 2013
// This program reports overlapping boxes
// Input = an array of box coordinates.  Each box is defined as an array of points.  The points
//         represent the lower left and upper right corner.
// Output = A two dimensional array.  Each row contains two values indicating the index value of the
//          boxes from the input that overlap
func BoxOverlap(boxes []Rect) []Overlap {
	events := generateEvents(boxes)
	common.Log.Debug("boxOverlap:\n\t boxes=%d %#v\n\t events=%d %#v",
		len(boxes), boxes, len(events), events)
	sort.Slice(events, func(i, j int) bool {
		a, b := events[i], events[j]
		if a.x < b.x {
			// common.Log.Error("@1 false a=%+v b=%+v", a, b)
			return true
		}
		if a.x > b.x {
			// common.Log.Error("@2  true a=%+v b=%+v", a, b)
			return false
		}
		if a.typ == "add" && b.typ == "remove" {
			// common.Log.Error("@3 false a=%+v b=%+v", a, b)
			return true // adding before removing allows for boxes that overlap
		}
		if a.typ == "remove" && b.typ == "add" {
			// common.Log.Error("@4  true a=%+v b=%+v", a, b)
			return false // only on the edge to count as overlapping.
		}
		// common.Log.Error("@5 false a=%+v b=%+v", a, b)
		return false
	})
	common.Log.Debug("boxOverlap:\n\t events=%d %#v", len(events), events)
	overlaps := generateOvelapList(boxes, events)
	sort.Slice(overlaps, func(i, j int) bool {
		oi, oj := overlaps[i], overlaps[j]
		if oi.I1 != oj.I1 {
			return oi.I1 < oj.I1
		}
		return oi.I2 < oj.I2
	})
	return overlaps
}

type oEvent struct {
	x     float64
	typ   string
	index int
}

func generateEvents(boxes []Rect) []oEvent {
	var leftEvents, rightEvents []oEvent

	for i, b := range boxes { // traverse the list of boxes
		leftx, rightx := math.Min(b.Llx, b.Urx), math.Max(b.Llx, b.Urx)
		leftEvents = append(leftEvents, oEvent{
			x:     leftx,
			typ:   "add",
			index: i,
		})
		rightEvents = append(rightEvents, oEvent{
			x:     rightx,
			typ:   "remove",
			index: i,
		})
	}
	events := make([]oEvent, len(leftEvents)+len(rightEvents))
	for i, e := range leftEvents {
		events[len(leftEvents)-1-i] = e
	}
	for i, e := range rightEvents {
		events[len(leftEvents)+i] = e
	}
	common.Log.Debug("generateEvents:\n\tboxes=%d %+v \n\t left=%d %+v\n\tright=%d %+v\n\t all=%d %+v",
		len(boxes), boxes, len(leftEvents), leftEvents, len(rightEvents), rightEvents, len(events), events)
	return events
}

func generateOvelapList(boxes []Rect, events []oEvent) []Overlap {
	var Q []int            // a list of indices into the boxes array of boxes that intersect the sweeping plane
	var overlaps []Overlap // pairs of boxes that overlap (indices into the boxes array

	common.Log.Debug("generateOvelapList:\n\t boxes=%d %+v\n\t events=%d %+v",
		len(boxes), boxes, len(events), events)
	for i, e := range events {
		common.Log.Debug("====================================")
		common.Log.Debug("i=%d e=%+v Q=%d", i, e, len(Q))
		if e.typ == "add" {
			overlaps = findOverlap(Q, e.index, overlaps, boxes)
			Q0 := Q
			Q = append(Q, e.index)
			common.Log.Debug("Q=%d %+v -> %d %+v (appended)", len(Q0), Q0, len(Q), Q)
		}
		if e.typ == "remove" {
			ind := indexOf(Q, e.index)
			if ind < 0 {
				continue
			}
			// for j, ee := range events {
			// 	common.Log.Debug("------------------------------------")
			// 	common.Log.Debug("j=%d (Q=%d) ee=%+v", j, len(Q), ee)
			// 	if Q[j] == ee.index {
			// 		ind = j
			// 		break
			// 	}
			// }
			// if ind < 0 {
			// 	panic("can't happen")
			// }
			// ind := Q.indexOf(events[i].index)
			// Q.splice(ind, 1);
			Q0 := Q
			Q = append(Q[:ind], Q[ind+1:]...)
			common.Log.Debug("Q=%d %+v -> %d %+v (removed %d)", len(Q0), Q0, len(Q), Q, ind)
		}
	}
	return overlaps
}

func indexOf(arr []int, v int) int {
	for i, a := range arr {
		if a == v {
			return i
		}
	}
	return -1
}
func findOverlap(Q []int, box int, overlaps []Overlap, boxes []Rect) []Overlap {
	if len(Q) == 0 {
		return overlaps
	}
	eb := boxes[box]
	ey1, ey2 := math.Min(eb.Lly, eb.Ury), math.Max(eb.Lly, eb.Ury)
	for _, q := range Q {
		b := boxes[q]
		y1, y2 := math.Min(b.Lly, b.Ury), math.Max(b.Lly, b.Ury)
		add := (y1 <= ey1 && ey1 <= y2) ||
			(y1 <= ey2 && ey2 <= y2) ||
			(ey1 < y1 && ey2 > y2)

		if add {
			o := createOverlap(q, box)
			overlaps = append(overlaps, o)

		}
	}
	return overlaps
}

func createOverlap(a, b int) Overlap {
	if a > b {
		a, b = b, a
	}
	return Overlap{a, b}
}
