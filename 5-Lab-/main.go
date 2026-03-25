package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

type GaussianFilter struct {
	Sigma  float64
	Radius int
}

func NewGaussianFilter(sigma float64) *GaussianFilter {

	r := int(math.Ceil(3 * sigma))
	return &GaussianFilter{Sigma: sigma, Radius: r}
}

func (f *GaussianFilter) Apply(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	r := f.Radius
	kernelSize := 2*r + 1

	kernel := f.generateKernel(kernelSize)

	tempW, tempH := w+2*r, h+2*r
	tempImg := image.NewRGBA(image.Rect(0, 0, tempW, tempH))

	for y := 0; y < tempH; y++ {
		for x := 0; x < tempW; x++ {
			srcX := x - r
			srcY := y - r
			if srcX < 0 {
				srcX = 0
			} else if srcX >= w {
				srcX = w - 1
			}
			if srcY < 0 {
				srcY = 0
			} else if srcY >= h {
				srcY = h - 1
			}
			tempImg.Set(x, y, src.At(srcX, srcY))
		}
	}

	result := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var rSum, gSum, bSum float64

			for ky := 0; ky < kernelSize; ky++ {
				for kx := 0; kx < kernelSize; kx++ {
					px, py := x+kx, y+ky
					pr, pg, pb, _ := tempImg.At(px, py).RGBA()

					weight := kernel[ky][kx]
					rSum += float64(pr>>8) * weight
					gSum += float64(pg>>8) * weight
					bSum += float64(pb>>8) * weight
				}
			}

			result.Set(x, y, color.RGBA{
				R: uint8(math.Min(255, math.Max(0, rSum))),
				G: uint8(math.Min(255, math.Max(0, gSum))),
				B: uint8(math.Min(255, math.Max(0, bSum))),
				A: 255,
			})
		}
	}
	return result
}

func (f *GaussianFilter) generateKernel(size int) [][]float64 {
	kernel := make([][]float64, size)
	sum := 0.0
	center := size / 2

	for i := 0; i < size; i++ {
		kernel[i] = make([]float64, size)
		for j := 0; j < size; j++ {
			y := float64(i - center)
			x := float64(j - center)
			val := (1.0 / (2 * math.Pi * f.Sigma * f.Sigma)) * math.Exp(-(x*x+y*y)/(2*f.Sigma*f.Sigma))
			kernel[i][j] = val
			sum += val
		}
	}

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			kernel[i][j] /= sum
		}
	}
	return kernel
}

func main() {
	fIn, _ := os.Open("test.png")
	defer fIn.Close()
	src, _, _ := image.Decode(fIn)

	filter := NewGaussianFilter(0.84)
	blurred := filter.Apply(src)

	fOut, _ := os.Create("test_blurred.png")
	defer fOut.Close()
	png.Encode(fOut, blurred)

	log.Println("Сглаживание по Гауссу завершено успешно.")
}
