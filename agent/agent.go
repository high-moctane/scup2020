package agent

import (
	"fmt"
	"os"
)

type Agent interface {
	Init() error
	Reset()
	Action(s []float64) (a []float64)
	Learn(s1, a1 []float64, r float64, s2, a2 []float64)
}

func SelectAgent() (Agent, error) {
	agentName, ok := os.LookupEnv("SCUP_AGENT_NAME")
	if !ok {
		return nil, fmt.Errorf("cannot find SCUP_AGENT_NAME")
	}

	var res Agent
	switch agentName {
	case "Q-Learning":
		res = new(QLearning)
	default:
		return nil, fmt.Errorf("invalid agent name")
	}

	return res, nil
}
