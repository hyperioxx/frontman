package loadbalancer

import (
	"fmt"
	"github.com/Frontman-Labs/frontman/service"
	"sync"
)

const (
	RoundRobin         string = "rr"
	WeightedRoundRobin string = "weighted_rr"
)

type LoadBalancer interface {
	ChooseTarget(targets []string) string
}

type basePolicy struct {
	mu           sync.Mutex
	currentIndex int
}

func validatePolicy(s *service.BackendService) error {

	switch s.LoadBalancerPolicy.Type {
	case RoundRobin:
		return nil
	case WeightedRoundRobin:
		if len(s.LoadBalancerPolicy.Options.WeightedOptions.Weights) != len(s.UpstreamTargets) {
			fmt.Errorf("mismatched lengts of weights and tergets")
		}
	default:
		return fmt.Errorf("unknown loadbalancer policy: %s", s.LoadBalancerPolicy.Type)
	}

	return nil
}
