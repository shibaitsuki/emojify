package color

import (
	"fmt"
	"math/rand"
	"regexp"
)

// 定数
const (
	COLOR_RED    = "#ff0000"
	COLOR_ORANGE = "#ff6600"
	COLOR_YELLOW = "#ffff00"
	COLOR_GREEN  = "#72e572"
	COLOR_CYAN   = "#72e5e5"
	COLOR_BLUE   = "#4169e1"
	COLOR_PURPLE = "#e572e5"
	COLOR_PINK   = "#ff1493"
	COLOR_WHITE  = "#ffffff"
	COLOR_GRAY   = "#a0a0a0"
	COLOR_BLACK  = "#333333"
	COLOR_BROWN  = "#a0522d"
)

var hexPattern = regexp.MustCompile(`^#?[0-9a-fA-F]{6}$`)

type ColorService struct {
}

func NewColorService() *ColorService {
	return &ColorService{}
}

func (s *ColorService) ConvHexColor(colorText string) (string, error) {
	switch colorText {
	case "red":
		return COLOR_RED, nil
	case "orange":
		return COLOR_ORANGE, nil
	case "yellow":
		return COLOR_YELLOW, nil
	case "green":
		return COLOR_GREEN, nil
	case "cyan":
		return COLOR_CYAN, nil
	case "blue":
		return COLOR_BLUE, nil
	case "purple":
		return COLOR_PURPLE, nil
	case "pink":
		return COLOR_PINK, nil
	case "white":
		return COLOR_WHITE, nil
	case "gray", "grey":
		return COLOR_GRAY, nil
	case "black":
		return COLOR_BLACK, nil
	case "brown":
		return COLOR_BROWN, nil
	}

	// HEX直接入力
	if hexPattern.MatchString(colorText) {
		if colorText[0] != '#' {
			return "#" + colorText, nil
		}
		return colorText, nil
	}

	return "", fmt.Errorf("invalid color: %s", colorText)
}

func (s *ColorService) GetRandomColor() string {
	colors := []string{
		COLOR_RED,
		COLOR_ORANGE,
		COLOR_YELLOW,
		COLOR_GREEN,
		COLOR_CYAN,
		COLOR_BLUE,
		COLOR_PURPLE,
	}

	// random
	color := colors[rand.Intn(len(colors))]

	return color
}
