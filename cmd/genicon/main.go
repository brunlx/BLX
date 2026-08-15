// Gera o ícone do BLX (ICO) a partir de assets/blx-logo.png e as
// imagens do wizard de instalação usadas pelo Inno Setup.
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// version é injetada no build via -ldflags "-X main.version=<versão>".
var version = "dev"

// iconSource é o arquivo da logo oficial do app, usado como fonte do icon.ico.
const iconSource = "assets/blx-logo.png"

type vec struct{ x, y float64 }

func v2(x, y float64) vec { return vec{x, y} }

var (
	bgTop    = color.RGBA{0x0e, 0x17, 0x2b, 0xff}
	bgBottom = color.RGBA{0x0b, 0x11, 0x20, 0xff}
	borderC  = color.RGBA{0x22, 0xd3, 0xee, 0xff}
	shieldC  = color.RGBA{0x22, 0xd3, 0xee, 0xff}
	checkC   = color.RGBA{0x10, 0xb9, 0x81, 0xff}
	gridC    = color.RGBA{0x22, 0xd3, 0xee, 0x16}
)

func main() {
	printVersion := flag.Bool("version", false, "exibe a versão e sai")
	flag.Parse()
	if *printVersion {
		fmt.Println(version)
		return
	}

	src, err := loadSource(iconSource)
	if err != nil {
		panic(err)
	}

	sizes := []int{16, 32, 48, 256}
	var pngs [][]byte

	for _, s := range sizes {
		var buf bytes.Buffer
		if err := png.Encode(&buf, resize(src, s)); err != nil {
			panic(err)
		}
		pngs = append(pngs, buf.Bytes())
	}

	if err := writeICO("assets/icon.ico", sizes, pngs); err != nil {
		panic(err)
	}

	if err := writePNG("assets/wizard-large.png", renderWizard(164, 314, 82, 120, 74)); err != nil {
		panic(err)
	}
	if err := writePNG("assets/wizard-small.png", renderWizard(55, 58, 27, 30, 30)); err != nil {
		panic(err)
	}

	fmt.Println("ícone (a partir de", iconSource, ") e wizard gerados em assets/")
}

// loadSource carrega a logo oficial do app.
func loadSource(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

// renderWizard desenha o painel do instalador: gradiente, grade sutil,
// brilho e o escudo. w/h definem o tamanho final (já em resolução nativa).
func renderWizard(w, h int, scx, scy, sw float64) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fh := float64(h)

	for y := 0; y < h; y++ {
		t := float64(y) / fh
		c := lerp(bgTop, bgBottom, t)
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x%16 == 0 || y%16 == 0 {
				blendAt(img, x, y, gridC)
			}
		}
	}

	glow(img, scx, scy, sw*1.05, shieldC)
	drawShield(img, scx, scy, sw)

	if w >= 100 {
		for x := 0; x < w; x++ {
			blendAt(img, x, h-24, borderC)
			blendAt(img, x, h-22, shieldC)
		}
	}
	return img
}

// drawShield desenha o escudo + check, centralizado em (cx, cy) com largura sw.
func drawShield(img *image.RGBA, cx, cy, sw float64) {
	sh := sw * 0.78 / 0.56
	left, top := cx-sw/2, cy-sh/2

	poly := []vec{
		v2(0.28, 0.10), v2(0.72, 0.10), v2(0.78, 0.30), v2(0.50, 0.88), v2(0.22, 0.30),
	}
	checkSegs := [][2]vec{
		{v2(0.36, 0.50), v2(0.46, 0.60)},
		{v2(0.46, 0.60), v2(0.66, 0.40)},
	}
	checkW := 0.05 * sw

	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			nx := (float64(x) - left) / sw
			ny := (float64(y) - top) / sh
			if inPoly(nx, ny, poly) {
				img.Set(x, y, shieldC)
			}
			for _, seg := range checkSegs {
				if distToSeg(float64(x), float64(y),
					left+seg[0].x*sw, top+seg[0].y*sh,
					left+seg[1].x*sw, top+seg[1].y*sh) <= checkW {
					img.Set(x, y, checkC)
				}
			}
		}
	}
}

// glow adiciona um brilho radial suave ao redor de (cx, cy).
func glow(img *image.RGBA, cx, cy, radius float64, c color.RGBA) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			d := math.Hypot(float64(x)-cx, float64(y)-cy)
			if d < radius {
				a := (1 - d/radius) * 0.22
				src := color.RGBA{c.R, c.G, c.B, uint8(a * 255)}
				blendAt(img, x, y, src)
			}
		}
	}
}

func blendAt(img *image.RGBA, x, y int, src color.RGBA) {
	dst := img.RGBAAt(x, y)
	a := float64(src.A) / 255
	dst.R = uint8(float64(dst.R)*(1-a) + float64(src.R)*a)
	dst.G = uint8(float64(dst.G)*(1-a) + float64(src.G)*a)
	dst.B = uint8(float64(dst.B)*(1-a) + float64(src.B)*a)
	dst.A = 0xff
	img.SetRGBA(x, y, dst)
}

func lerp(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 0xff,
	}
}

func inPoly(x, y float64, poly []vec) bool {
	inside := false
	for i, j := 0, len(poly)-1; i < len(poly); j, i = i, i+1 {
		xi, yi := poly[i].x, poly[i].y
		xj, yj := poly[j].x, poly[j].y
		if (yi > y) != (yj > y) && x < (xj-xi)*(y-yi)/(yj-yi)+xi {
			inside = !inside
		}
	}
	return inside
}

func distToSeg(px, py, ax, ay, bx, by float64) float64 {
	abx, aby := bx-ax, by-ay
	apx, apy := px-ax, py-ay
	len2 := abx*abx + aby*aby
	t := (apx*abx + apy*aby) / len2
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	qx, qy := ax+t*abx, ay+t*aby
	return math.Hypot(px-qx, py-qy)
}

// resize redimensiona a imagem de origem para n x n usando bilinear.
func resize(src image.Image, n int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, n, n))
	b := src.Bounds()
	sw, sh := float64(b.Dx()), float64(b.Dy())
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			sx := (float64(x)+0.5)*sw/float64(n) - 0.5
			sy := (float64(y)+0.5)*sh/float64(n) - 0.5
			dst.SetRGBA(x, y, sampleBilinear(src, sx, sy))
		}
	}
	return dst
}

func sampleBilinear(src image.Image, x, y float64) color.RGBA {
	b := src.Bounds()
	if b.Dx() < 2 || b.Dy() < 2 {
		return color.RGBAModel.Convert(src.At(b.Min.X, b.Min.Y)).(color.RGBA)
	}
	maxX, maxY := b.Dx()-1, b.Dy()-1
	xi := int(math.Floor(x))
	yi := int(math.Floor(y))
	if xi < 0 {
		xi = 0
	}
	if yi < 0 {
		yi = 0
	}
	if xi >= maxX {
		xi = maxX - 1
	}
	if yi >= maxY {
		yi = maxY - 1
	}
	tx := x - float64(xi)
	ty := y - float64(yi)

	c00 := color.RGBAModel.Convert(src.At(xi, yi)).(color.RGBA)
	c10 := color.RGBAModel.Convert(src.At(xi+1, yi)).(color.RGBA)
	c01 := color.RGBAModel.Convert(src.At(xi, yi+1)).(color.RGBA)
	c11 := color.RGBAModel.Convert(src.At(xi+1, yi+1)).(color.RGBA)

	lerp := func(a, b, c, d float64) float64 { return a + (b-a)*tx + (c-a)*ty + (a-b-c+d)*tx*ty }
	return color.RGBA{
		R: uint8(lerp(float64(c00.R), float64(c10.R), float64(c01.R), float64(c11.R))),
		G: uint8(lerp(float64(c00.G), float64(c10.G), float64(c01.G), float64(c11.G))),
		B: uint8(lerp(float64(c00.B), float64(c10.B), float64(c01.B), float64(c11.B))),
		A: uint8(lerp(float64(c00.A), float64(c10.A), float64(c01.A), float64(c11.A))),
	}
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func writeICO(path string, sizes []int, pngs [][]byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	header := make([]byte, 6)
	binary.LittleEndian.PutUint16(header[0:], 0) // reserved
	binary.LittleEndian.PutUint16(header[2:], 1) // tipo: ícone
	binary.LittleEndian.PutUint16(header[4:], uint16(len(sizes)))
	if _, err := f.Write(header); err != nil {
		return err
	}

	offset := uint32(6 + 16*len(sizes))
	for i, s := range sizes {
		entry := make([]byte, 16)
		if s >= 256 {
			entry[0], entry[1] = 0, 0
		} else {
			entry[0], entry[1] = byte(s), byte(s)
		}
		binary.LittleEndian.PutUint16(entry[4:], 1)  // planes
		binary.LittleEndian.PutUint16(entry[6:], 32) // bitcount
		binary.LittleEndian.PutUint32(entry[8:], uint32(len(pngs[i])))
		binary.LittleEndian.PutUint32(entry[12:], offset)
		if _, err := f.Write(entry); err != nil {
			return err
		}
		offset += uint32(len(pngs[i]))
	}
	for _, p := range pngs {
		if _, err := f.Write(p); err != nil {
			return err
		}
	}
	return nil
}
