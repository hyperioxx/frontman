package loadbalancer

import (
	"sync"
)

type LoadBalancer interface {
	ChooseTarget(targets []string) string
}

type basePolicy struct {
	mu sync.Mutex
}
