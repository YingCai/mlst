package mlst

import (
	"fmt" // format outputs and errors
)

/*** Data Structures for Storing Edges and Graphs ***/

type Edge struct {
	Ends [2]int
}
type EdgeSet map[Edge]bool

type AdjList []int // a list of adjacent nodes
type Graph struct {
	Neighbors       []AdjList // Neighbors[i] is the adjacency list of node i
	NumOfComponents int       // number of non-isolated components
	NumNodes        int       // number of non-isolated nodes
	NumLeaves       int
	HasCycle        bool
}

/*** Operations on Edge ***/

// convert to string
func (e Edge) String() string {
	return fmt.Sprintf("(%d,%d)", e.Ends[0], e.Ends[1])
}

// check for edge errors, return nil otherwise.
func (e *Edge) Error() error {
	for i := range e.Ends {
		if e.Ends[i] < 0 || e.Ends[i] >= MaxNumNodes {
			return fmt.Errorf("Node %d out of range [0--%d] in edge %s.",
				e.Ends[i], MaxNumNodes-1, e)
		}
	}

	if e.Ends[0] == e.Ends[1] {
		return fmt.Errorf("Self-loop not allowed in edge %s.", e)
	}

	return nil
}

// normalize an edge so that Ends[0] <= Ends[1]
func (e *Edge) Normalize() {
	if e.Ends[0] > e.Ends[1] {
		e.Ends[0], e.Ends[1] = e.Ends[1], e.Ends[0]
	}
}

/*** Operations on EdgeSet ***/

// convert to Graph
func (E EdgeSet) Graph() *Graph {
	G := NewEmptyGraph(MaxNumNodes)

	for e := range E {
		G.AddEdge(e)
	}

	return G
}

/*** Operations on Graph ***/

func NewEmptyGraph(numNodes int) *Graph {
	var G Graph
	G.Neighbors = make([]AdjList, numNodes)
	for i := 0; i < numNodes; i++ {
		G.Neighbors[i] = make([]int, 0)
	}
	return &G
}

func (G *Graph) AddEdge(e Edge) {
	G.addEdgeUV(e.Ends[0], e.Ends[1])
}

func (G *Graph) addEdgeUV(u, v int) {
	G.addDirectedEdgeUV(u, v)
	G.addDirectedEdgeUV(v, u)
}

func (G *Graph) addDirectedEdgeUV(u, v int) {
	G.Neighbors[u] = append(G.Neighbors[u], v)
}

func (G *Graph) EdgesInOneComponent() bool {
	return G.NumOfComponents == 1
}

type DFSFunc func(node int, parent int)

func (G *Graph) Search() {
	visited := make([]bool, MaxNumNodes) // initialize visited to false
	G.NumNodes = 0
	G.NumLeaves = 0
	G.NumOfComponents = 0
	G.HasCycle = false

	// closure of DFS function to refer to the visited array
	var dfsFunc DFSFunc
	dfsFunc = func(node int, parent int) {
		visited[node] = true
		for _, n := range G.Neighbors[node] {
			if n != parent {
				if !visited[n] {
					dfsFunc(n, node)
				} else {
					G.HasCycle = true
				}
			}
		}
	}

	for i := range G.Neighbors {
		if len(G.Neighbors[i]) > 0 {
			G.NumNodes++ // found a non-isolated node
			if len(G.Neighbors[i]) == 1 {
				G.NumLeaves++
			}
			if !visited[i] {
				G.NumOfComponents++ // found a non-isolated component
				dfsFunc(i, -1)      // -1 is not present in the graph, acts as null
			}
		}
	}
}
