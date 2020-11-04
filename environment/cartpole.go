package environment

import (
	"fmt"
	"math"
)

const CartpoleMaxAbsAction = 1.0
const CartpoleMaxAbsThetaDot = 10.0

type Cartpole struct {
	g, m, l, dt, ml, mass float64

	initState, s [4]float64 // [x, theta, xdot, thetadot]
}

func (cp *Cartpole) Init() error {
	cp.g = 9.80665  // 重力加速度
	cp.m = 0.1      // 棒の質量
	cp.l = 0.5      // 棒の長さ
	cp.dt = 0.02    // 制御周期
	cartMass := 1.0 // 台車の質量
	cp.ml = cp.m * cp.l
	cp.mass = cp.m + cartMass
	cp.initState = [4]float64{0., math.Pi, 0., 0.}
	cp.initState = [4]float64{0., 0., 0., 0.}
	cp.s = cp.initState
	return nil
}

func (cp *Cartpole) Reset() error {
	cp.s = cp.initState
	return nil
}

func (cp *Cartpole) State() ([]float64, error) {
	return cp.s[:], nil
}

func (cp *Cartpole) RunStep(a []float64) error {
	if len(a) != 1 {
		return fmt.Errorf("action len must be 1, but a = %v", a)
	}

	cp.s = cp.solveRungeKutta(cp.s, a[0], cp.dt)

	return nil
}

func (cp *Cartpole) isFinish(s []float64) bool {
	x := s[0]
	theta := s[1]
	thetaDot := s[3]
	return math.Abs(x) > 2. || math.Abs(theta) < math.Pi/8 && math.Abs(thetaDot) > 3*math.Pi
}

func (cp *Cartpole) IsFinish() (bool, error) {
	return cp.isFinish(cp.s[:]), nil
}

func (cp *Cartpole) RewardFuncUp() func(s []float64) float64 {
	// TODO
	// return func(s []float64) float64 {
	// 	x := s[0]
	// 	thetaDot := s[3]
	// 	if math.Abs(x) > 2. || math.Abs(thetaDot) > 2. {
	// 		return -2.
	// 	}
	// 	theta := s[1]
	// 	return -math.Abs(theta) + math.Pi/2. - 0.01*math.Abs(x)
	// }

	// return func(s []float64) float64 {
	// 	x := s[0]
	// 	thetaDot := s[3]
	// 	if math.Abs(x) > 2. || math.Abs(thetaDot) > CartpoleMaxAbsThetaDot {
	// 		return -10000000.
	// 	}
	// 	theta := s[1]
	// 	return -math.Abs(theta) + math.Pi/2. - 0.01*math.Abs(x)
	// }

	return func(s []float64) float64 {
		if cp.isFinish(s) {
			return -10000000.
		}
		x := s[0]
		theta := s[1]
		return -math.Abs(theta) + math.Pi/2. - 0.01*math.Abs(x)
	}
}

func (*Cartpole) RewardFuncDown() func(s []float64) float64 {
	// TODO
	return func(s []float64) float64 {
		x := s[0]
		thetaDot := s[3]
		if math.Abs(x) > 2. || math.Abs(thetaDot) > 4. {
			return -2.
		}
		theta := s[1]
		return math.Abs(theta) + math.Pi/2. - 0.01*math.Abs(x)
	}
}

func (cp *Cartpole) solveRungeKutta(s [4]float64, u, dt float64) [4]float64 {
	k1 := cp.differential(s, u)
	s1 := cp.solveEuler(s, k1, dt/2.)
	k2 := cp.differential(s1, u)
	s2 := cp.solveEuler(s, k2, dt/2.)
	k3 := cp.differential(s2, u)
	s3 := cp.solveEuler(s, k3, dt)
	k4 := cp.differential(s3, u)

	sNext := s
	for i := 0; i < len(s); i++ {
		sNext[i] += (k1[i] + 2.*k2[i] + 2.*k3[i] + k4[i]) * dt / 6.
	}
	sNext[1] = cp.normalize(sNext[1])

	return sNext
}

func (cp *Cartpole) differential(s [4]float64, u float64) [4]float64 {
	theta := s[1]
	xDot := s[2]
	thetaDot := s[3]

	sin := math.Sin
	cos := math.Cos

	sinTheta := sin(theta)
	cosTheta := cos(theta)

	l := cp.l
	g := cp.g
	m := cp.m
	ml := cp.ml
	mass := cp.mass

	thetaDot2 := math.Pow(thetaDot, 2.)
	cosTheta2 := math.Pow(cosTheta, 2.)

	xDDot := (4.*u/3. + 4.*ml*thetaDot2*sinTheta/3. - m*g*sin(2.*theta)/2.) /
		(4.*mass - m*cosTheta2)
	thetaDDot := (mass*g*sinTheta - ml*thetaDot2*sinTheta*cosTheta - u*cosTheta) /
		(4.*mass*l/3. - ml*cosTheta2)

	return [4]float64{xDot, thetaDot, xDDot, thetaDDot}
}

func (cp *Cartpole) solveEuler(s, sDot [4]float64, dt float64) [4]float64 {
	res := s
	for i := 0; i < len(s); i++ {
		res[i] += sDot[i] * dt
	}
	return res
}

func (*Cartpole) normalize(theta float64) float64 {
	return math.Mod(theta+3.*math.Pi, 2.*math.Pi) - math.Pi
}
