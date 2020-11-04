package main

import (
	"fmt"
	"os"
	"time"

	scup "github.com/high-moctane/lab_scup2020"
	"github.com/high-moctane/lab_scup2020/logger"
	"github.com/high-moctane/lab_scup2020/utils"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(os.Args); err != nil {
		logger.Get().Fatal("%v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("invalid args")
	}

	if err := godotenv.Load(args[1]); err != nil {
		return fmt.Errorf("dotenv failed: %w", err)
	}

	rl, err := scup.NewRL()
	if err != nil {
		return fmt.Errorf("run error: %w", err)
	}

	mode, err := utils.GetEnvInt("SCUP_MODE")
	if err != nil {
		return fmt.Errorf("run error: %w", err)
	}

	time.Sleep(5 * time.Second)

	if err := rl.Run(mode); err != nil {
		return fmt.Errorf("run error: %w", err)
	}

	return nil
}
