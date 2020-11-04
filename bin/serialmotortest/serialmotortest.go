package main

import (
	"fmt"
	"log"
	"time"

	scup "github.com/high-moctane/lab_scup2020"
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

	val := 0.5
	for {
		val *= -1.0

		time.Sleep(500 * time.Millisecond)

		input := scup.NewSendData(val).ToBytes()

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
		data, err := scup.NewEncodedReceiveData(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		d, err := data.ToReceiveData()
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("rx %v: %v\n", n, d)
	}
}
