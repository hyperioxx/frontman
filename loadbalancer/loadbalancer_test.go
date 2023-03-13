package loadbalancer

import "testing"

func TestLoadBalancer(t *testing.T) {
	var lb LoadBalancer
	targets := []string{"google.com", "bing.com"}

	// Round Robin
	lb = NewRoundRobinLoadBalancer()

	target := lb.ChooseTarget(targets)
	if target != targets[0] {
		t.Errorf("expected: %s, got: %s", targets[0], target)
	}

	target = lb.ChooseTarget(targets)
	if target != targets[1] {
		t.Errorf("expected: %s, got: %s", targets[1], target)
	}

	target = lb.ChooseTarget(targets)
	if target != targets[0] {
		t.Errorf("expected: %s, got: %s", targets[0], target)
	}

	// Weighted Round Robin
	weights := []int{3, 2}
	lb = NewWRoundRobinLoadBalancer(weights)

	for i := 0; i < weights[0]; i++ {
		target = lb.ChooseTarget(targets)
		if target != targets[0] {
			t.Errorf("expected: %s, got: %s", targets[0], target)
		}
	}

	for i := 0; i < weights[1]; i++ {
		target = lb.ChooseTarget(targets)
		if target != targets[1] {
			t.Errorf("expected: %s, got: %s", targets[1], target)
		}
	}
}
