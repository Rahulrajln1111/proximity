package roadnet

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"proximity/internal/domain"
)

type Router interface {
	Distances(src domain.Node) map[domain.Node]float64
}

type Network struct {
	adj map[domain.Node][]domain.Node
}

func NewNetwork() *Network {
	return &Network{adj: make(map[domain.Node][]domain.Node)}
}

func (n *Network) connect(a, b domain.Node) {
	n.adj[a] = append(n.adj[a], b)
	n.adj[b] = append(n.adj[b], a)
}

func (n *Network) Distances(src domain.Node) map[domain.Node]float64 {
	dist := make(map[domain.Node]float64, len(n.adj))
	dist[src] = 0

	queue := make([]domain.Node, 0, len(n.adj))
	queue = append(queue, src)

	for head := 0; head < len(queue); head++ {
		cur := queue[head]
		step := dist[cur] + 1
		for _, nb := range n.adj[cur] {
			if _, seen := dist[nb]; seen {
				continue
			}
			dist[nb] = step
			queue = append(queue, nb)
		}
	}
	return dist
}

func ParseNetwork(r io.Reader, g domain.Grid) (*Network, error) {
	net := NewNetwork()

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" {
			continue
		}

		fields := strings.Fields(text)
		if len(fields) != 4 {
			return nil, fmt.Errorf("link line %d: want 4 numbers, got %d", line, len(fields))
		}

		var v [4]float64
		for k, f := range fields {
			n, err := strconv.ParseFloat(f, 64)
			if err != nil {
				return nil, fmt.Errorf("link line %d: bad number %q: %w", line, f, err)
			}
			v[k] = n
		}

		a := g.Node(v[1], v[0])
		b := g.Node(v[3], v[2])
		if a == b {
			continue
		}
		net.connect(a, b)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read link: %w", err)
	}
	return net, nil
}

func CompleteNetwork(g domain.Grid, locations []domain.Location) *Network {
	net := NewNetwork()

	nodes := make(map[domain.Node]bool, len(locations))
	for _, loc := range locations {
		nodes[g.Node(loc.Lat, loc.Long)] = true
	}

	for node := range nodes {
		for _, d := range [2][2]int{{1, 0}, {0, 1}} {
			nb := domain.Node{I: node.I + d[0], J: node.J + d[1]}
			if nodes[nb] {
				net.connect(node, nb)
			}
		}
	}
	return net
}
