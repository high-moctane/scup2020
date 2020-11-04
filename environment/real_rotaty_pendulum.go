package environment

import (
	"bytes"
	"encoding/binary"
	"fmt"
	_ "github.com/tarm/serial"
	"math"
)

const RRPResetInput = 0.25

// type RealRotatyPendulum struct {
// 	seri *serial.Port
//
// 	s, sPrev RRPReceiveData
// }

// func (rrp *RealRotatyPendulum) Init() error {
// 	serialConf := serial.Config{
// 		Name: "/dev/ttyAMA0",
// 		Baud: 57600,
// 	}
// 	seri, err := serial.OpenPort(&serialConf)
// 	if err != nil {
// 		return fmt.Errorf("cannot init real rotaty pendulum: %w", err)
// 	}
//
// 	rrp.seri = seri
//
// 	return nil
// }

// func (rrp *RealRotatyPendulum) Reset() error {}

// func (rrp *RealRotatyPendulum) State() (s []float64, err error) {}

// func (rrp *RealRotatyPendulum) RunStep(a []float64) error {}

// func (rrp *RealRotatyPendulum) IsFinish(s []float64) (bool, error) {}

// func (rrp *RealRotatyPendulum) RewardFuncUp() func(s []float64) float64 {}

// func (rrp *RealRotatyPendulum) RewardFuncDown() func(s []float64) float64 {}

const RRPMaxEncoder = 200000
const RRPMaxPotentiomater = 1024
const RRPMaxPWMDuty = 12500
const RRPMaxPWMVoltage = 5

type RRPState struct {
	TimeStamp     uint32
	BaseAngle     float64
	PendulumAngle float64
	PWMVoltage    float64
}

const RRPEncodedReceiveDataLen = 14

type RRPEncodedReceiveData struct {
	TimeStamp     [4]byte
	BaseAngle     [3]byte
	PendulumAngle [2]byte
	PWMDuty       [3]byte
	CheckSum      byte
	Terminator    byte
}

func NewRRPEncodedReceiveData(buf []byte) (*RRPEncodedReceiveData, error) {
	if len(buf) != RRPEncodedReceiveDataLen {
		return nil, fmt.Errorf("invalid RRPEncodedReceiveDataLen (%d)", len(buf))
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

	res := RRPEncodedReceiveData{
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

func (erd *RRPEncodedReceiveData) ToRRPReceiveData() (*RRPReceiveData, error) {
	timeStamp, err := erd.decode(erd.TimeStamp[:])
	if err != nil {
		return nil, fmt.Errorf("cannot convert timeStamp %v to RRPReceiveData", erd)
	}

	baseAngle, err := erd.decode(erd.BaseAngle[:])
	if err != nil {
		return nil, fmt.Errorf("cannot convert baseAngle %v to RRPReceiveData", erd)
	}

	pendulumAngle, err := erd.decode(erd.PendulumAngle[:])
	if err != nil {
		return nil, fmt.Errorf("cannot convert pendulumAngle %v to RRPReceiveData", erd)
	}

	pwmDuty, err := erd.decode(erd.PWMDuty[:])
	if err != nil {
		return nil, fmt.Errorf("cannot convert pwmDuty %v to RRPReceiveData", erd)
	}

	res := &RRPReceiveData{timeStamp, baseAngle, pendulumAngle, pwmDuty}
	return res, nil
}

func (*RRPEncodedReceiveData) decode(data []byte) (uint32, error) {
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

type RRPReceiveData struct {
	TimeStamp, BaseAngle, PendulumAngle, PWMDuty uint32
}

func (rd *RRPReceiveData) ToRRPState() *RRPState {
	return &RRPState{
		rd.TimeStamp,
		rd.rawEncoderToRad(rd.BaseAngle),
		rd.rawPotentiomaterToRad(rd.PendulumAngle),
		rd.rawPWMDutyToVoltage(rd.PWMDuty),
	}
}

func (*RRPReceiveData) rawToSigned(raw uint32, max int64) int64 {
	var halfMax int64 = max / 2
	var raw64 int64 = int64(raw)

	mod := raw64 % halfMax

	var signed int64
	if raw64 < halfMax {
		signed = int64(mod)
	} else {
		signed = mod - int64(halfMax)
	}

	return signed
}

func (rd *RRPReceiveData) rawEncoderToRad(raw uint32) float64 {
	signed := rd.rawToSigned(raw, RRPMaxEncoder)
	return float64(signed) / float64(RRPMaxEncoder/2) * math.Pi
}

func (rd *RRPReceiveData) rawPotentiomaterToRad(raw uint32) float64 {
	signed := rd.rawToSigned(raw, RRPMaxPotentiomater)
	return float64(signed) / float64(RRPMaxPotentiomater/2) * math.Pi
}

func (rd *RRPReceiveData) rawPWMDutyToVoltage(raw uint32) float64 {
	var sign float64
	if raw>>16&1 == 0 {
		sign = 1.
	} else {
		sign = -1.
	}
	raw = raw & 0xffff

	return float64(sign) * float64(RRPMaxPWMDuty-raw) / RRPMaxPWMDuty * RRPMaxPWMVoltage
}

type RRPSendData struct {
	motor float64
}

func NewSendData(motor float64) *RRPSendData {
	return &RRPSendData{motor}
}

func (sd *RRPSendData) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, sd.motor)
	return buf.Bytes()
}
