package loadbalancer

import (
	"math/rand"
	"time"
)

type RandomPolicy struct{}

func NewRandomLoadBalancer() *RandomPolicy {
	return &RandomPolicy{}
}

func (p *RandomPolicy) ChooseTarget(targets []string) string {
	rand.Seed(time.Now().UnixNano())

	return targets[rand.Intn(len(targets))]
}

func (p *RandomPolicy) Done(_ string) {}
