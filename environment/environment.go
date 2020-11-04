package environment

import (
	"fmt"
	"os"
)

type Environment interface {
	Init() error
	Reset() error
	State() (s []float64, err error)
	RunStep(a []float64) (err error)
	IsFinish(s []float64) (bool, error)
	RewardFuncUp() func(s []float64) float64
	RewardFuncDown() func(s []float64) float64
}

func SelectEnvironment() (Environment, error) {
	envName, ok := os.LookupEnv("SCUP_ENV_NAME")
	if !ok {
		return nil, fmt.Errorf("cannot get SCUP_ENV_NAME")
	}

	var env Environment
	switch envName {
	case "Cartpole":
		env = new(Cartpole)
	// case "RealRotatyPendulum":
	// 	env = new(RealRotatyPendulum)
	default:
		return nil, fmt.Errorf("invalid env name")
	}

	return env, nil
}
