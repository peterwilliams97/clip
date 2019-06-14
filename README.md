# clip
Clipping code

http://library.utia.cas.cz/separaty/2012/ZOI/suk-rectangular%20decomposition%20of%20binary%20images.pdf
Having a binary object B (by a binary object we understand a set of all pixels of a binary image whose values equal one), we decompose it into K ≥ 1 blocks B1,B2, . . . , BK such that Bi ∩ Bj = ∅ for any i != j and B = union(Bk): k=1..B

https://www.sciencedirect.com/science/article/pii/0734189X84901397
Minimal Rectangular Partitions of Digitized Blobs
L. FERRAIU,* P.V. SANKAR,* AND J. SKLANSKY
Department of Electrical Engineering, Uniuersig of California, Irvine, California 92717
Received March 25,1983; accepted February 15,1984

An algorithm is presented for partitioning a finite region of the digital plane into a minimum number of rectangular regions. It is demonstrated that the partition problem is equivalent to finding the maximum number of independent vertices in a bipartite graph. The graph’s matching properties are used to develop an algorithm that solves the independent vertex problem. The solution of this graph-theoretical problem leads to a solution of the partition problem.
0 1984 by A&&c Press, Inc.

A rectangular partition of a blob on R, B, is a partition {Pi} i=1..M,
such that (∪ Pi = B) ∧ (Pi ∩ Pj = Ø i≠j) ∧ (Pi is a rectangle for all i).
M is defined as the order of the partition {Pi}.

LEMMA 1. For a blob on R whose boundary contains N noncogrid concave vertices and no cogrid concave
vertices, there exists a minimum order rectangular partition of order N + 1.

Let C’ denote a set of nonintersecting chords connecting points on the boundary
of a blob B. Let |C‘| = L’ and ci‘ denote an element of C‘, i = 1,2,...L’.
Let c denote a chord correcting two points on the boundary of B such that c and
C’ share no boundary points. Let x denote the number of intersections of c with C’.
We then state
LEMMA 2. The set C’ and the chord c partition B into L’ + x + 2 regions

THEOREM1. A blob B on a rectangular mosaic R has a minimum order rectangular partition of order
P = N - L + 1 where
    N = Total number of concave vertices on the boundary of B.
    L = Maximum number of nonintersecting chords that can be drawn between cogrid concave vertices.

The L nonintersecting chords partition B into (L + 1) subregions.
           +---+
           | 1 |
       +---+···|            N=1
       |   2   |            L=0
       +-------+            Rectangles=2

          +---+
          | 1 |
       +--+···|             N=2
       |  2   |             L=0
       +······+---+         Rectangles=3
       |    3     |
       +----------+

          +---+
          | 1 |
       +--+···+---+         N=2
       |    2     |         L=1
       +----------+         Rectangles=2

           +---+
           | 1 |
       +---+···+---+        N=4
       |     2     |        L=2  "Non-intersecting" must include not sharing a vertex.
       +---+···+---+        Rectangles=3
           | 3 |
           +---+

           +---+
           |   |
       +---+   +---+        N=4
       | 1 : 2 : 3 |        L=2  "Non-intersecting" must include not sharing a vertex.
       +---+   +---+        Rectangles=3
           |   |
           +---+

               +---+
               | 1 |
       +-------+···+---+    N=4
       |       2       |    L=2
       +---+···+-------+    Rectangles=3
           | 3 |
           +---+

           +---+
           | 1 |
       +---+···+---+        N=3
       |       : 3 |        L=1  "Non-intersecting" must include not sharing a vertex.
       |   2   +---+        Rectangles=3
       |       |
       +-------+

           +---+
           | 1 |
       +---+···|
       |   2   |            N=3
       |·······+------+     L=0
       |          : 4 |     Rectangles=4
       |    3     +---+
       |          |
       +----------+

       +-----------+        N=3
       |     1     |        L=1
       +---+···+---+        Rectangles=3
           |   |
       +---+ 3 |
       | 2 :   |
       +-------+


We first reduce the problem of finding L nonintersecting chords to a graph theory problem.
Define a graph G = (V, E) such that
(1) Each q ∈ V corresponds to cogrid chord, say i, of B.
(2) Each edge vi, vj E corresponds to the intersection of i and j in B.

Let G = (V, E) be a graph. See [6,7].
DEFINITION 6. A set of vertices (edges) which covers all the edges (vertices) of G
is called a vertex cover (edge cover) for G.
DEFINITION 7. The smallest  number of vertices (edges) in any vertex (edge) cover
for G is called a vertex (edge) covering number and denoted by α0(G) (α1(G)).
DEFINITION 8. A set of vertices (edges) in G is called independent if no two of its
members are adjacent.
DEFINITION 9. The largest number of vertices (edges) in an independent set is
called the vertex (edge) independence number, β0 (β1).
DEFINITION 10. A bipartite graph G is a graph whose vertex set V can be
partitioned into two subsets V1 and V2, such that every edge of G joins V1 with V2.

THEOREM 2. (Gallai): For any nontrivial connected graph G, p = α0 + β0 = α1 + β1 where p = |V|.

DEFINITION 11. A set of β1 independent edges in G is called a maximum matching of G.

THEOREM 3. (König): If G is bipartite, then the number of edges in a maximum
matching equals the vertex covering number, that is β1 = α0 .

6. THE GRAPH REDUCTION OF PROBLEM L CHORD
In graph-theoretic terms, we are able to restate problem L Chord as follows:
Graph Problem. For the graph G = (V, E) defined above, find the largest subset
of independent nodes of G, i.e., the largest subset of chords containing no intersections.
We note that in the rectangular partition, all chords are either horizontal or
vertical. Consequently we have the following lemma.

LEMMA 3. The graph G = (V, E) is a bipartite graph.

Suppose we have a maximum matching on a bipartite graph G = ((U, V), E). Let
the matching contain k edges (a k matching).
We observe the matching partitions the vertex sets U and V into sets U’, U”, V’,
and V”, respectively such that U ’ and V’ contain only matched vertices and U” and
V” contain independent vertices (Fig. 4). Let uivi designate the ith pair of vertices
in the matching.
LEMMA 4. There does not exist any path from U” to V” that contains an edge uivi.

Algorithm 1
Find the maximum independent set of vertices for a bipartite graph.
Step 1 - Find the maximum matching for the bipartite graph G = (U, V, E).
Step 2 - Color each pair of matched vertices (ui, vi) red. For each pair of red vertices do the
following:
   (a) If there exists an edge from ui to V” in G, color ui green and vi blue or,
       if there exists an edge from vi to U” in G, color vi green and vi blue.
   (b) Recursively color each remaining red vertex connected in G to a blue vertex         green, and color its matched vertex blue.
Step 3 - For all remaining pairs of red colored vertices ujvj: color uj blue and
vj green if vj is connected to a green vertex. Go to 2b.
Step 4-For all remaining pairs of red vertices color uj green and vj blue.
Step 5-Color all vertices u ∈ U” and all v ∈ V” blue. ∎


https://en.wikipedia.org/wiki/Matching_(graph_theory) A matching or independent edge set in a graph is a set of edges without common vertices.
   https://www.geeksforgeeks.org/maximum-bipartite-matching/
   https://www.geeksforgeeks.org/ford-fulkerson-algorithm-for-maximum-flow-problem/
   https://en.wikipedia.org/wiki/Edmonds_matrix

https://en.wikipedia.org/wiki/Independent_set_(graph_theory)
https://en.wikipedia.org/wiki/Hopcroft%E2%80%93Karp_algorithm
https://en.wikipedia.org/wiki/Blossom_algorithm
https://en.wikipedia.org/wiki/Flow_network
https://www.fileformat.info/info/unicode/block/mathematical_operators/utf8test.htm
https://en.wikipedia.org/wiki/Menger%27s_theorem
