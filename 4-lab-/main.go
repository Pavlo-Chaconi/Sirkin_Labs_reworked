package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

func interpolate(x, y float64, img image.Image, channel string) float64 {
	x1, y1 := int(math.Floor(x)), int(math.Floor(y))
	x2, y2 := x1+1, y1+1

	bounds := img.Bounds()
	if x2 >= bounds.Max.X {
		x2 = bounds.Max.X - 1
	}
	if y2 >= bounds.Max.Y {
		y2 = bounds.Max.Y - 1
	}

	getVal := func(px, py int) float64 {
		r, g, b, _ := img.At(px, py).RGBA()
		switch channel {
		case "R":
			return float64(r >> 8)
		case "G":
			return float64(g >> 8)
		case "B":
			return float64(b >> 8)
		default:
			return 0
		}
	}

	q11 := getVal(x1, y1)
	q21 := getVal(x2, y1)
	q12 := getVal(x1, y2)
	q22 := getVal(x2, y2)

	dx := x - float64(x1)
	dy := y - float64(y1)

	val := q11*(1-dx)*(1-dy) + q21*dx*(1-dy) + q12*(1-dx)*dy + q22*dx*dy
	return val
}

func main() {
	file, err := os.Open("chessboard.png")
	if err != nil {
		log.Fatal("Нужно положить файл input.png в папку с программой")
	}
	defer file.Close()

	srcImg, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	bounds := srcImg.Bounds()
	srcW, srcH := bounds.Max.X, bounds.Max.Y

	scale := 1.8 //80 процентов
	dstW := int(float64(srcW) * scale)
	dstH := int(float64(srcH) * scale)
	dstImg := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			srcX := float64(x) / scale
			srcY := float64(y) / scale

			if srcX > float64(srcW-1) {
				srcX = float64(srcW - 1)
			}
			if srcY > float64(srcH-1) {
				srcY = float64(srcH - 1)
			}

			r := uint8(interpolate(srcX, srcY, srcImg, "R"))
			g := uint8(interpolate(srcX, srcY, srcImg, "G"))
			b := uint8(interpolate(srcX, srcY, srcImg, "B"))

			dstImg.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	outFile, _ := os.Create("output_scaled.png")
	defer outFile.Close()
	png.Encode(outFile, dstImg)

	log.Printf("Изображение масштабировано: %dx%d -> %dx%d\n", srcW, srcH, dstW, dstH)
}
