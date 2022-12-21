package node

import (
	"sync"
)

// peers contains a list of peer nodes.
type peers struct {
	mu       *sync.RWMutex
	addrs    map[string]struct{}
	hostAddr string
}

// newPeers returns a list of peer addresses that is safe for concurrent access.
func newPeers(hostAddr string, peerAddrs []string) *peers {
	addrs := make(map[string]struct{}, len(peerAddrs))
	for _, addr := range peerAddrs {
		if addr != hostAddr {
			addrs[addr] = struct{}{}
		}
	}

	return &peers{
		mu:       &sync.RWMutex{},
		addrs:    addrs,
		hostAddr: hostAddr,
	}
}

// Add includes the peers passed to the list.
// Returns the updated number of peers in the list.
func (p *peers) Add(addrs ...string) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, addr := range addrs {
		if addr != p.hostAddr {
			p.addrs[addr] = struct{}{}
		}
	}
	return len(p.addrs)
}

// Contains returns true if the peer is part of the list.
func (p *peers) Contains(addr string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.addrs[addr]
	return ok
}

// List returns the list of peers nodes.
func (p *peers) List() []string {
	p.mu.RLock()
	list := make([]string, 0, len(p.addrs))
	for addr := range p.addrs {
		list = append(list, addr)
	}
	p.mu.RUnlock()

	return list
}

// ForEach calls f on each node in the peer list.
//
// It iterates over a copy of the list instead of the underlying map
// to prevent data races when the latter is modified in f.
func (p *peers) ForEach(f func(addr string) error) error {
	for _, addr := range p.List() {
		if err := f(addr); err != nil {
			return err
		}
	}
	return nil
}

// Remove takes the peers out of the list. Returns the updated number of peers remaining.
func (p *peers) Remove(addrs ...string) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, recvAddr := range addrs {
		delete(p.addrs, recvAddr)
	}

	return len(p.addrs)
}
