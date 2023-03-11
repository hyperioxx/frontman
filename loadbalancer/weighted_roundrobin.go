package loadbalancer

type WeightedRoundRobinPolicy struct {
	basePolicy
	baseWeights []int // TODO should validate that len(baseWeights) == len(targets) when adding new service and weights are > 0
	// TODO update those fields when service is modified via api call
	currentIndex  int
	currentWeight int
}

// TODO in the constructor currentWeight = baseWeights[0]
// [5, 2, 3]
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
