package loadbalancer

import (
	"sync"
)

const (
	RoundRobin         string = "round_robin"
	WeightedRoundRobin string = "weighted_round_robin"
	LeastConnection    string = "least_conn"
)

type LoadBalancer interface {
	ChooseTarget(targets []string) string
	Done(target string)
}

type basePolicy struct {
	mu           sync.Mutex
	currentIndex int
}
