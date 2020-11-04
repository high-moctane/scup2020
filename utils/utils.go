package utils

import (
	"fmt"
	"os"
	"strconv"
)

func GetEnvInt(env string) (int, error) {
	str, ok := os.LookupEnv(env)
	if !ok {
		return 0, fmt.Errorf("cannot get %v", env)
	}

	res, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("invalid env %v: %w", env, err)
	}

	return res, nil
}

func GetEnvFloat64(env string) (float64, error) {
	str, ok := os.LookupEnv(env)
	if !ok {
		return 0, fmt.Errorf("cannot get %v", env)
	}

	res, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid env %v: %w", env, err)
	}

	return res, nil
}
