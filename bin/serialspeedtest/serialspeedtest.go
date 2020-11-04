package main

import (
	"fmt"
	"log"
	"os"
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

	input := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	receive := make([]byte, 14)

	history := []uint32{}
	txErrorIdx := []int{}
	rxErrorIdx := []int{}

	for i := 0; i < 10000; i++ {
		time.Sleep(10 * time.Millisecond)
		n, err := s.Write(input)
		if err != nil {
			log.Fatal(err)
		}
		if n != 8 {
			txErrorIdx = append(txErrorIdx, i)
			continue
		}

		n, err = s.Read(receive)
		if err != nil {
			log.Fatal(err)
		}
		if n != 14 {
			fmt.Println(n, receive)
			rxErrorIdx = append(rxErrorIdx, i)
			continue
		}
		raw, err := scup.NewEncodedReceiveData(receive)
		if err != nil {
			rxErrorIdx = append(rxErrorIdx, i)
			continue
		}
		data, err := raw.ToReceiveData()
		if err != nil {
			rxErrorIdx = append(rxErrorIdx, i)
			continue
		}
		history = append(history, data.TimeStamp)
	}

	f, err := os.Create("timestamp.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	for _, ts := range history {
		if _, err := f.WriteString(fmt.Sprintf("%v\n", ts)); err != nil {
			log.Fatal(err)
		}
	}

	g, err := os.Create("txerror.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()
	for _, idx := range txErrorIdx {
		if _, err := g.WriteString(fmt.Sprintf("%v\n", idx)); err != nil {
			log.Fatal(err)
		}
	}

	h, err := os.Create("rxerror.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()
	for _, idx := range rxErrorIdx {
		if _, err := h.WriteString(fmt.Sprintf("%v\n", idx)); err != nil {
			log.Fatal(err)
		}
	}
}
