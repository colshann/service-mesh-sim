package loadbalancer

type BalancingAlgorithm int

const (
	RoundRobin BalancingAlgorithm = 0
	Random     BalancingAlgorithm = 1
)
