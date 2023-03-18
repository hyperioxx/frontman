package loadbalancer

type WeightedRoundRobinPolicy struct {
	basePolicy
	baseWeights   []int
	currentWeight int
}

func NewWRoundRobinLoadBalancer(weights []int) *WeightedRoundRobinPolicy {
	return &WeightedRoundRobinPolicy{
		baseWeights:   weights,
		currentWeight: weights[0],
	}
}

func (p *WeightedRoundRobinPolicy) ChooseTarget(targets []string) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	curr := p.currentIndex

	if p.currentWeight == 0 {
		p.currentWeight = p.baseWeights[p.currentIndex]
	}

	p.currentWeight--
	if p.currentWeight == 0 {
		p.currentIndex = (p.currentIndex + 1) % len(targets)
	}

	return targets[curr]
}

func (p *WeightedRoundRobinPolicy) Done(_ string) {}
