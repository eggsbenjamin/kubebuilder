package bpmn

type Graph struct {
	AdjacenyList map[string][]string
}

func NewGraph() *Graph {
	return &Graph{
		AdjacenyList: map[string][]string{},
	}
}

func (g *Graph) AddEdge(v1, v2 string) {
	g.AdjacenyList[v1] = append(g.AdjacenyList[v1], v2)
}
