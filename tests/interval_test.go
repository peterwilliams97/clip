package clip_test

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"

	"github.com/peterwilliams97/clip"
	"github.com/unidoc/unipdf/common"
)

func init() {
	level := common.LogLevelDebug
	common.SetLogger(common.NewConsoleLogger(level))
}

// TestIntervals runs testPoint on some random intervals.
func TestInterval(t *testing.T) {
	for m := 1; m <= 55; m++ {
		for k := 1; k <= 5; k++ {
			common.Log.Info("==============*****================")

			points := []float64{1, -1, float64(clip.MinInt), float64(clip.MaxInt)}
			common.Log.Info("m=%d k=%d: %d intervals %d points", m, k, k*m, len(points))

			for j := 0; j < 10000; j++ {
				intervals := makeIntervals(k * m)
				validateIntervals(intervals)
				tree := createTree(intervals)
				validateIntervals(intervals)

				// for i := 0; i < 1000; i++ {
				// 	points = append(points, random())
				// }
				// for _, iv := range intervals {
				// 	x0, x1 := iv.Range()
				// 	points = append(points, x0, x1)
				// }

				for _, x := range points {
					testPoint(t, tree, intervals, x)
				}
			}
			common.Log.Info("PASS")

		}
	}
}

const r0 = -10.0
const r1 = 10.0
const fac = 1e5

var randHistory = map[float64]struct{}{}

// random returns a random float64 in the range [r0..r1]
func random() float64 {
	var x float64
	for i := 0; i < 100; i++ {
		x = r0 + (r1-r0)*rand.Float64()
		x = math.Round(x*fac) / fac
		if _, ok := randHistory[x]; !ok {
			break
		}
	}
	if _, ok := randHistory[x]; ok {
		panic(fmt.Errorf("%g is a repeat", x))
	}
	randHistory[x] = struct{}{}
	return x
}

// makeIntervals returns a slice of random intervals [x0, x1]:  r0 =< x0 <= x1 <= 2*r1
func makeIntervals(n int) []clip.Interval {
	var intervals []clip.Interval
	for i := 0; i < n; i++ {
		x0 := random()
		dx := random()
		x1 := x0 + math.Abs(dx)
		iv := clip.NewIntv(x0, x1)
		intervals = append(intervals, iv)
	}
	validateIntervals(intervals)
	return intervals
}

// createTree returns an IntervalTree for `intervals`.
func createTree(intervals []clip.Interval) *clip.IntervalTree {
	validateIntervals(intervals)
	segments := make([]*clip.Segment, len(intervals))
	for i, iv := range intervals {
		segments[i] = iv.Segment
	}
	validateIntervals(intervals)
	tree := clip.CreateIntervalTree(segments)
	validateIntervals(intervals)
	return tree
}

// testPoint checks that `p` is matched by the correct intervals in `tree`. `tree` must be
// constructed from `intervals`.
func testPoint(t *testing.T, tree *clip.IntervalTree, intervals []clip.Interval, p float64) {
	validateIntervals(intervals)
	var expected []clip.Interval
	for _, v := range intervals {
		x0, x1 := v.Range()
		if x0 <= p && p < x1 {
			expected = append(expected, v)
		}
	}
	sortIntervals(expected)

	var actual []clip.Interval
	tree.QueryPoint(p, func(s *clip.Segment) bool {
		actual = append(actual, clip.Interval{Segment: s})
		return false
	})
	sortIntervals(actual)

	if !sameIntervals(expected, actual) {
		sortIntervals(intervals)
		common.Log.Error("===============================")
		a0, a1 := 0.0, 0.0
		for i, iv := range intervals {
			x0, x1 := iv.Range()
			common.Log.Error("%3d: %v (%+g %+g)", i, iv, x0-a0, x1-a1)
			a0, a1 = x0, x1
		}
		common.Log.Error("p=%g", p)

		showDifference(expected, actual)
		t.Fatalf("QueryPoint:\n\texpected=%d %v\n\tactual=%d %v",
			len(expected), expected, len(actual), actual)
	}
}

// sortIntervals sorts `intervals` by their lower then their bounds in ascending order.
func sortIntervals(intervals []clip.Interval) {
	validateIntervals(intervals)
	sort.Slice(intervals, func(i, j int) bool {
		a, b := intervals[i], intervals[j]
		a0, a1 := a.Range()
		b0, b1 := b.Range()
		if a0 != b0 {
			return a0 < b0
		}
		return a1 < b1
	})
	validateIntervals(intervals)
}

func validateIntervals(intervals []clip.Interval) {
	clip.ValidateIntervals(intervals)
	// x0Counts := map[float64]int{}
	// x1Counts := map[float64]int{}
	// facX := fac * 100.0
	// for i, iv := range intervals {
	// x0, x1 := iv.Range()
	// x0 = math.Round(x0*facX) / facX
	// x1 = math.Round(x1*facX) / facX
	// x0Counts[x0]++
	// x1Counts[x1]++
	// if x0Counts[x0] > 1 || x1Counts[x1] > 1 {
	// 	common.Log.Error("-----------------------------")
	// 	for j, jv := range intervals[:i+1] {
	// 		common.Log.Error("%4d: %v", j, jv)
	// 	}

	// 	panic(fmt.Errorf("Duplicate interval i=%d iv=%v", i, iv))
	// }
	// }
}

// sameIntervals returns true if `intervals0` and `intervals1` are the same.
func sameIntervals(intervals0, intervals1 []clip.Interval) bool {
	if len(intervals0) != len(intervals1) {
		return false
	}
	for i, a := range intervals0 {
		b := intervals1[i]
		a0, a1 := a.Range()
		b0, b1 := b.Range()
		if a0 != b0 || a1 != b1 {
			return false
		}
	}
	return true
}

func showDifference(intervals0, intervals1 []clip.Interval) {
	n := len(intervals0)
	if len(intervals1) > n {
		n = len(intervals1)
	}
	for i := 0; i < n; i++ {
		iv0, iv1 := clip.Interval{}, clip.Interval{}

		if i < len(intervals0) {
			iv0 = intervals0[i]
		}
		if i < len(intervals1) {
			iv1 = intervals1[i]
		}
		marker := "***"
		if i < len(intervals0) && i < len(intervals1) {
			a0, a1 := iv0.Range()
			b0, b1 := iv1.Range()
			if a0 == b0 && a1 == b1 {
				marker = ""
			}
		}

		common.Log.Info("%3d: %v %v %s", i, iv0, iv1, marker)
	}
}

// tape('fuzz test', function(t) {
//   function verifyTree(tree, intervals) {
//     function testInterval(x) {
//       var expected = []
//       if(x[0] <= x[1])
//       for(var j=0; j<intervals.length; ++j) {
//         var y = intervals[j]
//         if(x[1] >= y[0] && y[1] >= x[0]) {
//           expected.push(y)
//         }
//       }
//       expected.sort(cmpInterval)

//       var actual = []
//       tree.queryInterval(x[0], x[1], function(j) {
//         actual.push(j)
//       })
//       actual.sort(cmpInterval)

//       t.same(actual, expected, 'query interval: ' + x)
//     }
//     for(var i=0; i<intervals.length; ++i) {
//       testInterval(intervals[i])
//     }
//     testInterval([-Infinity, Infinity])
//     testInterval([0,0])
//     testInterval([Infinity, -Infinity])
//     for(var i=0; i<100; ++i) {
//       testInterval([Math.random(), 2*Math.random()])
//     }

//     //Verify queryPoint
//     function testPoint(p) {
//       var expected = []
//       for(var j=0; j<intervals.length; ++j) {
//         var y = intervals[j]
//         if(y[0] <= p && p <= y[1]) {
//           expected.push(y)
//         }
//       }
//       expected.sort(cmpInterval)

//       var actual = []
//       tree.queryPoint(p, function(y) {
//         actual.push(y)
//       })
//       actual.sort(cmpInterval)

//       t.same(actual, expected, 'query point: ' + p)
//     }
//     for(var i=0; i<intervals.length; ++i) {
//       testPoint(intervals[i][0])
//       testPoint(intervals[i][1])
//     }
//     testPoint(0)
//     testPoint(-1)
//     testPoint(1)
//     testPoint(-Infinity)
//     testPoint(Infinity)
//     for(var i=0; i<100; ++i) {
//       testPoint(Math.random())
//     }

//     //Check tree contents
//     t.equals(tree.count, intervals.length, 'interval count ok')
//     var treeIntervals = tree.intervals.slice()
//     treeIntervals.sort(cmpInterval)
//     var expectedIntervals = intervals.slice()
//     expectedIntervals.sort(cmpInterval)
//     t.same(treeIntervals, expectedIntervals, 'intervals same')

//     //Check tree invariants
//     function verifyNode(node, left, right) {
//       if(!node) {
//         return 0
//       }
//       var midp = node.mid
//       t.ok(left < midp && midp < right, 'mid point in range: ' + node.mid + ' in ' + [left,right])

//       //Verify left end points in ascending order
//       var leftP = node.leftPoints.slice()
//       for(var i=0; i<leftP.length; ++ i) {
//         if(i > 0) {
//           t.ok(leftP[i][0] >= leftP[i-1][0], 'order ok')
//         }
//         var y = leftP[i]
//         t.ok(y[0] <= midp && midp <= y[1], 'interval ok')
//       }

//       //Verify right end points in ascending order
//       var rightP = node.rightPoints.slice()
//       for(var i=0; i<rightP.length; ++ i) {
//         if(i > 0) {
//           t.ok(rightP[i][1] >= rightP[i-1][1], 'order ok')
//         }
//         var y = rightP[i]
//         t.ok(y[0] <= midp && midp <= y[1], 'interval ok')
//       }

//       leftP.sort(cmpInterval)
//       rightP.sort(cmpInterval)
//       t.same(leftP, rightP, 'intervals are consistent')

//       var leftCount = verifyNode(node.left, left, node.mid)
//       var rightCount = verifyNode(node.right, node.mid, right)
//       var actualCount = leftCount + rightCount + leftP.length
//       t.equals(node.count, actualCount, 'node count consistent')

//       return actualCount
//     }
//     verifyNode(tree.root, -Infinity, Infinity)
//   }

//   //Check empty tree
//   verifyTree(createIntervalTree(), [])

//   //Try trees with uniformly distributed end points
//   for(var count=0; count<10; ++count) {
//     //Create empty tree and insert 100 intervals
//     var intervals = []
//     var tree = createIntervalTree()
//     for(var i=0; i<100; ++i) {
//       var a = Math.random()
//       var b = a + Math.random()
//       var x = [a,b]
//       intervals.push(x)
//       tree.insert(x)
//     }
//     verifyTree(tree, intervals)

//     //Remove half the intervals
//     for(var i=99; i>=50; --i) {
//       tree.remove(intervals.pop())
//     }
//     verifyTree(tree, intervals)
//   }

//   //Trees with quantized end points
//   for(var count=0; count<10; ++count) {
//     //Create empty tree and insert 100 intervals
//     var intervals = []
//     var tree = createIntervalTree()
//     for(var i=0; i<100; ++i) {
//       var a = Math.floor(8.0*Math.random())/8.0
//       var b = Math.max(a + Math.floor(8.0*Math.random())/8.0, 1.0)
//       var x = [a,b]
//       intervals.push(x)
//       tree.insert(x)
//     }
//     verifyTree(tree, intervals)

//     //Remove half the intervals
//     for(var i=99; i>=50; --i) {
//       tree.remove(intervals.pop())
//     }
//     verifyTree(tree, intervals)
//   }

//   var stackIntervals = []
//   for(var i=0; i<100; ++i) {
//     stackIntervals.push([i/200, 1.0-(i/200)])
//   }
//   var tree = createIntervalTree(stackIntervals)
//   verifyTree(tree, stackIntervals)

//   t.end()
// })

// tape('containment', function(t) {
//   var tree = createIntervalTree([[0, 100]])
//   var count = 0
//   function incr() { count++ }

//   tree.queryInterval(10, 20, incr)
//   t.equals(count, 1)

//   count = 0;
//   tree.queryInterval(100, 100, incr)
//   t.equals(count, 1)

//   count = 0;
//   tree.queryInterval(110, 111, incr)
//   t.equals(count, 0)

//   var tree = createIntervalTree([[0, 20], [30, 50]])
//   count = 0
//   tree.queryInterval(10, 15, incr)
//   t.equals(count, 1)

//   count = 0
//   tree.queryInterval(25, 26, incr)
//   t.equals(count, 0)

//   count = 0
//   tree.queryInterval(35, 40, incr)
//   t.equals(count, 1)

//   t.end()
// })
