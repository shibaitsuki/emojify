package textemoji

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const CANVAS_HEIGHT = 128
const CANVAS_WIDTH = 128

type TextEmojiService struct {
	fontPath string
}

func NewTextEmojiService(fontPath string) *TextEmojiService {
	return &TextEmojiService{fontPath: fontPath}
}

func (s *TextEmojiService) GenerateTextEmoji(text string, hexColor string) (string, error) {
	fontBytes, err := os.ReadFile(s.fontPath)
	if err != nil {
		return "", err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return "", err
	}

	col, err := parseHexColor(hexColor)
	if err != nil {
		return "", err
	}

	lines := strings.Split(text, "_")
	if len(lines) == 0 {
		return "", fmt.Errorf("too few lines")
	}
	if len(lines) >= 4 {
		return "", fmt.Errorf("too many lines")
	}

	// 2倍解像度で描画してから128x128に縮小（スーパーサンプリング）
	const scale = 2
	canvasW := CANVAS_WIDTH * scale
	canvasH := CANVAS_HEIGHT * scale
	cellHeight := canvasH / len(lines)

	img := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))
	draw.Draw(img, img.Bounds(), image.Transparent, image.Point{}, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(col))

	for i, line := range lines {
		fontSize := float64(cellHeight)
		face := truetype.NewFace(f, &truetype.Options{Size: fontSize})
		metrics := face.Metrics()
		ascent := int(metrics.Ascent >> 6)

		// テキスト幅が canvasW を超える場合はフォントサイズを縮小
		txtWidth := measureString(face, line).Round()
		if txtWidth > canvasW {
			fontSize *= float64(canvasW) / float64(txtWidth)
			face = truetype.NewFace(f, &truetype.Options{Size: fontSize})
			metrics = face.Metrics()
			ascent = int(metrics.Ascent >> 6)
			txtWidth = measureString(face, line).Round()
		}

		// 行をtempキャンバスに描画
		tmp := image.NewRGBA(image.Rect(0, 0, canvasW, cellHeight*2))
		draw.Draw(tmp, tmp.Bounds(), image.Transparent, image.Point{}, draw.Src)
		tmpC := freetype.NewContext()
		tmpC.SetDPI(72)
		tmpC.SetFont(f)
		tmpC.SetFontSize(fontSize)
		tmpC.SetClip(tmp.Bounds())
		tmpC.SetDst(tmp)
		tmpC.SetSrc(image.NewUniform(col))
		x := (canvasW - txtWidth) / 2
		if _, err := tmpC.DrawString(line, freetype.Pt(x, ascent)); err != nil {
			return "", err
		}

		// 実際に描画されたピクセルの範囲を取得し、cellHeightにフィット
		bounds := tightBounds(tmp)
		if bounds.Empty() {
			continue
		}
		scaleY := float64(cellHeight) / float64(bounds.Dy())
		scaledW := int(float64(bounds.Dx()) * scaleY)
		if scaledW > canvasW {
			scaledW = canvasW
		}
		lineImg := image.NewRGBA(image.Rect(0, 0, scaledW, cellHeight))
		xdraw.BiLinear.Scale(lineImg, lineImg.Bounds(), tmp, bounds, xdraw.Src, nil)

		xOff := (canvasW - scaledW) / 2
		draw.Draw(img, image.Rect(xOff, i*cellHeight, xOff+scaledW, (i+1)*cellHeight), lineImg, image.Point{}, draw.Src)
	}

	// 128x128に縮小
	out := image.NewRGBA(image.Rect(0, 0, CANVAS_WIDTH, CANVAS_HEIGHT))
	draw.Draw(out, out.Bounds(), image.Transparent, image.Point{}, draw.Src)
	xdraw.BiLinear.Scale(out, out.Bounds(), img, img.Bounds(), xdraw.Src, nil)

	outFile, err := os.CreateTemp("", "textemoji-*.png")
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	if err := png.Encode(outFile, out); err != nil {
		return "", err
	}

	return outFile.Name(), nil
}

func tightBounds(img *image.RGBA) image.Rectangle {
	b := img.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X-1, b.Min.Y-1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y).A > 0 {
				if x < minX { minX = x }
				if x > maxX { maxX = x }
				if y < minY { minY = y }
				if y > maxY { maxY = y }
			}
		}
	}
	if minX > maxX || minY > maxY {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

func parseHexColor(s string) (color.Color, error) {
	var r, g, b uint8
	_, err := fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return nil, err
	}
	return color.NRGBA{R: r, G: g, B: b, A: 0xff}, nil
}

func measureString(face font.Face, text string) (width fixed.Int26_6) {
	for _, x := range text {
		if awidth, ok := face.GlyphAdvance(rune(x)); ok {
			width += awidth
		}
	}
	return width
}
