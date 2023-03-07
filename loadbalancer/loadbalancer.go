package loadbalancer

type LoadBalancer interface {
	ChooseTarget(targets []string) string
}
