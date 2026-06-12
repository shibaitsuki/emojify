package textemoji_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	textemoji "emojify/src/domain/TextEmoji"
)

const (
	FONT_PATH = "../../../public/fonts/ZenMaruGothic-Medium.ttf"
)

type noopRepository struct{}

func (r *noopRepository) UploadToBucket(_ context.Context, _ string) (string, error) {
	return "", nil
}

func TestGenerateTextEmoji(t *testing.T) {
	service := textemoji.NewTextEmojiService(FONT_PATH, &noopRepository{})

	text := "はやく_これに_なりたい"
	hexColor := "#FF5733"

	fileName, _, err := service.GenerateTextEmoji(text, hexColor, true)
	if err != nil {
		log.Fatalf("Failed to generate text emoji: %v", err)
	}

	fmt.Printf("Generated text emoji: %s\n", fileName)
}
