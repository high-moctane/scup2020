package lab_scup2020

import (
	"errors"
	"fmt"
	"os"

	"github.com/high-moctane/lab_scup2020/agent"
	"github.com/high-moctane/lab_scup2020/environment"
	"github.com/high-moctane/lab_scup2020/utils"
)

const (
	RLRunUpDown = iota
	RLRunUp
	RLRunDown
)

var EndOfEpisode = errors.New("end of episode")

type RL struct {
	env environment.Environment

	agentUp, agentDown                 agent.Agent
	rewardFuncUp, rewardFuncDown       func(s []float64) float64
	agentUpDataPath, agentDownDataPath string
	agentSaveFreq                      int

	maxEpisode             int
	maxStepUp, maxStepDown int
}

func NewRL() (*RL, error) {
	// Env
	env, err := environment.SelectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}
	if err := env.Init(); err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

	// Agent
	agentUp, err := agent.SelectAgent()
	if err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}
	if err := agentUp.Init(); err != nil {
		return nil, fmt.Errorf("new agentup failed: %w", err)
	}

	agentDown, err := agent.SelectAgent()
	if err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}
	if err := agentDown.Init(); err != nil {
		return nil, fmt.Errorf("new agentup failed: %w", err)
	}

	// Reward func
	rewardFuncUp := env.RewardFuncUp()
	rewardFuncDown := env.RewardFuncDown()

	// AgentDataPath and loading agent data
	agentDataNotFoundError := &agent.AgentDataNotFound{}

	agentUpDataPath, ok := os.LookupEnv("SCUP_RL_AGENT_UP_DATA_PATH")
	if !ok {
		return nil, fmt.Errorf("new rl failed: not found SCUP_RL_AGENT_UP_DATA_PATH")
	}
	if err := agentUp.Load(agentUpDataPath); err != nil && !errors.As(err, &agentDataNotFoundError) {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

	agentDownDataPath, ok := os.LookupEnv("SCUP_RL_AGENT_DOWN_DATA_PATH")
	if !ok {
		return nil, fmt.Errorf("new rl failed: not found SCUP_RL_AGENT_DOWN_DATA_PATH")
	}
	if err := agentDown.Load(agentDownDataPath); err != nil && !errors.As(err, &agentDataNotFoundError) {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

	// AgentSaveFreq
	agentSaveFreq, err := utils.GetEnvInt("SCUP_RL_AGENT_SAVE_FREQUENT")
	if err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

	// Max episodes and steps
	maxEpisode, err := utils.GetEnvInt("SCUP_RL_MAX_EPISODE")
	if err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

	maxStepUp, err := utils.GetEnvInt("SCUP_RL_MAX_STEP_UP")
	if err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

	maxStepDown, err := utils.GetEnvInt("SCUP_RL_MAX_STEP_UP")
	if err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

	res := &RL{
		env,
		agentUp,
		agentDown,
		rewardFuncUp,
		rewardFuncDown,
		agentUpDataPath,
		agentDownDataPath,
		agentSaveFreq,
		maxEpisode,
		maxStepUp,
		maxStepDown,
	}

	return res, nil
}

func (rl *RL) RunUpDown() error {
	for episode := 0; rl.maxEpisode == -1 || episode < rl.maxEpisode; episode++ {
		returns, err := rl.RunEpisodeUp(episode)
		if err != nil && !errors.Is(EndOfEpisode, err) {
			return fmt.Errorf("rl run error: %w", err)
		}
		fmt.Println("returnsUp:", returns)

		returns, err = rl.RunEpisodeDown(episode)
		if err != nil && !errors.Is(EndOfEpisode, err) {
			return fmt.Errorf("rl run error: %w", err)
		}
		fmt.Println("returnsDown:", returns)
	}

	return nil
}

func (rl *RL) RunUp() error {
	for episode := 0; rl.maxEpisode == -1 || episode < rl.maxEpisode; episode++ {
		returns, err := rl.RunEpisodeUp(episode)
		if err != nil && !errors.Is(EndOfEpisode, err) {
			return fmt.Errorf("rl run error: %w", err)
		}
		fmt.Println("returnsUp:", returns)
	}

	return nil
}

func (rl *RL) RunDown() error {
	for episode := 0; rl.maxEpisode == -1 || episode < rl.maxEpisode; episode++ {
		returns, err := rl.RunEpisodeDown(episode)
		if err != nil && !errors.Is(EndOfEpisode, err) {
			return fmt.Errorf("rl run error: %w", err)
		}
		fmt.Println("returnsDown:", returns)
	}

	return nil
}

func (rl *RL) Run(mode int) error {
	switch mode {
	case RLRunUpDown:
		return rl.RunUpDown()
	case RLRunUp:
		return rl.RunUp()
	case RLRunDown:
		return rl.RunDown()
	default:
		return fmt.Errorf("rl run error: invalid mode: %d", mode)
	}
}

func (rl *RL) RunEpisodeUp(episode int) (returns float64, err error) {
	return rl.RunEpisode(episode, RLRunUp)
}

func (rl *RL) RunEpisodeDown(episode int) (returns float64, err error) {
	return rl.RunEpisode(episode, RLRunDown)
}

func (rl *RL) RunEpisode(episode, mode int) (returns float64, err error) {
	// Reset
	if err = rl.env.Reset(); err != nil {
		err = fmt.Errorf("rl run error: %w", err)
		return
	}

	// Init
	var ag agent.Agent
	var maxStep int
	var rewardFunc func(s []float64) float64
	var agentDataPath string

	switch mode {
	case RLRunUp:
		ag = rl.agentUp
		maxStep = rl.maxStepUp
		rewardFunc = rl.rewardFuncUp
		agentDataPath = rl.agentUpDataPath
	case RLRunDown:
		ag = rl.agentDown
		maxStep = rl.maxStepDown
		rewardFunc = rl.rewardFuncDown
		agentDataPath = rl.agentDownDataPath
	}

	var s1, s2, a1, a2 []float64
	s1, err = rl.env.State()
	if err != nil {
		return
	}
	r := rewardFunc(s1)
	a1 = ag.Action(s1)

	returns += r

	// Run
	var isFinish bool

	for step := 0; step == -1 || step < maxStep; step++ {
		if err = rl.env.RunStep(a1); err != nil {
			return
		}

		s2, err = rl.env.State()
		if err != nil {
			return
		}

		r = rewardFunc(s2)

		a2 = ag.Action(s2)

		ag.Learn(s1, a1, r, s2, a2)

		if isFinish {
			break
		}

		isFinish, err = rl.env.IsFinish(s2)
		if err != nil {
			return
		}

		s1 = s2
		a1 = a2
		returns += r
	}

	// Save
	if rl.agentSaveFreq == -1 || episode%rl.agentSaveFreq == 0 {
		if err = ag.Save(agentDataPath); err != nil {
			err = fmt.Errorf("rl run error: %w", err)
			return
		}
	}

	return
}

func (rl *RL) Close() error {
	return rl.env.Close()
}
