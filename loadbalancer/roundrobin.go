package loadbalancer

type RoundRobinPolicy struct {
	currentIndex int
}

func (p *RoundRobinPolicy) ChooseTarget(targets []string) string {
	p.currentIndex = (p.currentIndex + 1) % len(targets)
	return targets[p.currentIndex]
}
