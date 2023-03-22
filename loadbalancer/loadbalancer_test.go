package loadbalancer

import "testing"

func TestLoadBalancer(t *testing.T) {
	var lb LoadBalancer

	errFmt := "expected: %s, got: %s"
	targets := []string{"google.com", "bing.com"}

	// Round Robin
	lb = NewRoundRobinLoadBalancer()

	target := lb.ChooseTarget(targets)
	if target != targets[0] {
		t.Errorf(errFmt, targets[0], target)
	}

	target = lb.ChooseTarget(targets)
	if target != targets[1] {
		t.Errorf(errFmt, targets[1], target)
	}

	target = lb.ChooseTarget(targets)
	if target != targets[0] {
		t.Errorf(errFmt, targets[0], target)
	}

	// Weighted Round Robin
	weights := []int{3, 2}
	lb = NewWRoundRobinLoadBalancer(weights)

	for i := 0; i < weights[0]; i++ {
		target = lb.ChooseTarget(targets)
		if target != targets[0] {
			t.Errorf(errFmt, targets[0], target)
		}
	}

	for i := 0; i < weights[1]; i++ {
		target = lb.ChooseTarget(targets)
		if target != targets[1] {
			t.Errorf(errFmt, targets[1], target)
		}
	}

	// Least Connections
	lb = NewLeastConnLoadBalancer(targets, nil)

	target = lb.ChooseTarget(targets)
	target2 := lb.ChooseTarget(targets)

	lb.Done(target2)

	target3 := lb.ChooseTarget(targets)

	if target2 != target3 {
		t.Errorf(errFmt, target2, target3)
	}

	lb.Done(target)
	target4 := lb.ChooseTarget(targets)

	if target != target4 {
		t.Errorf(errFmt, target, target4)
	}

	lb.Done(target4)
	lb.Done(target3)

	// Weighted Least Connections
	weights = []int{2, 3}
	lb = NewLeastConnLoadBalancer(targets, weights)

	target = lb.ChooseTarget(targets)

	if target != targets[1] {
		t.Errorf(errFmt, targets[1], target)
	}

	target2 = lb.ChooseTarget(targets)

	if target2 != targets[0] {
		t.Errorf(errFmt, targets[0], target2)
	}

}
