package gateway

import (
	"github.com/Frontman-Labs/frontman/service"
	"net/http"
	"strings"
	"sync"
)

type RoutingTrie struct {
	Mutex *sync.RWMutex
	root  *Route
}

type Route struct {
	label    string
	isEnd    bool
	service  *service.BackendService
	children map[string]*Route
}

func (rt *RoutingTrie) BuildRoutes(services []*service.BackendService) {
	root := &Route{
		label:    "",
		children: make(map[string]*Route),
	}
	for _, s := range services {
		insertNode(root, s)
	}

	rt.root = root
}

func insertNode(node *Route, service *service.BackendService) {
	// Handle domain-based routing first
	if service.Domain != "" {
		domainNode, ok := node.children[service.Domain]
		if !ok {
			domainNode = &Route{
				label:    service.Domain,
				children: make(map[string]*Route),
			}
			node.children[service.Domain] = domainNode
		}
		node = domainNode
	}

	segments := strings.Split(service.Path, "/")
	for _, s := range segments {
		if s == "" {
			continue
		}
		child, ok := node.children[s]
		if !ok {
			child = &Route{
				label:    s,
				children: make(map[string]*Route),
			}
			node.children[s] = child
		}
		node = child
	}

	node.isEnd = true
	node.service = service
}

func findBackendService(trie *RoutingTrie, r *http.Request) *service.BackendService {
	trie.Mutex.RLock()
	defer trie.Mutex.RUnlock()

	node := trie.root
	pathSegments := strings.Split(r.URL.Path, "/")
	domain := strings.Split(r.Host, ":")[0]

	// Check for domain-based routing first
	domainNode, ok := node.children[domain]
	if ok {
		if domainNode.service != nil {
			return domainNode.service
		}
		node = domainNode
	}

	// Check for path-based routing
	for i, segment := range pathSegments {
		if segment == "" {
			continue
		}

		child, ok := node.children[segment]
		if !ok {
			return node.service
		}

		if child.service != nil && child.service.Domain == domain {
			if i == len(pathSegments)-1 {
				return child.service
			}
			node = child
			continue
		}

		if child.service != nil && child.service.Domain == "" {
			node = child
			continue
		}

		node = child
	}

	return nil
}
