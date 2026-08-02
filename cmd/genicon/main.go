// Gera os ícones do BLX (PNG/ICO) e as imagens do wizard de
// instalação usadas pelo Inno Setup.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

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
	sizes := []int{16, 32, 48, 256}
	var pngs [][]byte

	for _, s := range sizes {
		img := downsample(renderAppIcon(s*4), s)
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			panic(err)
		}
		pngs = append(pngs, buf.Bytes())
	}

	if err := writeICO("assets/icon.ico", sizes, pngs); err != nil {
		panic(err)
	}

	if err := writePNG("assets/icon.png", downsample(renderAppIcon(1024), 256)); err != nil {
		panic(err)
	}

	if err := writePNG("assets/wizard-large.png", renderWizard(164, 314, 82, 120, 74)); err != nil {
		panic(err)
	}
	if err := writePNG("assets/wizard-small.png", renderWizard(55, 58, 27, 30, 30)); err != nil {
		panic(err)
	}

	fmt.Println("ícones e wizard gerados em assets/")
}

// renderAppIcon desenha o ícone do app (fundo arredondado + escudo).
func renderAppIcon(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	s := float64(size)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			c := color.RGBA{0, 0, 0, 0}
			px, py := float64(x), float64(y)

			rr := 0.22 * s
			dx := math.Max(math.Abs(px-s/2)-(s/2-rr), 0)
			dy := math.Max(math.Abs(py-s/2)-(s/2-rr), 0)
			sdf := math.Hypot(dx, dy)

			switch {
			case sdf <= rr:
				c = bgBottom
				if sdf > rr-0.035*s {
					c = borderC
				}
			}
			img.Set(x, y, c)
		}
	}

	drawShield(img, s/2, s*0.46, s*0.62)
	return img
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

func downsample(src *image.RGBA, n int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, n, n))
	f := src.Bounds().Dx() / n
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			var r, g, b, a, count int
			for dy := 0; dy < f; dy++ {
				for dx := 0; dx < f; dx++ {
					c := src.RGBAAt(x*f+dx, y*f+dy)
					r += int(c.R)
					g += int(c.G)
					b += int(c.B)
					a += int(c.A)
					count++
				}
			}
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(r / count),
				G: uint8(g / count),
				B: uint8(b / count),
				A: uint8(a / count),
			})
		}
	}
	return dst
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
