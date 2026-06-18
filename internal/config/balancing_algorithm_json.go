package config

import (
	"encoding/json"
	"fmt"
)

type BalancingAlgorithm int

const (
	RoundRobinAlgo BalancingAlgorithm = iota //TODO remove 'Algo from the name after moving to separate packages'
	LeastConnectionsAlgo
)

var stringToAlgo = map[string]BalancingAlgorithm{
	"round-robin":       RoundRobinAlgo,
	"least-connections": LeastConnectionsAlgo,
}

var AlgoToString = map[BalancingAlgorithm]string{
	RoundRobinAlgo:       "round-robin",
	LeastConnectionsAlgo: "least-connections",
}

func (algo *BalancingAlgorithm) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	algorithm, ok := GetBalacingAlgorithm(s)
	if !ok {
		return fmt.Errorf("invalid balancing algorithm: %s", s)
	}

	*algo = algorithm
	return nil
}

func GetBalancingAlgorithmName(b BalancingAlgorithm) string {
	return AlgoToString[b]
}

func GetBalacingAlgorithm(name string) (BalancingAlgorithm, bool) {
	bl, ok := stringToAlgo[name]
	return bl, ok
}
