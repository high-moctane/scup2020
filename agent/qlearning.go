package agent

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	utils "github.com/high-moctane/lab_scup2020/utils"
)

func makeQTable(stateSize, actionSize int) ([][]float64, error) {
	initQ, err := utils.GetEnvFloat64("SCUP_AGENT_INIT_QVALUE")
	if err != nil {
		return nil, fmt.Errorf("cannot init qtable: %w", err)
	}

	res := make([][]float64, stateSize)
	for i := 0; i < stateSize; i++ {
		res[i] = make([]float64, actionSize)
		for j := 0; j < actionSize; j++ {
			res[i][j] = initQ
		}
	}

	return res, nil
}

func parseStateThresh() ([][]float64, error) {
	str, ok := os.LookupEnv("SCUP_AGENT_STATE_THRESH")
	if !ok {
		return nil, fmt.Errorf("cannot find SCUP_AGENT_STATE_THRESH")
	}

	res := [][]float64{}
	var err error

	for i, threshstr := range strings.Split(str, ":") {
		res = append(res, make([]float64, 2))

		threshs := strings.Split(threshstr, ",")
		if len(threshs) != 2 {
			return nil, fmt.Errorf("invalid format state thresh")
		}

		for j := 0; j < 2; j++ {
			res[i][j], err = strconv.ParseFloat(threshs[j], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid state thresh value: %w", err)
			}
		}
	}

	return res, nil
}

func parseStateNumber() ([]int, error) {
	str, ok := os.LookupEnv("SCUP_AGENT_STATE_NUMBER")
	if !ok {
		return nil, fmt.Errorf("cannot find SCUP_AGENT_STATE_NUMBER")
	}

	res := []int{}

	for _, elem := range strings.Split(str, ":") {
		v, err := strconv.Atoi(elem)
		if err != nil {
			return nil, fmt.Errorf("invalid state number value: %w", err)
		}

		res = append(res, v)
	}

	return res, nil
}

func parseActions() ([][]float64, error) {
	str, ok := os.LookupEnv("SCUP_AGENT_ACTION")
	if !ok {
		return nil, fmt.Errorf("cannot find SCUP_AGENT_ACTION")
	}

	res := [][]float64{}

	for i, actionsStr := range strings.Split(str, ":") {
		res = append(res, []float64{})

		for _, aStr := range strings.Split(actionsStr, ",") {
			a, err := strconv.ParseFloat(aStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid action value: %w", err)
			}

			res[i] = append(res[i], a)
		}
	}

	if len(res) == 0 || len(res[0]) == 0 {
		return nil, fmt.Errorf("empty actions")
	}

	for i := 0; i < len(res); i++ {
		if len(res[0]) != len(res[i]) {
			return nil, fmt.Errorf("actions vector demention error")
		}
	}

	return res, nil
}

func getStateIndex(thresh [][]float64, number []int, s []float64) int {
	indices := []int{}
	for i, v := range s {
		indices = append(indices, digitize(v, thresh[i], number[i]))
	}

	res := indices[0]
	for i := 1; i < len(s); i++ {
		res = res*number[i] + indices[i]
	}

	return res
}

func digitize(val float64, thresh []float64, number int) int {
	minThresh := thresh[0]
	maxThresh := thresh[1]

	if val < minThresh {
		return 0
	} else if val >= maxThresh {
		return number - 1
	}

	width := (maxThresh - minThresh) / float64(number-2)
	return int((val-minThresh)/width) + 1
}

func argmax(values []float64) int {
	res := 0
	for i := 1; i < len(values); i++ {
		if values[res] < values[i] {
			res = i
		}
	}
	return res
}

func encodeFloat64Slice(slice []float64) string {
	return fmt.Sprintf("%v", slice)
}

type QLearning struct {
	alpha, gamma, eps float64

	stateSize   int
	stateThresh [][]float64 // [[min, max], [min, max], ...]
	stateNumber []int

	actionSize     int
	actions        [][]float64
	actionsIndices map[string]int

	qtable [][]float64

	rewardFunc func(s []float64) float64
}

func (ql *QLearning) Init(rewardFunc func(s []float64) float64) error {
	if err := ql.loadEnv(); err != nil {
		return fmt.Errorf("cannot init qlearning: %w", err)
	}

	qtable, err := makeQTable(ql.stateSize, ql.actionSize)
	if err != nil {
		return fmt.Errorf("cannot init qlearning: %w", err)
	}
	ql.qtable = qtable

	ql.rewardFunc = rewardFunc

	return nil
}

func (ql *QLearning) Reset() {}

func (ql *QLearning) Action(s []float64) []float64 {
	var idx int
	if rand.Float64() < ql.eps {
		idx = rand.Intn(ql.actionSize)
	} else {
		sIdx := getStateIndex(ql.stateThresh, ql.stateNumber, s)
		idx = argmax(ql.qtable[sIdx])
	}

	return ql.actions[idx]
}

func (ql *QLearning) Reward(s []float64) float64 {
	return ql.rewardFunc(s)
}

func (ql *QLearning) Learn(s1, a1 []float64, r float64, s2, a2 []float64) {
	alpha := ql.alpha
	gamma := ql.gamma

	s1Idx := getStateIndex(ql.stateThresh, ql.stateNumber, s1)
	s2Idx := getStateIndex(ql.stateThresh, ql.stateNumber, s2)

	a1Idx := ql.actionsIndices[encodeFloat64Slice(a1)]

	max := ql.qtable[s2Idx][0]
	for i := 1; i < ql.actionSize; i++ {
		if max < ql.qtable[s2Idx][i] {
			max = ql.qtable[s2Idx][i]
		}
	}

	ql.qtable[s1Idx][a1Idx] =
		(1.-alpha)*ql.qtable[s1Idx][a1Idx] + alpha*(r+gamma*max)
}

func (ql *QLearning) loadEnv() error {
	if err := ql.loadParamsEnv(); err != nil {
		return fmt.Errorf("cannot load env: %w", err)
	}
	if err := ql.loadStateEnv(); err != nil {
		return fmt.Errorf("cannot load env: %w", err)
	}
	if err := ql.loadActionEnv(); err != nil {
		return fmt.Errorf("cannot load env: %w", err)
	}
	return nil
}

func (ql *QLearning) loadParamsEnv() error {
	alpha, err := utils.GetEnvFloat64("SCUP_AGENT_ALPHA")
	if err != nil {
		return fmt.Errorf("cannot load params env: %w", err)
	}

	gamma, err := utils.GetEnvFloat64("SCUP_AGENT_GAMMA")
	if err != nil {
		return fmt.Errorf("cannot load params env: %w", err)
	}

	eps, err := utils.GetEnvFloat64("SCUP_AGENT_EPSILON")
	if err != nil {
		return fmt.Errorf("cannot load params env: %w", err)
	}

	ql.alpha = alpha
	ql.gamma = gamma
	ql.eps = eps

	return nil
}

func (ql *QLearning) loadStateEnv() error {
	stateThresh, err := parseStateThresh()
	if err != nil {
		return fmt.Errorf("cannot load state env: %w", err)
	}

	stateNumber, err := parseStateNumber()
	if err != nil {
		return fmt.Errorf("cannot load state env: %w", err)
	}

	if len(stateThresh) != len(stateNumber) {
		return fmt.Errorf("len(stateThresh) == %v, but len(stateNumber) == %v",
			len(stateThresh), len(stateNumber))
	}

	stateSize := 1
	for _, n := range stateNumber {
		stateSize *= n
	}

	ql.stateSize = stateSize
	ql.stateThresh = stateThresh
	ql.stateNumber = stateNumber

	return nil
}

func (ql *QLearning) loadActionEnv() error {
	actions, err := parseActions()
	if err != nil {
		return fmt.Errorf("cannot load actions: %w", err)
	}

	actionsIndices := map[string]int{}
	for i, a := range actions {
		aEnc := encodeFloat64Slice(a)
		if _, ok := actionsIndices[aEnc]; ok {
			return fmt.Errorf("duplicate action: %v", a)
		}
		actionsIndices[aEnc] = i
	}

	ql.actionSize = len(actions)
	ql.actions = actions
	ql.actionsIndices = actionsIndices

	return nil
}
