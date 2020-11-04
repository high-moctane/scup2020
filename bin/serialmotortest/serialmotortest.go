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
	defer s.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})

	val := 0.5
	for {
		val *= -1.0

		time.Sleep(500 * time.Millisecond)

		input := environ.NewSendData(val).ToBytes()

		n, err := s.Write(input)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("send: %v, %v\n", n, input)

		buf := make([]byte, 14)
		n, err = s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		data, err := environ.NewRRPEncodedReceiveData(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		d, err := data.ToRRPReceiveData()
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("rx %v: %v\n", n, d.ToRRPState())
	}
}
