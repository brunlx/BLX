package main

import (
	"image"
	"image/color"
	"testing"
)

func TestSampleBilinearDegenerateSource(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.SetRGBA(0, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255})

	got := sampleBilinear(src, 0.5, 0.5)
	want := color.RGBA{R: 10, G: 20, B: 30, A: 255}
	if got != want {
		t.Fatalf("imagem 1x1 deveria retornar o pixel único, obtive %+v", got)
	}

	if got := sampleBilinear(src, -5, 3); got != want {
		t.Fatalf("coordenadas fora dos limites não deveriam panificar, obtive %+v", got)
	}
}

func TestResizeDegenerateSource(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.SetRGBA(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	dst := resize(src, 4)
	if dst.Bounds().Dx() != 4 || dst.Bounds().Dy() != 4 {
		t.Fatalf("esperava 4x4, obtive %v", dst.Bounds())
	}
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			if got := dst.RGBAAt(x, y); got.R != 255 {
				t.Errorf("pixel (%d,%d) inesperado: %+v", x, y, got)
			}
		}
	}
}
