package main

import (
	"image/color"
	"log"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func signal(t float64) float64 {
	return math.Cos(t) * math.Cos(math.Abs(math.Sin(t)))
}

func getNewTime(newIndex int, newDt float64) float64 {
	return float64(newIndex) * newDt
}

func main() {
	const N = 200
	const periods = 4.0
	T_total := periods * 2.0 * math.Pi
	dt := T_total / float64(N-1)

	origT := make([]float64, N)
	origY := make([]float64, N)
	origPts := make(plotter.XYs, N)

	for i := 0; i < N; i++ {
		origT[i] = float64(i) * dt
		origY[i] = signal(origT[i])
		origPts[i].X = origT[i]
		origPts[i].Y = origY[i]
	}

	newDt := dt / 1.2
	var resampledPts plotter.XYs
	newIndex := 0

	for i := 0; i < N-3; i++ {
		t0, t1, t2, t3 := origT[i], origT[i+1], origT[i+2], origT[i+3]
		y0, y1, y2, y3 := origY[i], origY[i+1], origY[i+2], origY[i+3]

		a3 := (y3-y0)/6.0 + (y1-y2)/2.0
		a1 := (y3-y1)/2.0 - a3
		a2 := (y3 - y2) - a1 - a3
		a0 := y2

		for {
			tTarget := getNewTime(newIndex, newDt)

			if tTarget >= t2 && i != N-4 {
				break
			}
			if i == N-4 && tTarget > t3 {
				break
			}

			if (i == 0 && tTarget >= t0) || tTarget >= t1 {
				x := (tTarget-t0)*3.0/(t3-t0) - 2.0
				newY := ((a3*x+a2)*x+a1)*x + a0
				resampledPts = append(resampledPts, plotter.XY{X: tTarget, Y: newY})
			}
			newIndex++
		}
	}

	p1 := plot.New()
	p1.Title.Text = "Шаг 2: Исходный сигнал (200 отсчетов)"
	p1.X.Label.Text = "Время t (с)"
	p1.Y.Label.Text = "Амплитуда"
	p1.Add(plotter.NewGrid())

	origLine1, _ := plotter.NewLine(origPts)
	origLine1.Color = color.RGBA{B: 200, A: 255}
	origLine1.Width = vg.Points(2)
	p1.Add(origLine1)

	if err := p1.Save(10*vg.Inch, 5*vg.Inch, "1_original_signal.png"); err != nil {
		log.Fatal(err)
	}

	p2 := plot.New()
	p2.Title.Text = "Шаг 4-9: Передискретизированный сигнал (+20%)"
	p2.X.Label.Text = "Время t (с)"
	p2.Y.Label.Text = "Амплитуда"
	p2.Add(plotter.NewGrid())

	resampledLine2, _ := plotter.NewLine(resampledPts)
	resampledLine2.Color = color.RGBA{R: 220, A: 255}
	resampledLine2.Width = vg.Points(1.5)

	resampledScatter2, _ := plotter.NewScatter(resampledPts)
	resampledScatter2.Color = color.RGBA{R: 220, A: 255}
	resampledScatter2.Radius = vg.Points(2)
	p2.Add(resampledLine2, resampledScatter2)

	if err := p2.Save(10*vg.Inch, 5*vg.Inch, "2_resampled_signal.png"); err != nil {
		log.Fatal(err)
	}

	p3 := plot.New()
	p3.Title.Text = "Сравнение: Исходный сигнал vs Ресэмплинг"
	p3.X.Label.Text = "Время t (с)"
	p3.Y.Label.Text = "Амплитуда"
	p3.Add(plotter.NewGrid())

	origLine3, _ := plotter.NewLine(origPts)
	origLine3.Color = color.RGBA{B: 200, A: 255}
	origLine3.Width = vg.Points(2)

	resampledLine3, _ := plotter.NewLine(resampledPts)
	resampledLine3.Color = color.RGBA{R: 255, A: 255}
	resampledLine3.Dashes = []vg.Length{vg.Points(4), vg.Points(4)}

	resampledScatter3, _ := plotter.NewScatter(resampledPts)
	resampledScatter3.Color = color.RGBA{R: 255, A: 255}
	resampledScatter3.Radius = vg.Points(1.5)

	p3.Add(origLine3, resampledLine3, resampledScatter3)
	p3.Legend.Add("Исходный (N=200)", origLine3)
	p3.Legend.Add("Ресэмплинг (+20%)", resampledLine3, resampledScatter3)

	if err := p3.Save(10*vg.Inch, 5*vg.Inch, "3_combined_plot.png"); err != nil {
		log.Fatal(err)
	}

	log.Println("Успешно сгенерированы 3 файла с графиками!")
}
