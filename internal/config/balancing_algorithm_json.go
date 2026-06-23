package config

import (
	"encoding/json"
	"fmt"
)

type BalancingAlgorithm int

const (
	RoundRobin BalancingAlgorithm = iota
	LeastConnections
)

var stringToAlgo = map[string]BalancingAlgorithm{
	"round-robin":       RoundRobin,
	"least-connections": LeastConnections,
}

var AlgoToString = map[BalancingAlgorithm]string{
	RoundRobin:       "round-robin",
	LeastConnections: "least-connections",
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
