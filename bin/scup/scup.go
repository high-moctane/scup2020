package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	scup "github.com/high-moctane/lab_scup2020"
	_ "github.com/high-moctane/lab_scup2020/logger"
	"github.com/high-moctane/lab_scup2020/utils"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(os.Args); err != nil {
		// logger.Get().Fatal("%v", err)
		log.Println(err)
		os.Exit(1)
	}
}

func run(args []string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	wg := new(sync.WaitGroup)

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
	defer rl.Close()

	mode, err := utils.GetEnvInt("SCUP_MODE")
	if err != nil {
		return fmt.Errorf("run error: %w", err)
	}

	time.Sleep(2 * time.Second)

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := rl.Run(ctx, mode); err != nil {
			log.Println(fmt.Errorf("run error: %w", err))
			return
		}
	}()

	sig := make(chan os.Signal, 1)

	signal.Notify(
		sig,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	<-sig
	wg.Wait()
	log.Println("Interrupted")

	return nil
}
