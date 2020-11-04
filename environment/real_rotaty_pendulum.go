package environment

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/high-moctane/lab_scup2020/utils"
	"github.com/tarm/serial"
)

const RRPResetInput = 0.25
const RRPInitialBaseAngleRange = math.Pi / 32
const RRPMaxBaseAngleRange = math.Pi / 2
const RRPMaxTopPendulumAngleRange = math.Pi / 32
const RRPMaxTopPendulumVelocityRange = 10

type RRPSerialTxError struct {
	n int
}

func NewRRPSerialTxError(n int) *RRPSerialTxError {
	return &RRPSerialTxError{n}
}

func (e *RRPSerialTxError) Error() string {
	return fmt.Sprintf("tx data len must be %d, but %d", RRPSendDataLen, e.n)
}

type RRPSerialRxError struct {
	n int
}

func NewRRPSerialRxError(n int) *RRPSerialRxError {
	return &RRPSerialRxError{n}
}

func (e *RRPSerialRxError) Error() string {
	return fmt.Sprintf("rx data len must be %d, but %d", RRPEncodedReceiveDataLen, e.n)
}

type RealRotatyPendulum struct {
	seri *serial.Port

	dt time.Duration

	s, sPrev          *RRPState
	initPendulumAngle float64

	goodReward, badReward float64
}

func (rrp *RealRotatyPendulum) Init() error {
	serialConf := serial.Config{
		Name: "/dev/ttyAMA0",
		Baud: 57600,
	}
	seri, err := serial.OpenPort(&serialConf)
	if err != nil {
		return fmt.Errorf("cannot init real rotaty pendulum: %w", err)
	}

	dtRaw, err := utils.GetEnvInt("SCUP_RRP_DT")
	if err != nil {
		return fmt.Errorf("cannot init real rotaty pendulum: %w", err)
	}
	dt := time.Duration(dtRaw) * time.Millisecond

	goodReward, err := utils.GetEnvFloat64("SCUP_RRP_BAD_REWARD")
	if err != nil {
		return fmt.Errorf("cannot init real rotaty pendulum: %w", err)
	}

	badReward, err := utils.GetEnvFloat64("SCUP_RRP_BAD_REWARD")
	if err != nil {
		return fmt.Errorf("cannot init real rotaty pendulum: %w", err)
	}

	rrp.seri = seri
	rrp.dt = dt
	rrp.goodReward = goodReward
	rrp.badReward = badReward

	for rrp.sPrev == nil {
		rrp.RunStep([]float64{0})
	}
	rrp.initPendulumAngle = rrp.s.ToState(rrp.sPrev)[3]

	return nil
}

func (rrp *RealRotatyPendulum) Reset() error {
	var rxError *RRPSerialRxError

	for {
		if rrp.s != nil && math.Abs(rrp.s.BaseAngle) < RRPInitialBaseAngleRange {
			rrp.RunStep([]float64{0})
			return nil
		}

		direction := 1.
		if rrp.s.BaseAngle > 0 {
			direction = -1
		}

		if err := rrp.RunStep([]float64{direction * RRPResetInput}); err != nil {
			if errors.As(err, &rxError) {
				continue
			}
			return fmt.Errorf("reset error: %w", err)
		}
	}
}

func (rrp *RealRotatyPendulum) State() (s []float64, err error) {
	return rrp.s.ToState(rrp.sPrev), nil
}

func (rrp *RealRotatyPendulum) RunStep(a []float64) error {
	// TODO
	time.Sleep(20 * time.Millisecond)

	// Send
	if len(a) != 1 {
		panic(fmt.Errorf("action len must be 1, but %d", len(a)))
	}
	sendData := NewRRPSendData(a[0])

	n, err := rrp.seri.Write(sendData.ToBytes())
	if err != nil {
		return fmt.Errorf("run step error: %w", err)
	}
	if n != RRPSendDataLen {
		return NewRRPSerialTxError(n)
	}

	// Receive
	buf := make([]byte, 14)
	n, err = rrp.seri.Read(buf)
	if err != nil {
		return fmt.Errorf("run step error: %w", err)
	}
	if n != RRPEncodedReceiveDataLen {
		return NewRRPSerialRxError(n)
	}
	encData, err := NewRRPEncodedReceiveData(buf)
	if err != nil {
		return fmt.Errorf("run step error: %w", err)
	}
	rsvData, err := encData.ToRRPReceiveData()
	if err != nil {
		return fmt.Errorf("run step error: %w", err)
	}
	s := rsvData.ToRRPState()

	// Update
	rrp.s, rrp.sPrev = s, rrp.s

	return nil
}

func (rrp *RealRotatyPendulum) IsFinish(s []float64) (bool, error) {
	baseAngle := math.Abs(s[0])
	pendAngle := math.Abs(s[1])
	pendVel := math.Abs(s[3])

	res := baseAngle >= RRPMaxBaseAngleRange ||
		pendAngle < RRPMaxTopPendulumAngleRange && pendVel > RRPMaxTopPendulumVelocityRange
	return res, nil
}

func (rrp *RealRotatyPendulum) RewardFuncUp() func(s []float64) float64 {
	return func(s []float64) float64 {
		if isFinish, _ := rrp.IsFinish(s); isFinish {
			return rrp.badReward
		}
		baseAngle := math.Abs(s[0])
		relPendAngle := math.Abs(relativeAngle(rrp.initPendulumAngle, s[1]))
		return -relPendAngle + math.Pi/2. - 0.01*baseAngle - 1.0
	}
}

func (rrp *RealRotatyPendulum) RewardFuncDown() func(s []float64) float64 {
	return func(s []float64) float64 {
		if isFinish, _ := rrp.IsFinish(s); isFinish {
			return rrp.goodReward
		}
		baseAngle := math.Abs(s[0])
		relPendAngle := math.Abs(relativeAngle(rrp.initPendulumAngle, s[1]))
		return relPendAngle - math.Pi/2. - 0.01*baseAngle - 1.0
	}
}

func (rrp *RealRotatyPendulum) Close() error {
	if err := rrp.RunStep([]float64{0.}); err != nil {
		return fmt.Errorf("rrp close error: %w", err)
	}

	if err := rrp.seri.Close(); err != nil {
		return fmt.Errorf("rrp close error: %w", err)
	}

	return nil
}

func relativeAngle(theta, other float64) float64 {
	res := other - theta
	if res < -math.Pi {
		res += 2 * math.Pi
	} else if res > math.Pi {
		res -= 2 * math.Pi
	}
	return res
}

const RRPMaxEncoder = 262000
const RRPMaxPotentiomater = 1024
const RRPMaxPWMDuty = 12500
const RRPMaxPWMVoltage = 5

type RRPState struct {
	TimeStamp     uint32
	BaseAngle     float64
	PendulumAngle float64
	PWMVoltage    float64
}

func (rrps *RRPState) ToState(prev *RRPState) []float64 {
	if rrps.TimeStamp <= prev.TimeStamp {
		return nil
	}

	dt := time.Duration(rrps.TimeStamp-prev.TimeStamp) * time.Millisecond

	baseVel := rrps.velocity(rrps.BaseAngle, prev.BaseAngle, dt)
	pendulumVel := rrps.velocity(rrps.PendulumAngle, prev.PendulumAngle, dt)

	return []float64{rrps.BaseAngle, rrps.PendulumAngle, baseVel, pendulumVel}
}

func (*RRPState) velocity(cur, prev float64, dt time.Duration) float64 {
	return (cur - prev) / dt.Seconds()
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

const RRPSendDataLen = 8

type RRPSendData struct {
	motor float64
}

func NewRRPSendData(motor float64) *RRPSendData {
	return &RRPSendData{motor}
}

func (sd *RRPSendData) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, sd.motor)
	return buf.Bytes()
}
