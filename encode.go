// エンコードデコードをする
package lab_scup2020

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const EncodedReceiveDataLen = 14

type EncodedReceiveData struct {
	TimeStamp     [4]byte
	BaseAngle     [3]byte
	PendulumAngle [2]byte
	PWMDuty       [3]byte
	CheckSum      byte
	Terminator    byte
}

func NewEncodedReceiveData(buf []byte) (*EncodedReceiveData, error) {
	if len(buf) != EncodedReceiveDataLen {
		return nil, fmt.Errorf("invalid EncodedReceiveDataLen (%d)", len(buf))
	}

	data := buf[:12]
	checkSum := buf[12]
	terminator := buf[13]

	if !verifyTerminator(terminator) {
		return nil, fmt.Errorf("invalid terminator %v", buf)
	}

	if sum := calcCheckSum(data); sum != checkSum {
		return nil, fmt.Errorf("checksum expected %v, but %v: %v", checkSum, sum, buf)
	}

	res := EncodedReceiveData{
		[4]byte{buf[0], buf[1], buf[2], buf[3]},
		[3]byte{buf[4], buf[5], buf[6]},
		[2]byte{buf[7], buf[8]},
		[3]byte{buf[9], buf[10], buf[11]},
		checkSum,
		terminator,
	}

	return &res, nil
}

func verifyTerminator(terminator byte) bool {
	return terminator == '\n'
}

func calcCheckSum(data []byte) byte {
	var res uint
	for _, v := range data {
		res += uint(v)
	}
	return byte(res&0x3f + 0x30)
}

func (erd *EncodedReceiveData) ToReceiveData() (*ReceiveData, error) {
	timeStamp, err := erd.decode(erd.TimeStamp[:])
	if err != nil {
		return nil, fmt.Errorf("cannot convert timeStamp %v to ReceiveData", erd)
	}

	baseAngle, err := erd.decode(erd.BaseAngle[:])
	if err != nil {
		return nil, fmt.Errorf("cannot convert baseAngle %v to ReceiveData", erd)
	}

	pendulumAngle, err := erd.decode(erd.PendulumAngle[:])
	if err != nil {
		return nil, fmt.Errorf("cannot convert pendulumAngle %v to ReceiveData", erd)
	}

	pwmDuty, err := erd.decode(erd.PWMDuty[:])
	if err != nil {
		return nil, fmt.Errorf("cannot convert pwmDuty %v to ReceiveData", erd)
	}

	res := &ReceiveData{timeStamp, baseAngle, pendulumAngle, pwmDuty}
	return res, nil
}

func (*EncodedReceiveData) decode(data []byte) (uint32, error) {
	var res uint32

	for _, v := range data {
		if v < 0x30 {
			return 0, fmt.Errorf("value %v too small", v)
		}

		decoded := v - 0x30
		if decoded > 0x3f {
			return 0, fmt.Errorf("value %v too large", v)
		}

		res = res<<6 + uint32(v-0x30)
	}

	return res, nil
}

type ReceiveData struct {
	TimeStamp, BaseAngle, PendulumAngle, PWMDuty uint32
}

type SendData struct {
	motor float64
}

func NewSendData(motor float64) *SendData {
	return &SendData{motor}
}

func (sd *SendData) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, sd.motor)
	return buf.Bytes()
}
