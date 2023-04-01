package service

import (
	"net/http"
	"strings"
	"sync"
)

type RoutingTrie struct {
	mutex *sync.RWMutex
	root  *Route
}

type Route struct {
	label    string
	isEnd    bool
	service  *BackendService
	children map[string]*Route
}

func (rt *RoutingTrie) BuildRoutes(services []*BackendService) {
	rt.root = &Route{
		label:    "",
		children: make(map[string]*Route),
	}

	for _, s := range services {
		rt.insertNode(s)
	}
}

func (rt *RoutingTrie) FindBackendService(r *http.Request) *BackendService {
	rt.mutex.RLock()
	defer rt.mutex.RUnlock()

	node := rt.root
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

func (rt *RoutingTrie) insertNode(service *BackendService) {
	node := rt.root
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
