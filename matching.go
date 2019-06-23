package clip

import (
	"sort"

	"github.com/unidoc/unipdf/v3/common"
)

// bipartiteIndependentSet computes a maximum independent set for a bipartite graph.
// It takes O(sqrt(V) * E) time.
//  `n`: the number of vertices in the first component
//  `m`: the number of vertices in the second component
//  `edges`: a list of edges in the bipartite graph represented by pairs of integers
// Returns: A pair of lists representing the maximum independent set for the graph
//   http://en.wikipedia.org/wiki/Maximum_independent_set
//   http://en.wikipedia.org/wiki/Bipartite_graph
// A set is independent if and only if its complement is a vertex cover.
// PROOF: A set V of vertices is an independent set
//    IFF every edge in the graph is adjacent to at most one member of V
//    IFF every edge in the graph is adjacent to at least one member not in V
//    IFF the complement of V is a vertex cover.
func bipartiteIndependentSet(n, m int, edges [][2]int) ([]int, []int) {
	coverL, coverR := BipartiteVertexCover(n, m, edges)
	setL, setR := complement(coverL, n), complement(coverR, m)
	common.Log.Info("bipartiteIndependentSet: n=%d m=%d edges=%v", n, m, edges)
	common.Log.Info("   coverL=%d %v -> setL=%d %v", len(coverL), coverL, len(setL), setL)
	common.Log.Info("   coverR=%d %v -> setR=%d %v", len(coverR), coverR, len(setR), setR)

	return setL, setR
}

// complement returns [0:`n`) / `list`.
func complement(list []int, n int) []int {
	common.Log.Debug("complement: n=%d list=%d %v", n, len(list), list)
	sort.Ints(list)

	result := make([]int, n-len(list))
	a, b := 0, 0
	for i := 0; i < n; i++ {
		if len(list) > 0 && list[a] == i {
			a++
		} else {
			result[b] = i
			b++
		}
	}
	common.Log.Debug("complement: result=%d %v", len(result), result)
	return result
}

// BipartiteVertexCover computes a minimum vertex cover of a bipartite graph.
//  `n`: number of vertices in the left component.
//  `m`: number of vertices in the right component.
//  `edges`: list of edges from the left component connecting to the right component represented
//      by pairs of integers between 0 and n-1,m-1 respectively
// Returns: A pair of lists representing the vertices in the left component and the right component
//   respectively which are in the cover.
// Internally, this implementation uses the Hopcroft-Karp algorithm and König's theorem to compute
// the minimal vertex cover of a bipartite graph in O(sqrt(V) * E) time.
// BipartiteMatching uses Hopscroft-Karp, BipartiteVertexCover function uses König's theorem as in
//    http://tryalgo.org/en/matching/2016/08/05/konig/
// https://en.wikipedia.org/wiki/Hopcroft%E2%80%93Karp_algorithm
// https://en.wikipedia.org/wiki/K%C5%91nig%27s_theorem_(graph_theory)
// A vertex cover in a graph is a set of vertices that includes at least one endpoint of every edge,
//  and a vertex cover is minimum if no other vertex cover has fewer vertices.
// A matching in a graph is a set of edges no two of which share an endpoint, and a matching is
//  maximum if no other matching has more edges.
func BipartiteVertexCover(n, m int, edges [][2]int) ([]int, []int) {
	match := BipartiteMatching(n, m, edges)

	// Initialize adjacency lists.
	matchCount := make([]int, n) // matchCount[l] = number of matchings containing left vertex `l`.

	matchL := make([]int, n) // matchL[l] = right vertex for  left vertex `l`.
	adjL := make([][]int, n) // Left vertex adjacency list.
	coverL := make([]int, n)
	for i := 0; i < n; i++ {
		matchL[i] = -1
	}

	matchR := make([]int, m)
	adjR := make([][]int, m)
	coverR := make([]int, m)
	for i := 0; i < m; i++ {
		matchR[i] = -1
	}

	// Unpack matching.
	for _, m := range match {
		l, r := m[0], m[1]
		matchL[l] = r
		matchR[r] = l
	}

	common.Log.Info("matchL=%d %v", len(matchL), matchL)
	for i, a := range matchL {
		common.Log.Info("%6d: %v", i, a)
	}
	common.Log.Info("matchR=%d %v", len(matchR), matchR)
	for i, a := range matchR {
		common.Log.Info("%6d: %v", i, a)
	}

	// Loop over edges. Fill adjacency lists with edged not in matching.
	for i, e := range edges {
		l, r := e[0], e[1]
		matched := matchL[l] == r
		common.Log.Info(" @ i=%d e=%v matched=%t count[%d]=%d", i, e, matched, l, matchCount[l])
		if matched {
			cnt := matchCount[l]
			matchCount[l]++
			if cnt == 0 {
				continue
			}
		}
		adjL[l] = append(adjL[l], r)
		adjR[r] = append(adjR[r], l)
		common.Log.Info("          adjL[%d]=%v adjR[%d]=%v", l, adjL[l], r, adjR[r])
	}

	common.Log.Info("matchCount=%d %v", len(matchCount), matchCount)
	for i, a := range matchCount {
		common.Log.Info("%6d: %v", i, a)
	}

	common.Log.Info("adjL=%d %v", len(adjL), adjL)
	for i, a := range adjL {
		common.Log.Info("%6d: %v", i, a)
	}
	common.Log.Info("adjR=%d %v", len(adjR), adjR)
	for i, a := range adjR {
		common.Log.Info("%6d: %v", i, a)
	}
	// panic("*")

	// Construct cover.
	var left, right []int

	for i := 0; i < n; i++ {
		list := bpWalk(i, adjL, matchL, coverL, matchR, coverR)
		right = append(right, list...)
		common.Log.Info("right=%v", right)
	}
	for i := 0; i < m; i++ {
		list := bpWalk(i, adjR, matchR, coverR, matchL, coverL)
		left = append(left, list...)
		common.Log.Info("left=%v", left)
	}

	// Clean up any left over edges
	for i := 0; i < n; i++ {
		if coverL[i] == 0 && matchL[i] >= 0 {
			coverR[matchL[i]] = 1
			coverL[i] = 1 // !@#$ Does this have any effect?
			left = append(left, i)
		}
	}

	sort.Ints(left)
	sort.Ints(right)

	common.Log.Info("matchL=%d %v", len(left), left)
	common.Log.Info("matchR=%d %v", len(right), right)

	return left, right
}

// bipartite walk
func bpWalk(v int, adjL [][]int, matchL, coverL, matchR, coverR []int) []int {
	var list []int
	common.Log.Info("-bpWalk: list=%v v=%d adjL=%v\n"+
		"\t\tmatchL=%v coverL=%v\n"+
		"\t\tmatchR=%v coverR=%v",
		list, v, adjL, matchL, coverL, matchR, coverR)
	if coverL[v] != 0 || matchL[v] >= 0 {
		return nil
	}
	for v >= 0 {
		coverL[v] = 1
		adj := adjL[v]
		next := -1
		// !@#$ Seems like an inefficient way to find max u: !coverR[u]
		for _, u := range adj {
			if coverR[u] != 0 {
				continue
			}
			next = u
		}
		if next < 0 {
			break
		}
		coverR[next] = 1
		list = append(list, next)
		v = matchR[next]
	}
	common.Log.Info("+bpWalk: list=%v v=%d adjL=%v\n"+
		"\t\tmatchL=%v coverL=%v\n"+
		"\t\tmatchR=%v coverR=%v",
		list, v, adjL, matchL, coverL, matchR, coverR)
	return list
}

// BipartiteMatching finds a maximum bipartite matching in an unweighted graph.
// The current implementation uses the Hopcroft-Karp algorithm and runs in O(sqrt(V) * E + V) time.
// `n`: The number of vertices in the first component.
// `m`: The number of vertices in the second component.
// `edges`: The list of edges, represented by pairs of integers between 0 and n-1,m-1 respectively.
// Returns: A list of edges representing the matching.
// https://en.wikipedia.org/wiki/Matching_(graph_theory)
// https://en.wikipedia.org/wiki/Hopcroft%E2%80%93Karp_algorithm#Pseudocode
// https://en.wikipedia.org/wiki/Berge%27s_lemma
// A maximal matching is a matching M of a graph G with the property that if any edge not in M is
//    added to M, it is no longer a matching, that is, M is maximal if it is not a subset of any other
//    matching in graph G. In other words, a matching M of a graph G is maximal if every edge in G has
//    a non-empty intersection with at least one edge in M.
// Given a matching M,
//    an alternating path is a path that begins with an unmatched vertex and[2] whose edges belong
//      alternately to the matching and not to the matching.
//    an augmenting path is an alternating path that starts from and ends on free (unmatched) vertices.
//    One can prove that a matching is maximum if and only if it does not have any augmenting path.
//      (This result is sometimes called Berge's lemma.)
func BipartiteMatching(n, m int, edges [][2]int) [][2]int {
	common.Log.Debug("BipartiteMatching: n=%d m=%d\nedges=%d %v", n, m, len(edges), edges)
	for i, e := range edges {
		common.Log.Debug("%6d: %v", i, e)
	}
	if len(edges) == 0 {
		// panic("no edges")
		return nil
	}
	validateEdges(n, m, edges)

	// Initalize adjacency list, visit flag, distance.
	adjN := make([][]int, n)
	gN := make([]int, n)
	dist := make([]int, n)
	for i := 0; i < n; i++ {
		gN[i] = -1
		// adjN[i] = nil
		dist[i] = MaxInt
	}
	adjM := make([][]int, m)
	gM := make([]int, m)
	for i := 0; i < m; i++ {
		gM[i] = -1
		// adjM[i] = nil
	}

	// Build adjacency matrix
	for _, e := range edges {
		common.Log.Debug("adjN=%d adjM=%d e=%v", len(adjN), len(adjM), e)
		adjN[e[0]] = append(adjN[e[0]], e[1])
		adjM[e[1]] = append(adjM[e[1]], e[0])
	}
	common.Log.Debug("adjN=%d", len(adjN))
	for i, a := range adjN {
		common.Log.Debug("%6d: %d %v", i, len(a), a)
	}
	common.Log.Debug("adjM=%d", len(adjM))
	for i, a := range adjM {
		common.Log.Debug("%6d: %d %v", i, len(a), a)
	}

	// Why isn't adjM used any more? !@#$
	dmax := MaxInt

	// Depth-first search
	var dfs func(v int) bool
	dfs = func(v int) bool {
		if v < 0 {
			return true
		}
		for _, u := range adjN[v] {
			pu := gM[u]
			dpu := dmax
			if pu >= 0 {
				dpu = dist[pu]
			}
			if dpu == dist[v]+1 {
				if dfs(pu) {
					gN[v], gM[u] = u, v
					return true
				}
			}
		}
		dist[v] = MaxInt
		return false
	}

	// Run search
	toVisit := make([]int, n)
	matching := 0
	for {
		// Initialize queue
		count := 0
		for i := 0; i < n; i++ {
			if gN[i] < 0 {
				dist[i] = 0
				toVisit[count] = i
				count++
			} else {
				dist[i] = MaxInt
			}
		}

		// Run BFS
		// Let G = (A∪B, E) be a bipartite graph and let M be a matching of G. We want to find
		// a maximum matching of G. Denote by A0,B0 the sets of M-unsaturated vertices in A,B
		// respectively.
		// First we will use breadth-first-search BFS to find the length k of a shortest path from
		// B0 to A0. Simultaneosuly, we produce the sequence of disjoint layers
		// B0 = L0, L1, ... Lk ⊆ A0 where
		//    Li is the set of vertices at distance i from B0 for all 0 ≤ i < k, and
		//    Lk is the subset of A0 which is at distance k from B0.
		// To avoid multiple BFSs from each vertex in B0, we add a super-vertex β and draw edges
		// from it to all vertices of B0. Start a BFS from β to get distance of β from A0. Subtract
		// one to get length of shortest path from B0 to A0. This takes O(m) time.
		dmax = MaxInt
		// for ptr := 0;  ptr < count; ptr++ {
		ptr := 0
		for ptr < count {
			v := toVisit[ptr]
			ptr++
			dv := dist[v]
			if dv < dmax {
				adj := adjN[v]
				l := len(adj)
				// for _, u := range adjN[v]{
				for j := 0; j < l; j++ {
					u := adj[j]
					pu := gM[u]
					if pu < 0 {
						if dmax == MaxInt {
							dmax = dv + 1
						}
					} else if dist[pu] == MaxInt {
						dist[pu] = dv + 1
						toVisit[count] = pu
						count++
					}
				}
			}
		}

		// Check for termination
		if dmax == MaxInt {
			break
		}

		// Run DFS on each vertex in N
		for v := 0; v < n; v++ {
			if gN[v] < 0 {
				if dfs(v) {
					matching += 1
				}
			}
		}
	}

	// Construct result
	count := 0
	result := make([][2]int, matching)
	for i := 0; i < n; i++ {
		if gN[i] < 0 {
			continue
		}
		result[count] = [2]int{i, gN[i]}
		count++
	}

	if count != matching {
		panic("Didn't expect this.")
	}
	common.Log.Info("BipartiteMatching: n=%d m=%d\n\t   edges=%d %v\n\tmatching=%d %v",
		n, m, len(edges), edges, len(result), result)

	return result
}

// validateEdges checks that `edges` is a valid set of edges over ranges `n` and `m`.
func validateEdges(n, m int, edges [][2]int) {
	for _, e := range edges {
		common.Log.Debug("n=%d m=%d e=%v", n, m, e)
		if e[0] >= n {
			panic("Bad e[0]")
		}
		if e[1] >= m {
			panic("Bad e[1]")
		}
	}
	i0min, i1min := MaxInt, MaxInt
	i0max, i1max := MinInt, MinInt
	for _, e := range edges {
		i0, i1 := e[0], e[1]
		if i0 < i0min {
			i0min = i0
		}
		if i0 > i0max {
			i0max = i0
		}
		if i1 < i1min {
			i1min = i1
		}
		if i1 > i1max {
			i1max = i1
		}
	}
	if i0min != 0 || i0max != n-1 || i1min != 0 || i1max != m-1 {
		common.Log.Error("Invalid edges: n=%d (i0min=%d i0max=%d) m=%d (i1min=%d i1max=%d)",
			n, i0min, i0max, m, i1min, i1max)
		panic("bad edges")
	}
}
