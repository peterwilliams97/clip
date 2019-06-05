package clip

import (
	"sort"

	"github.com/unidoc/unidoc/common"
)

// bipartiteIndependentSet computes a maximum independent set for a bipartite graph.
// It takes O(sqrt(V) * E) time.
// * `n` is a the number of vertices in the first component
// * `m` is the number of vertices in the second component
// * `edges` is a list of edges in the bipartite graph represented by pairs of integers
// **Returns** A pair of lists representing the maximum independent set for the graph
// http://en.wikipedia.org/wiki/Maximum_independent_set
// http://en.wikipedia.org/wiki/Bipartite_graph
// A set is independent if and only if its complement is a vertex cover.
// PROOF: A set V of vertices is an independent set
//    IFF every edge in the graph is adjacent to at most one member of V
//    IFF every edge in the graph is adjacent to at least one member not in V
//    IFF the complement of V is a vertex cover.
func bipartiteIndependentSet(n, m int, edges [][2]int) ([]int, []int) {
	coverL, coverR := bipartiteVertexCover(n, m, edges)
	return complement(coverL, n), complement(coverR, m)
}

// func compareInt(a, b int) {
//   return a - b
// }
// complement returns [0:`n`) \ `list`
func complement(list []int, n int) []int {
	sort.Ints(list)

	result := make([]int, n-len(list))
	a, b := 0, 0
	for i := 0; i < n; i++ {
		if list[a] == i {
			a++
		} else {
			result[b] = i
			b++
		}
	}
	return result
}

// bipartiteVertexCover computes a minimum vertex cover of a bipartite graph.
//  `n`: number of vertices in the left component
//  `m`: number of vertices in the right component
//  `edges`:  list of edges from the left component connecting to the right component represented
//      by pairs of integers between 0 and n-1,m-1 respectively
// Returns A pair of lists representing the vertices in the left component and the right component
//   respectively which are in the cover.
// Internally, this implementation uses the Hopcroft-Karp algorithm and König's theorem to compute
// the minimal vertex cover of a bipartite graph in O(sqrt(V) * E) time.
// BipartiteMatching uses Hopscroft-Karp, this function uses König's theorem as in
//    http://tryalgo.org/en/matching/2016/08/05/konig/
// https://en.wikipedia.org/wiki/Hopcroft%E2%80%93Karp_algorithm
// https://en.wikipedia.org/wiki/K%C5%91nig%27s_theorem_(graph_theory)
func bipartiteVertexCover(n, m int, edges [][2]int) ([]int, []int) {
	match := BipartiteMatching(n, m, edges)

	// Initialize adjacency lists
	adjL := make([][]int, n)
	matchL := make([]int, n)
	matchCount := make([]int, n)
	coverL := make([]int, n)
	for i := 0; i < n; i++ {
		// adjL[i] = nil
		matchL[i] = -1
		// matchCount[i] = 0
		// coverL[i] = 0
	}
	adjR := make([][]int, m)
	matchR := make([]int, m)
	coverR := make([]int, m)
	for i := 0; i < m; i++ {
		// adjR[i] = nil
		matchR[i] = -1
		// coverR[i] = 0
	}

	// Unpack matching.
	for _, m := range match {
		s, t := m[0], m[1]
		matchL[s] = t
		matchR[t] = s
	}

	// Loop over edges.
	for _, e := range edges {
		s, t := e[0], e[1]
		if matchL[s] == t {
			cnt := matchCount[s]
			matchCount[s]++
			if cnt == 0 {
				continue
			}
		}
		adjL[s] = append(adjL[s], t)
		adjR[t] = append(adjR[t], s)
	}

	// Construct cover
	var left []int
	var right []int
	for i := 0; i < n; i++ {
		bpWalk(right, i, adjL, matchL, coverL, matchR, coverR)
	}
	for i := 0; i < m; i++ {
		bpWalk(left, i, adjR, matchR, coverR, matchL, coverL)
	}

	// Clean up any left over edges
	for i := 0; i < n; i++ {
		if coverL[i] == 0 && matchL[i] >= 0 {
			coverR[matchL[i]] = 1
			coverL[i] = 1 // !@#$ Does this have any effect?
			left = append(left, i)
		}
	}

	return left, right
}

// bipartite walk
func bpWalk(list []int, v int, adjL [][]int, matchL, coverL, matchR, coverR []int) {
	if coverL[v] != 0 || matchL[v] >= 0 {
		return
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
}

// BipartiteMatching finds a maximum bipartite matching in an unweighted graph.
// The current implementation uses the Hopcroft-Karp algorithm and runs in O(sqrt(V) * E + V) time.
// `n`: The number of vertices in the first component.
// `m`:The number of vertices in the second component.
// `edges`: The list of edges, represented by pairs of integers between 0 and n-1,m-1 respectively.
// Returns: A list of edges representing the matching.
// https://en.wikipedia.org/wiki/Matching_(graph_theory)
// https://en.wikipedia.org/wiki/Hopcroft%E2%80%93Karp_algorithm#Pseudocode
func BipartiteMatching(n, m int, edges [][2]int) [][2]int {
	common.Log.Debug("BipartiteMatching: n=%d m=%d\nedges=%d %v", n, m, len(edges), edges)
	if len(edges) == 0 {
		return nil
	}
	validateEdges(n, m, edges)

	// Initalize adjacency list, visit flag, distance.
	adjN := make([][]int, n)
	g1 := make([]int, n)
	dist := make([]int, n)
	for i := 0; i < n; i++ {
		g1[i] = -1
		// adjN[i] = nil
		dist[i] = MaxInt
	}
	adjM := make([][]int, m)
	g2 := make([]int, m)
	for i := 0; i < m; i++ {
		g2[i] = -1
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

	// Depth-first search?
	var dfs func(v int) bool
	dfs = func(v int) bool {
		if v < 0 {
			return true
		}
		adj := adjN[v]
		for _, u := range adj {
			pu := g2[u]
			dpu := dmax
			if pu >= 0 {
				dpu = dist[pu]
			}
			if dpu == dist[v]+1 {
				if dfs(pu) {
					g1[v] = u
					g2[u] = v
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
			if g1[i] < 0 {
				dist[i] = 0
				toVisit[count] = i
				count++
			} else {
				dist[i] = MaxInt
			}
		}

		// Run BFS
		ptr := 0
		dmax = MaxInt
		for ptr < count {
			v := toVisit[ptr]
			ptr++
			dv := dist[v]
			if dv < dmax {
				adj := adjN[v]
				l := len(adj)
				for j := 0; j < l; j++ {
					u := adj[j]
					pu := g2[u]
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
		for i := 0; i < n; i++ {
			if g1[i] < 0 {
				if dfs(i) {
					matching += 1
				}
			}
		}
	}

	// Construct result
	count := 0
	result := make([][2]int, matching)
	for i := 0; i < n; i++ {
		if g1[i] < 0 {
			continue
		}
		result[count] = [2]int{i, g1[i]}
		count++
	}

	if count != matching {
		panic("Didn't expect this.")
	}

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
