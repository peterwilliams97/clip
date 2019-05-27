package main

import (
	"fmt"

	"github.com/peterwilliams97/clip"
)

func main() {
	testLine()
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
