package main

import (
	"fmt"
	"log"
	"time"

	environ "github.com/high-moctane/lab_scup2020/environment"
	"github.com/tarm/serial"
)

func main() {
	c := &serial.Config{
		Name: "/dev/ttyAMA0",
		Baud: 57600,
	}

	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	for {
		time.Sleep(500 * time.Millisecond)

		n, err := s.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("send", n)

		buf := make([]byte, 14)
		n, err = s.Read(buf)
		if n != 14 {
			continue
		}
		if err != nil {
			log.Fatal(err)
		}
		envData, err := environ.NewRRPEncodedReceiveData(buf)
		if err != nil {
			log.Fatal(err)
		}
		rcvData, err := envData.ToRRPReceiveData()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("rx %v: %v\n", n, rcvData)
		fmt.Printf("rx %v: %v\n", n, rcvData.ToRRPState())
	}
}
