package main

import (
	"fmt"

	"github.com/peterwilliams97/clip"
)

func main() {
	testLine()
}

func testLine() {
	r := clip.Rect{5, 5, 10, 10}
	l := clip.NewLiangBarsky(r)

	fmt.Printf("l=%+v\n", l)
	for i := 0.0; i < 10.0; i++ {
		a := clip.Point{0, i}
		b := clip.Point{20, 20 + i}
		c, d, ok := l.ClipLine(a, b)
		fmt.Printf("a,b=%+v,%+v --> ", a, b)
		if ok {
			fmt.Printf("c,d=%+v,%+v\n", c, d)
		} else {
			fmt.Println("outside")
		}
	}
}
func testPoly() {
	x := []int{1, 2, 3}
	y := []int{1, 2, 3}
	rc, u, v := clip.LiangBarskyPolygonClip(x, y)
	fmt.Printf("rc=%d\n", rc)
	fmt.Printf("u=%+v\n", u)
	fmt.Printf("v=%+v\n", v)
}
