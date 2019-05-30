package clip

import "sort"

func GetContours(array NDArray, clockwise bool) []Path {

	// First extract horizontal contours and vertices
	hcontours := getParallelCountours(array, false)
	hvertices := getVertices(hcontours)
	//   hvertices.sort(compareVertex)
	sort.Slice(hvertices, func(i, j int) bool { return compareVertex(hvertices[i], hvertices[j]) })

	// Extract vertical contours and vertices
	vcontours := getParallelCountours(array.Transpose(), true)
	vvertices := getVertices(vcontours)
	//   vvertices.sort(compareVertex)
	sort.Slice(vvertices, func(i, j int) bool { return compareVertex(vvertices[i], vvertices[j]) })

	// Glue horizontal and vertical vertices together
	for i, h := range hvertices {
		v := vvertices[i]
		if h.orientation {
			h.segment.next = v.segment
			v.segment.prev = h.segment
		} else {
			h.segment.prev = v.segment
			v.segment.next = h.segment
		}
	}

	// Unwrap loops
	var loops []Path
	for _, h := range hcontours {
		if !h.visited {
			loops = append(loops, walk(h, clockwise))
		}
	}
	return loops
}

type cSegment struct { // !@#$ Same as Segment?
	start, end int
	direction  bool
	height     int
	visited    bool
	prev, next *cSegment
}

func newCSegment(start, end int, direction bool, height int) *cSegment {
	return &cSegment{
		start:     start,
		end:       end,
		direction: direction,
		height:    height,
	}
}

type cVertex struct { // !@#$ Same as Vertex?
	x, y        int
	segment     *cSegment
	orientation bool
}

func newVertex(x, y int, segment *cSegment, orientation bool) *cVertex {
	return &cVertex{
		x:           x,
		y:           y,
		segment:     segment,
		orientation: orientation,
	}
}

func getParallelCountours(array NDArray, direction bool) []*cSegment {
	h, w := array.Shape()

	var contours []*cSegment
	// Scan top row
	var a, b, c, d bool

	x0 := 0
	var i, j int
	for j = 0; j < w; j++ {
		b = array[0][j] != 0.0
		if b == a {
			continue
		}
		if a {
			contours = append(contours, newCSegment(x0, j, direction, 0))
		}
		if b {
			x0 = j
		}
		a = b
	}
	if a {
		contours = append(contours, newCSegment(x0, j, direction, 0))
	}
	// Scan center
	for i = 1; i < h; i++ {
		a, b = false, false
		b = false
		x0 = 0
		for j := 0; j < w; j++ {
			c = array[i-1][j] != 0.0
			d = array[i][j] != 0.0
			if c == a && d == b {
				continue
			}
			if a != b {
				if a {
					contours = append(contours, newCSegment(j, x0, direction, i))
				} else {
					contours = append(contours, newCSegment(x0, j, direction, i))
				}
			}
			if c != d {
				x0 = j
			}
			a = c
			b = d
		}
		if a != b {
			if a {
				contours = append(contours, newCSegment(j, x0, direction, i))
			} else {
				contours = append(contours, newCSegment(x0, j, direction, i))
			}
		}
	}
	// Scan bottom row
	a = false
	x0 = 0
	for j = 0; j < w; j++ {
		b = array[h-1][j] != 0.0
		if b == a {
			continue
		}
		if a {
			contours = append(contours, newCSegment(j, x0, direction, h))
		}
		if b {
			x0 = j
		}
		a = b
	}
	if a {
		contours = append(contours, newCSegment(j, x0, direction, h))
	}
	return contours
}

func getVertices(contours []*cSegment) []*cVertex {
	vertices := make([]*cVertex, len(contours)*2)
	for i, h := range contours {
		if !h.direction {
			vertices[2*i] = newVertex(h.start, h.height, h, false)
			vertices[2*i+1] = newVertex(h.end, h.height, h, true)
		} else {
			vertices[2*i] = newVertex(h.height, h.start, h, false)
			vertices[2*i+1] = newVertex(h.height, h.end, h, true)
		}
	}
	return vertices
}

func walk(v *cSegment, clockwise bool) Path {
	var result Path
	for !v.visited {
		v.visited = true
		if v.direction {
			result = append(result, Point{X: float64(v.height), Y: float64(v.end)})
		} else {
			result = append(result, Point{X: float64(v.start), Y: float64(v.height)})
		}
		if clockwise {
			v = v.next
		} else {
			v = v.prev
		}
	}
	return result
}

func compareVertex(a, b *cVertex) bool {
	d := a.x - b.x
	if d != 0.0 {
		return d > 0
	}
	d = a.y - b.y
	if d != 0.0 {
		return d > 0
	}
	return a.orientation && !b.orientation
}
