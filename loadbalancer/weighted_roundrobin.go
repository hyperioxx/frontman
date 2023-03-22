package loadbalancer

type WeightedRoundRobinPolicy struct {
	basePolicy
	weights       []int
	currentWeight int
}

func NewWRoundRobinLoadBalancer(weights []int) *WeightedRoundRobinPolicy {
	return &WeightedRoundRobinPolicy{
		weights:       weights,
		currentWeight: weights[0],
	}
}

func (p *WeightedRoundRobinPolicy) ChooseTarget(targets []string) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	curr := p.currentIndex

	if p.currentWeight == 0 {
		p.currentWeight = p.weights[p.currentIndex]
	}

	p.currentWeight--
	if p.currentWeight == 0 {
		p.currentIndex = (p.currentIndex + 1) % len(targets)
	}

	return targets[curr]
}

func (p *WeightedRoundRobinPolicy) Done(_ string) {}
