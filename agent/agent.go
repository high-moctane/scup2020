package agent

import (
	"fmt"
	"os"

	_ "github.com/high-moctane/lab_scup2020/logger"
)

type Agent interface {
	Init() error
	Reset()
	Action(s []float64) (a []float64)
	Learn(s1, a1 []float64, r float64, s2, a2 []float64)
	Save(string) error
	Load(string) error
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

	// logger.Get().Info("agent name: %s", agentName)

	return res, nil
}

type AgentDataNotFound struct {
	src string
}

func NewAgentDataNotFound(src string) *AgentDataNotFound {
	return &AgentDataNotFound{src}
}

func (e *AgentDataNotFound) Error() string {
	return fmt.Sprintf("agent data not found: %s", e.src)
}
