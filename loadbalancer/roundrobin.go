package loadbalancer

type RoundRobinPolicy struct {
	basePolicy
}

func (p *RoundRobinPolicy) ChooseTarget(targets []string) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	curr := p.currentIndex
	p.currentIndex = (p.currentIndex + 1) % len(targets)
	return targets[curr]
}
