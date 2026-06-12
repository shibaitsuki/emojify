package main

import (
	textemoji "emojify/src/domain/TextEmoji"
	"fmt"
	"log"
)

const (
	FONT_PATH = "public/fonts/MPLUSRounded1c-Bold.ttf"
)

func main() {
	service := textemoji.NewTextEmojiService(FONT_PATH)

	text := "はやく_これに_なりたい"
	hexColor := "#FF5733"

	fileName, err := service.GenerateTextEmoji(text, hexColor)
	if err != nil {
		log.Fatalf("Failed to generate text emoji: %v", err)
	}

	fmt.Printf("Generated text emoji: %s\n", fileName)
}
