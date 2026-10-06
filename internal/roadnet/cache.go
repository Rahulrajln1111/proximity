package roadnet

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"proximity/internal/domain"
)

const maxLinkBytes = 512 << 20

type Source interface {
	Resolve(link string) (Router, error)
}

type Cache struct {
	grid   domain.Grid
	base   *Network
	client *http.Client

	mu    sync.RWMutex
	items map[string]*Network
}

func NewCache(g domain.Grid, base *Network) *Cache {
	return &Cache{
		grid:   g,
		base:   base,
		client: &http.Client{Timeout: 30 * time.Second},
		items:  make(map[string]*Network),
	}
}

func (c *Cache) Resolve(link string) (Router, error) {
	link = strings.TrimSpace(link)
	if link == "" {
		return c.base, nil
	}

	c.mu.RLock()
	hit, ok := c.items[link]
	c.mu.RUnlock()
	if ok {
		return hit, nil
	}

	var (
		net *Network
		err error
	)
	switch {
	case strings.HasPrefix(link, "http://"), strings.HasPrefix(link, "https://"):
		net, err = c.fetch(link)
	case strings.ContainsAny(link, "\n\r"):
		net, err = ParseNetwork(strings.NewReader(link), c.grid)
	default:
		net, err = c.open(link)
	}
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.items[link] = net
	c.mu.Unlock()
	return net, nil
}

func (c *Cache) fetch(url string) (*Network, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch link: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch link: status %d", resp.StatusCode)
	}
	return ParseNetwork(io.LimitReader(resp.Body, maxLinkBytes), c.grid)
}

func (c *Cache) open(path string) (*Network, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open link: %w", err)
	}
	defer f.Close()

	return ParseNetwork(io.LimitReader(f, maxLinkBytes), c.grid)
}
