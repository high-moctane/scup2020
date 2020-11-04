package main

import (
	"log"
	"os"

	scup "github.com/high-moctane/lab_scup2020"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("invalid args")
	}

	if err := godotenv.Load(os.Args[1]); err != nil {
		log.Fatal(err)
	}

	rl, err := scup.NewRL()
	if err != nil {
		log.Fatal(err)
	}

	if err := rl.RunUpDown(); err != nil {
		log.Fatal(err)
	}
}
