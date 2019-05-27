package main

import (
	"fmt"

	"github.com/peterwilliams97/clip"
)

func main() {
	testLine()
	testPolygon()
}

func testLine() {
	window := clip.Rect{5, 5, 10, 10}
	l := clip.NewLiangBarsky(window)

	fmt.Printf("window=%+v\n", l)
	for i := 0.0; i <= 15.0; i++ {
		line := clip.NewLine(0, i, 20, 20-i)
		clipped, ok := l.ClipLine(line)
		fmt.Printf("line=%+v --> ", line)
		if ok {
			fmt.Printf("clipped=%+v\n", clipped)
		} else {
			fmt.Println("outside")
		}
	}
}

func testPolygon() {
	window := clip.Rect{5, 5, 10, 10}
	l := clip.NewLiangBarsky(window)

	fmt.Printf("window=%+v\n", l)
	path := []clip.Point{
		clip.Point{7.5, 12.5},
		clip.Point{12.5, 7.5},
		clip.Point{7.5, 2.5},
		clip.Point{2.5, 7.5},
	}
	clipped := l.ClipPolygon(path)
	fmt.Println("=========================")
	fmt.Printf("l=%+v\n", l)
	fmt.Printf("path=%d %+v\n", len(path), path)
	fmt.Printf("clipped=%d %+v\n", len(clipped), clipped)

	path = []clip.Point{
		clip.Point{6.5, 11.5},
		clip.Point{11.5, 6.5},
		clip.Point{6.5, 3.5},
		clip.Point{3.5, 6.5},
	}
	clipped = l.ClipPolygon(path)
	fmt.Println("=========================")
	fmt.Printf("l=%+v\n", l)
	fmt.Printf("path=%d %+v\n", len(path), path)
	fmt.Printf("clipped=%d %+v\n", len(clipped), clipped)

	dx, dy := 1.0, 2.0
	llx, lly := l.Llx+dx, l.Lly+dy
	urx, ury := l.Urx+dx, l.Ury+dy
	path = []clip.Point{
		clip.Point{llx, lly},
		clip.Point{llx, ury},
		clip.Point{urx, ury},
		clip.Point{llx, ury},
	}
	clipped = l.ClipPolygon(path)
	fmt.Println("=========================")
	fmt.Printf("l=%+v\n", l)
	fmt.Printf("path=%d %+v\n", len(path), path)
	fmt.Printf("clipped=%d %+v\n", len(clipped), clipped)
}
