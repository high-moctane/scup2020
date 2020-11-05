package lab_scup2020

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/high-moctane/lab_scup2020/agent"
	"github.com/high-moctane/lab_scup2020/environment"
	_ "github.com/high-moctane/lab_scup2020/logger"
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
	isFinishFuncUp, isFinishFuncDown   func(s []float64) bool
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

	// IsFinishFunc
	isFinishFuncUp := env.IsFinishUp
	isFinishFuncDown := env.IsFinishDown

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

	maxStepDown, err := utils.GetEnvInt("SCUP_RL_MAX_STEP_DOWN")
	if err != nil {
		return nil, fmt.Errorf("new rl failed: %w", err)
	}

	res := &RL{
		env,
		agentUp,
		agentDown,
		rewardFuncUp,
		rewardFuncDown,
		isFinishFuncUp,
		isFinishFuncDown,
		agentUpDataPath,
		agentDownDataPath,
		agentSaveFreq,
		maxEpisode,
		maxStepUp,
		maxStepDown,
	}

	return res, nil
}

func (rl *RL) RunUpDown(ctx context.Context) error {
	for episode := 0; rl.maxEpisode == -1 || episode < rl.maxEpisode; episode++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		_, err := rl.RunEpisodeUp(ctx, episode)
		if err != nil {
			if !errors.Is(EndOfEpisode, err) {
				return fmt.Errorf("rl run error: %w", err)
			}
		}

		_, err = rl.RunEpisodeDown(ctx, episode)
		if err != nil {
			if !errors.Is(EndOfEpisode, err) {
				return fmt.Errorf("rl run error: %w", err)
			}
		}
	}

	return nil
}

func (rl *RL) RunUp(ctx context.Context) error {
	for episode := 0; rl.maxEpisode == -1 || episode < rl.maxEpisode; episode++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		_, err := rl.RunEpisodeUp(ctx, episode)
		if err != nil {

			if !errors.Is(EndOfEpisode, err) {
				return fmt.Errorf("rl run error: %w", err)
			}
		}
	}

	return nil
}

func (rl *RL) RunDown(ctx context.Context) error {
	for episode := 0; rl.maxEpisode == -1 || episode < rl.maxEpisode; episode++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		_, err := rl.RunEpisodeDown(ctx, episode)
		if err != nil {
			if !errors.Is(EndOfEpisode, err) {
				return fmt.Errorf("rl run error: %w", err)
			}
		}
	}

	return nil
}

func (rl *RL) Run(ctx context.Context, mode int) error {
	switch mode {
	case RLRunUpDown:
		return rl.RunUpDown(ctx)
	case RLRunUp:
		return rl.RunUp(ctx)
	case RLRunDown:
		return rl.RunDown(ctx)
	default:
		return fmt.Errorf("rl run error: invalid mode: %d", mode)
	}
}

func (rl *RL) RunEpisodeUp(ctx context.Context, episode int) (returns float64, err error) {
	log.Printf("up start episode %d", episode)
	returns, err = rl.RunEpisode(ctx, episode, RLRunUp)
	log.Printf("up end episode %d reward %v", episode, returns)
	return
}

func (rl *RL) RunEpisodeDown(ctx context.Context, episode int) (returns float64, err error) {
	log.Printf("down start episode %d", episode)
	returns, err = rl.RunEpisode(ctx, episode, RLRunDown)
	log.Printf("down end episode %d returns %v", episode, returns)
	return
}

func (rl *RL) RunEpisode(ctx context.Context, episode, mode int) (returns float64, err error) {
	defer rl.env.RunStep([]float64{0})

	// Reset
	if err = rl.env.Reset(); err != nil {
		err = fmt.Errorf("rl run error: %w", err)
		return
	}

	// Init
	var ag agent.Agent
	var maxStep int
	var rewardFunc func(s []float64) float64
	var isFinishFunc func(s []float64) bool
	var agentDataPath string

	switch mode {
	case RLRunUp:
		ag = rl.agentUp
		maxStep = rl.maxStepUp
		rewardFunc = rl.rewardFuncUp
		isFinishFunc = rl.isFinishFuncUp
		agentDataPath = rl.agentUpDataPath
	case RLRunDown:
		ag = rl.agentDown
		maxStep = rl.maxStepDown
		rewardFunc = rl.rewardFuncDown
		isFinishFunc = rl.isFinishFuncDown
		agentDataPath = rl.agentDownDataPath
	}

	ag.Reset()

	var s1, s2, a1, a2 []float64
	s1, err = rl.env.State()
	if err != nil {
		return
	}
	r := rewardFunc(s1)
	a1 = ag.Action(s1)

	returns += r

	// Run
	// logger.Get().Info("rl start episode %d", episode)

	var isFinish bool
	var rxError *environment.RRPSerialRxError

	for step := 0; step == -1 || step < maxStep; step++ {
		select {
		case <-ctx.Done():
			return returns, nil
		default:
		}

		log.Println(s1, a1, r, s2, a2)

		if err = rl.env.RunStep(a1); err != nil {
			if errors.As(err, &rxError) {
				continue
			}
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

		isFinish = isFinishFunc(s2)

		s1 = s2
		a1 = a2
		returns += r
	}

	if err = rl.env.RunStep([]float64{0}); err != nil {
		if errors.As(err, &rxError) {
			return
		}
		return
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
