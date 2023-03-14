package loadbalancer

import (
	"sync"
)

const (
	RoundRobin         string = "round_robin"
	WeightedRoundRobin string = "weighted_round_robin"
)

type LoadBalancer interface {
	ChooseTarget(targets []string) string
}

type basePolicy struct {
	mu           sync.Mutex
	currentIndex int
}
