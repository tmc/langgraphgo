package graph

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
)

// END is a special constant used to represent the end node in the graph.
const END = "END"

var (
	// ErrEntryPointNotSet is returned when the entry point of the graph is not set.
	ErrEntryPointNotSet = errors.New("entry point not set")

	// ErrNodeNotFound is returned when a node is not found in the graph.
	ErrNodeNotFound = errors.New("node not found")

	// ErrNoOutgoingEdge is returned when no outgoing edge is found for a node.
	ErrNoOutgoingEdge = errors.New("no outgoing edge found for node")
)

// Node represents a node in the message graph.
type Node struct {
	// Name is the unique identifier for the node.
	Name string

	// Function is the function associated with the node.
	// It takes a context and a slice of MessageContent as input and returns a slice of MessageContent and an error.
	Function func(ctx context.Context, state []llms.MessageContent) ([]llms.MessageContent, error)
}

// Edge represents an edge in the message graph.
type Edge struct {
	// From is the name of the node from which the edge originates.
	From string

	// To is the name of the node to which the edge points.
	To string
}

// MessageGraph represents a message graph.
type MessageGraph struct {
	// nodes is a map of node names to their corresponding Node objects.
	nodes map[string]Node

	// edges is a slice of Edge objects representing the connections between nodes.
	edges []Edge

	// entryPoint is the name of the entry point node in the graph.
	entryPoint string
}

// NewMessageGraph creates a new instance of MessageGraph.
func NewMessageGraph() *MessageGraph {
	return &MessageGraph{
		nodes: make(map[string]Node),
	}
}

// AddNode adds a new node to the message graph with the given name and function.
func (g *MessageGraph) AddNode(name string, fn func(ctx context.Context, state []llms.MessageContent) ([]llms.MessageContent, error)) {
	g.nodes[name] = Node{
		Name:     name,
		Function: fn,
	}
}

// AddEdge adds a new edge to the message graph between the "from" and "to" nodes.
func (g *MessageGraph) AddEdge(from, to string) {
	g.edges = append(g.edges, Edge{
		From: from,
		To:   to,
	})
}

// SetEntryPoint sets the entry point node name for the message graph.
func (g *MessageGraph) SetEntryPoint(name string) {
	g.entryPoint = name
}

// Runnable represents a compiled message graph that can be invoked.
type Runnable struct {
	// graph is the underlying MessageGraph object.
	graph *MessageGraph
}

// Compile compiles the message graph and returns a Runnable instance.
// It returns an error if the entry point is not set.
func (g *MessageGraph) Compile() (*Runnable, error) {
	if g.entryPoint == "" {
		return nil, ErrEntryPointNotSet
	}

	return &Runnable{
		graph: g,
	}, nil
}

// Invoke executes the compiled message graph with the given input messages.
// It returns the resulting messages and an error if any occurs during the execution.
func (r *Runnable) Invoke(ctx context.Context, messages []llms.MessageContent) ([]llms.MessageContent, error) {
	state := messages
	currentNode := r.graph.entryPoint

	for {
		node, ok := r.graph.nodes[currentNode]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, currentNode)
		}

		var err error
		state, err = node.Function(ctx, state)
		if err != nil {
			return nil, fmt.Errorf("error in node %s: %w", currentNode, err)
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
			return nil, fmt.Errorf("%w: %s", ErrNoOutgoingEdge, currentNode)
		}
	}

	return state, nil
}
