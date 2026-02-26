package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// Вариант 6: исходный сигнал x(t)
func xOfT(t float64) float64 {
	return math.Cos(t) * math.Cos(math.Abs(math.Sin(t)))
}

// Дискретизация одного периода [0, 2π] в N точек
func discretizeOnePeriod(N int) (t []float64, x []float64, dt float64) {
	if N <= 1 {
		return nil, nil, 0
	}
	T := 4 * math.Pi
	dt = T / float64(N)
	t = make([]float64, N)
	x = make([]float64, N)
	for n := 0; n < N; n++ {
		tn := float64(n) * dt
		t[n] = tn
		x[n] = xOfT(tn)
	}
	return
}

var invSqrt2Pi = 1.0 / math.Sqrt(2*math.Pi)

// WAVE-вейвлет: ψ(t) = -(1/√(2π)) * t * exp(-t²/2)
func psiWave(t float64) float64 {
	return -invSqrt2Pi * t * math.Exp(-t*t/2.0)
}

func psiSombrero(t float64) float64 {
	tt := t * t
	return -invSqrt2Pi * (tt - 1) * math.Exp(-tt/2)
}

func ensureOutDir() string {
	out := "out"
	_ = os.MkdirAll(out, 0o755)
	return out
}

func toXY(t []float64, y []float64) plotter.XYs {
	pts := make(plotter.XYs, len(t))
	for i := range t {
		pts[i].X = t[i]
		pts[i].Y = y[i]
	}
	return pts
}

func saveLinePNG(filename, title, xlabel, ylabel string, pts plotter.XYs) error {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = xlabel
	p.Y.Label.Text = ylabel
	line, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}
	p.Add(line)
	return p.Save(7*vg.Inch, 3.5*vg.Inch, filename)
}

// Численный CWT: интеграл заменяем суммой (метод прямоугольников), далее строим скалограмму.
type waveletFn func(float64) float64

func cwtScalogramAbs(t []float64, x []float64, dt float64, psi waveletFn, scales []float64) [][]float64 {
	N := len(t)
	H := len(scales)
	out := make([][]float64, H)

	for yi := 0; yi < H; yi++ {
		s := scales[yi]
		if s == 0 {
			s = 1e-9
		}
		norm := 1.0 / math.Sqrt(math.Abs(s))
		row := make([]float64, N)

		for bi := 0; bi < N; bi++ {
			b := t[bi]
			sum := 0.0
			for n := 0; n < N; n++ {
				u := (t[n] - b) / s
				sum += x[n] * psi(u)
			}
			sum *= norm * dt
			row[bi] = math.Abs(sum)
		}
		out[yi] = row
	}
	return out
}

func logspace(min, max float64, n int) []float64 {
	if n <= 1 {
		return []float64{min}
	}
	if min <= 0 || max <= 0 {
		out := make([]float64, n)
		step := (max - min) / float64(n-1)
		for i := 0; i < n; i++ {
			out[i] = min + float64(i)*step
		}
		return out
	}
	out := make([]float64, n)
	a := math.Log(min)
	b := math.Log(max)
	step := (b - a) / float64(n-1)
	for i := 0; i < n; i++ {
		out[i] = math.Exp(a + float64(i)*step)
	}
	return out
}

// Нормализация в диапазон [0..255]
func saveScalogramGrayPNG(filename string, data [][]float64) error {
	if len(data) == 0 || len(data[0]) == 0 {
		return fmt.Errorf("empty scalogram")
	}
	H := len(data)
	W := len(data[0])

	minV := data[0][0]
	maxV := data[0][0]
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			v := data[y][x]
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}
	}
	den := maxV - minV
	if den == 0 {
		den = 1
	}

	img := image.NewGray(image.Rect(0, 0, W, H))
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			u := (data[y][x] - minV) / den
			if u < 0 {
				u = 0
			}
			if u > 1 {
				u = 1
			}
			img.SetGray(x, y, color.Gray{Y: uint8(math.Round(u * 255))})
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func main() {
	outDir := ensureOutDir()

	// 1) Два графика вейвлетов ψ(t)
	{
		tMin, tMax := -5.0, 5.0
		step := 0.01
		M := int(math.Round((tMax-tMin)/step)) + 1

		tt := make([]float64, M)
		wave := make([]float64, M)
		sombrero := make([]float64, M)

		for i := 0; i < M; i++ {
			ti := tMin + float64(i)*step
			tt[i] = ti
			wave[i] = psiWave(ti)
			sombrero[i] = psiSombrero(ti)
		}

		if err := saveLinePNG(filepath.Join(outDir, "wavelet_wave.png"),
			"WAVE ψ(t)", "t", "ψ(t)", toXY(tt, wave)); err != nil {
			panic(err)
		}
		if err := saveLinePNG(filepath.Join(outDir, "wavelet_sombrero.png"),
			"SOMBRERO ψ(t) (t-1)", "t", "ψ(t)", toXY(tt, sombrero)); err != nil {
			panic(err)
		}
	}

	// 2) График исходного сигнала x(t) за один период
	N := 256
	t, x, dt := discretizeOnePeriod(N)
	{
		if err := saveLinePNG(filepath.Join(outDir, "signal_x.png"),
			fmt.Sprintf("x(t), N=%d", N), "t", "x", toXY(t, x)); err != nil {
			panic(err)
		}
	}

	// 3) Скалограмма (2D картинка) — именно по WAVE-вейвлету
	scales := logspace(0.01, 1.0, 128)
	{
		data := cwtScalogramAbs(t, x, dt, psiWave, scales)
		if err := saveScalogramGrayPNG(filepath.Join(outDir, "scalogram_wave.png"), data); err != nil {
			panic(err)
		}
	}

	fmt.Println("Рисунки сохранены в:", outDir)
}
