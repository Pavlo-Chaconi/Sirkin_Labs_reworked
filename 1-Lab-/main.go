package main

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
)

func xOft(t float64) float64 {
	return math.Cos(t) * math.Cos(math.Abs(math.Sin(t))) //Определяем сигнал согласно варианта
}

func discretizeOnePeriod(N int) (x []float64, t []float64) { //Дискретизация одного периода в N точек
	if N <= 0 {
		return
	}

	T := 2 * math.Pi
	dt := T / float64(N)
	x = make([]float64, N)
	t = make([]float64, N)
	for n := 0; n < N; n++ {
		tn := float64(n) * dt
		t[n] = tn
		x[n] = xOft(tn)
	}
	return

}

func dtfReal(x []float64) (X []complex128) { //Дискретное преобразование Фурье X[k] = Σ x[n] * exp(-j*2π*k*n/N)
	N := len(x)
	X = make([]complex128, N)
	for k := 0; k < N; k++ {
		var sum complex128 = 0
		for n := 0; n < N; n++ {
			angle := -2 * math.Pi * float64(k*n) / float64(N)
			w := cmplx.Exp(complex(0, angle))
			sum += complex(x[n], 0) * w
		}
		X[k] = sum
	}
	return
}

func idft(X []complex128) (xRec []float64) { //Обратное преобразование
	N := len(X)
	xRec = make([]float64, N)
	for n := 0; n < N; n++ {
		sum := complex(0, 0)
		for k := 0; k < N; k++ {
			angle := +2 * math.Pi * float64(k*n) / float64(N)
			w := cmplx.Exp(complex(0, angle))
			sum += X[k] * w
		}
		sum = sum / complex(float64(N), 0)
		xRec[n] = real(sum)
	}
	return
}

func minmax(a []float64) (min, max float64) {
	if len(a) == 0 {
		return 0, 0
	}
	min = a[0]
	max = a[0]
	for n := 1; n < len(a); n++ {
		v := a[n]
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return
}

func addNoise(x []float64, noiseAmp float64) (xNoisy []float64) {
	N := len(x)
	xNoisy = make([]float64, N)
	for n := 0; n < N; n++ {
		noise := (rand.Float64()*2 - 1) * noiseAmp
		xNoisy[n] = x[n] + noise
	}
	return
}

func maxAbsError(a []float64, b []float64) float64 {
	N := len(a)
	max := 0.0
	for n := 0; n < N; n++ {
		err := math.Abs(a[n] - b[n])
		if err > max {
			max = err
		}
	}
	return max
}

func maxSpectrumAmp(x []complex128) float64 {
	N := len(x)
	maxAmp := 0.0
	for k := 0; k < N; k++ {
		a := cmplx.Abs(x[k])
		if a > maxAmp {
			maxAmp = a
		}
	}
	return maxAmp
}

func thresholdFilter(x []complex128, thr float64) (Xfilt []complex128, zeroed int) {
	N := len(x)
	zeroed = 0
	Xfilt = make([]complex128, N)
	copy(Xfilt, x)
	for k := 0; k < N; k++ {
		if cmplx.Abs(Xfilt[k]) < thr {
			Xfilt[k] = 0
			zeroed++
		}
	}
	return Xfilt, zeroed
}

func main() {
	//fmt.Println(xOft(0))
	//fmt.Println(discretizeOnePeriod(32))
	// for _, N := range []int{32, 64, 128} {
	// 	x, t := discretizeOnePeriod(N)
	// 	T := 2 * math.Pi
	// 	dt := T / float64(N)
	// 	fmt.Println(N, dt)
	// 	fmt.Println(t[0], t[1], t[N-1])
	// 	fmt.Println(x[0], x[1], x[N-1])
	// }

	rand.Seed(1)
	x, _ := discretizeOnePeriod(128)
	minX, maxX := minmax(x)
	span := maxX - minX
	noiseAmp := 0.05 * span
	N = len(x)
	alpha := 0.05

	Xnoisy := addNoise(x, noiseAmp)
	maxAmp := maxSpectrumAmp(Xnoisy)

	// X := dtfReal(x)
	//N := len(x)
	//xNoisy := make([]float64, N)

	// xNoisy := addNoise(x, noiseAmp)
	Xfilt, zeroed := thresholdFilter(xNoisy, t)
	// xRec := idft(X)
	// maxErr := 0.0
	for n := 0; n < N; n++ {
		noise := (rand.Float64()*2 - 1) * noiseAmp
		xNoisy[n] = x[n] + noise
		// e := math.Abs(x[n] - xRec[n])
		// if e > maxErr {
		// 	maxErr = e
		// }
	}

	fmt.Println(minX)
	fmt.Println(maxX)
	fmt.Println(span)
	fmt.Println(noiseAmp)
	fmt.Println(xNoisy[0])
	fmt.Println(xNoisy[1])
	fmt.Println(xNoisy[2])

	Xnoisy := dtfReal(xNoisy)
	copy(Xfilt, Xnoisy)
	maxAmp := 0.0
	// fmt.Println(maxErr)
	for k := 0; k < N; k++ {
		a := cmplx.Abs(Xnoisy[k])
		if a > maxAmp {
			maxAmp = a
		}
		//amp := cmplx.Abs(X[k])
		A_norm := 2 * cmplx.Abs(Xnoisy[k]) / float64(N)
		fmt.Println(k, A_norm)
	}

	fmt.Println(maxAmp)

	thr := alpha * maxAmp

	fmt.Println("Порог равен как", thr)

	count := 0
	for k := 0; k < N; k++ {
		a := cmplx.Abs(Xfilt[k])
		if a < thr {
			Xfilt[k] = 0
			count++
		}
	}

	fmt.Println("Счетчик Zeroed", count)

	maxErrNoisy := 0.0
	for n := 0; n < N; n++ {
		e := math.Abs(x[n] - xNoisy[n])
		if e > float64(maxErrNoisy) {
			maxErrNoisy = e
		}
	}

	xFilt := idft(Xfilt)

	maxErrFilt := 0.0
	for n := 0; n < N; n++ {
		e := math.Abs(x[n] - xFilt[n])
		if e > float64(maxErrFilt) {
			maxErrFilt = e
		}

	}

}
