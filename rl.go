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
	RLRunUp = iota
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
	env, err := environment.SelectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}
	if err := env.Init(); err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

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

	rewardFuncUp := env.RewardFuncUp()
	rewardFuncDown := env.RewardFuncDown()

	agentUpDataPath, ok := os.LookupEnv("SCUP_RL_AGENT_UP_DATA_PATH")
	if !ok {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

	agentDownDataPath, ok := os.LookupEnv("SCUP_RL_AGENT_UP_DATA_PATH")
	if !ok {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

	agentSaveFreq, err := utils.GetEnvInt("SCUP_RL_AGENT_SAVE_FREQUENT")
	if err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

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

func (rl *RL) Run() error {
	for episode := 0; rl.maxEpisode == -1 || episode < rl.maxEpisode; episode++ {
		// Up
		if err := rl.env.Reset(); err != nil {
			return fmt.Errorf("rl run error: %w", err)
		}

		returns, err := rl.RunEpisode(RLRunUp)
		if err != nil && !errors.Is(EndOfEpisode, err) {
			return fmt.Errorf("rl run error: %w", err)
		}
		fmt.Println("returnsUp:", returns)

		// Down
		if err := rl.env.Reset(); err != nil {
			return fmt.Errorf("rl run error: %w", err)
		}

		returns, err = rl.RunEpisode(RLRunDown)
		if err != nil && !errors.Is(EndOfEpisode, err) {
			return fmt.Errorf("rl run error: %w", err)
		}
		fmt.Println("returnsDown:", returns)

		// Save
		if rl.agentSaveFreq == -1 || episode%rl.agentSaveFreq == 0 {
			if err := rl.agentUp.Save(rl.agentUpDataPath); err != nil {
				return fmt.Errorf("rl run error: %w", err)
			}
			if err := rl.agentDown.Save(rl.agentDownDataPath); err != nil {
				return fmt.Errorf("rl run error: %w", err)
			}
		}
	}

	return nil
}

func (rl *RL) RunEpisode(mode int) (returns float64, err error) {
	var ag agent.Agent
	var maxStep int
	var rewardFunc func(s []float64) float64
	switch mode {
	case RLRunUp:
		ag = rl.agentUp
		maxStep = rl.maxStepUp
		rewardFunc = rl.rewardFuncUp
	case RLRunDown:
		ag = rl.agentDown
		maxStep = rl.maxStepDown
		rewardFunc = rl.rewardFuncDown
	}

	var s1, s2, a1, a2 []float64
	s1, err = rl.env.State()
	if err != nil {
		return
	}
	r := rewardFunc(s1)
	a1 = ag.Action(s1)

	returns += r

	for step := 0; step == -1 || step < maxStep; step++ {
		if err = rl.env.RunStep(a1); err != nil {
			return
		}

		s2, err = rl.env.State()
		if err != nil {
			return
		}

		r = rewardFunc(s2)
		returns += r

		a2 = ag.Action(s2)

		ag.Learn(s1, a1, r, s2, a2)

		var isFinish bool
		isFinish, err = rl.env.IsFinish(s2)
		if err != nil {
			return
		} else if isFinish {
			err = EndOfEpisode
			return
		}

		s1 = s2
		a1 = a2
	}

	return
}
