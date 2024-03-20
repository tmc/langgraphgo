package graph

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
)

const END = "END"

type Node struct {
	Name     string
	Function func(ctx context.Context, state []llms.MessageContent) ([]llms.MessageContent, error)
}

type Edge struct {
	From string
	To   string
}

type MessageGraph struct {
	nodes      map[string]Node
	edges      []Edge
	entryPoint string
}

func NewMessageGraph() *MessageGraph {
	return &MessageGraph{
		nodes: make(map[string]Node),
	}
}

func (g *MessageGraph) AddNode(name string, fn func(ctx context.Context, state []llms.MessageContent) ([]llms.MessageContent, error)) {
	g.nodes[name] = Node{
		Name:     name,
		Function: fn,
	}
}

func (g *MessageGraph) AddEdge(from, to string) {
	g.edges = append(g.edges, Edge{
		From: from,
		To:   to,
	})
}

func (g *MessageGraph) SetEntryPoint(name string) {
	g.entryPoint = name
}

type Runnable struct {
	graph *MessageGraph
}

func (g *MessageGraph) Compile() (*Runnable, error) {
	if g.entryPoint == "" {
		return nil, fmt.Errorf("entry point not set")
	}

	return &Runnable{
		graph: g,
	}, nil
}

func (r *Runnable) Invoke(ctx context.Context, messages []llms.MessageContent) ([]llms.MessageContent, error) {
	state := messages
	currentNode := r.graph.entryPoint

	for {
		node, ok := r.graph.nodes[currentNode]
		if !ok {
			return nil, fmt.Errorf("node not found: %s", currentNode)
		}

		var err error
		state, err = node.Function(ctx, state)
		if err != nil {
			return nil, err
		}

		if currentNode == END {
			break
		}

		foundNext := false
		for _, edge := range r.graph.edges {
			if edge.From == currentNode {
				currentNode = edge.To
				foundNext = true
				break
			}
		}

		if !foundNext {
			return nil, fmt.Errorf("no outgoing edge found for node: %s", currentNode)
		}
	}

	return state, nil
}
