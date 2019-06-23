   Concave vertices
   ================
        ^          ^      >--+      +---<
    a)  |     b)   |    c)   |   d) |
        +--<    >--+         v      v

        v          v      <--+      +--->
    e)  |     f)   |    g)   |   h) |
        +-->    <--+         ^      ^

    anti-clockwise  a),c),f),h)
    ---------------------------
         +-<-+              +--<---+    +------+
        f|   |a             |      |    |      |
     +---+   +---+          |      |a  f|      |
     |           |          +---+  +----+  +---+
     +---+   +---+             c|          |h
        c|   |h                f|          |a
         +->-+              +---+  +----+  +---+
                            |      |h  c|      |
                            |      |    |      |
                            +-->---+    +------+

    clockwise  b),d),e),g)
    ----------------------
         +->-+              +-->---+    +------+
        b|   |e             |      |    |      |
     +---+   +---+          |      |e  b|      |
     |           |          +---+  +----+  +---+
     +---+   +---+             g|          |d
        g|   |d                b|          |e
         +-<-+              +---+  +----+  +---+
                            |      |d  g|      |
                            |      |    |      |
                            +--<---+    +------+


[[{0 0} {0 2} {3 2} {3 0} {2 0} {2 1} {1 1} {1 0}]]

    +-<-+   +---+
    |   |   |   |
    |   +---+   |    v={2 1}
    |       v0  |
    +-------|---+
   e0       v1  e1

    +-<-+   +---+
    |   |   |   |
    |   +---+   |    v={1 1}
    |   v0      |
    +---|-------+
   e0   v1      e1

[INFO]  rectangular-decomposition.go:81      0: {0 0 1 2}  A
[INFO]  rectangular-decomposition.go:81      1: {2 0 3 2}  B
[INFO]  rectangular-decomposition.go:81      2: {1 1 2 2}  C

    +-<-+   +---+
    |   |   |   |
    | A +---+ B |
    |   : C :   |
    +---+---+---+
   e0  v0   v1  e1

    +-<-+   +---+                         +--<
    |   |   |   |                  v0     |
    |   +---+v  |    v={2 1}    <--+      |o0
    |       :   |                  :      :
    +-------|---+               >--+      +-->
   e0       o   e1                 o0     o1

[INFO]  rectangular-decomposition.go:82 *** 8 rectangles
[INFO]  rectangular-decomposition.go:84      0: {2 0 3 2}
    +-<-+   +---+
    |   |   :   |
    |   +---: X |
    |   v0  :   |
    +---|-------+
[INFO]  rectangular-decomposition.go:84      1: {0 0 2 1}
    +-<-+   +---+
    |       :   |
    |········   |   BAD
    |   v0      |
    +---|-------+
[INFO]  rectangular-decomposition.go:84      2: {2 2 2 2}
[INFO]  rectangular-decomposition.go:84      3: {1 1 2 2}
[INFO]  rectangular-decomposition.go:84      4: {2 2 2 2}
[INFO]  rectangular-decomposition.go:84      5: {1 2 1 2}
[INFO]  rectangular-decomposition.go:84      6: {1 1 1 1}
[INFO]  rectangular-decomposition.go:84      7: {1 2 1 2}
[INFO]  rectangle_test.go:277 verifyDecomp:


[INFO]  rectangular-decomposition.go:64 **** splitters=2
[INFO]  rectangular-decomposition.go:66      0: CHORD{{1 2} 2 `vertical`  v={2 1} s={0 2} {3 2}} other={2 2}
[INFO]  rectangular-decomposition.go:66      1: CHORD{{1 2} 1 `vertical`  v={1 1} s={0 2} {3 2}} other={1 2}

[INFO]  rectangular-decomposition.go:578   findIntersection(CHORD{{1 2} 2 `vertical`  v={2 1} s={0 2} {3 2}}) -> 1 2
[INFO]  rectangular-decomposition.go:578   findIntersection(CHORD{{1 2} 1 `vertical`  v={1 1} s={0 2} {3 2}}) -> 1 2
[INFO]  rectangular-decomposition.go:586 intersections=2
[INFO]  rectangular-decomposition.go:590      0: [1 2] CHORD{{1 2} 2 `vertical`  v={2 1} s={0 2} {3 2}} intersects {0 2}-{3 2}

  something has gone wrong here
[INFO]  rectangular-decomposition.go:719   adjacent 1
[INFO]  rectangular-decomposition.go:720 	v0=VERTEX{index:0 prev:{1 0} point:{0 0} next:{0 2} concave:false visited:false}
[INFO]  rectangular-decomposition.go:721 	e0=VERTEX{index:1 prev:{0 0} point:{0 2} next:{3 2} concave:false visited: true}
[INFO]  rectangular-decomposition.go:722 	e1=VERTEX{index:2 prev:{0 2} point:{3 2} next:{3 0} concave:false visited:false}


