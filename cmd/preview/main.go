package main

import (
	"fmt"
	"io"
	"os"

	color "emojify/src/domain/Color"
	textemoji "emojify/src/domain/TextEmoji"
)

const FONT_PATH = "public/fonts/MPLUSRounded1c-Bold.ttf"
const OUTPUT_PATH = "preview.png"

func main() {
	text := "Hello"
	colorInput := "red"

	if len(os.Args) >= 2 {
		text = os.Args[1]
	}
	if len(os.Args) >= 3 {
		colorInput = os.Args[2]
	}

	col := color.NewColorService()
	hexColor, err := col.ConvHexColor(colorInput)
	if err != nil {
		fmt.Println("エラー: 無効な色です:", colorInput)
		fmt.Println("色名 (red, pink, #ff0000 など) を指定してください")
		os.Exit(1)
	}

	if _, err := os.Stat(FONT_PATH); os.IsNotExist(err) {
		fmt.Println("エラー: フォントファイルが見つかりません:", FONT_PATH)
		fmt.Println("プロジェクトルート（emojify/）から実行してください")
		os.Exit(1)
	}

	te := textemoji.NewTextEmojiService(FONT_PATH)
	filePath, err := te.GenerateTextEmoji(text, hexColor)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer os.Remove(filePath)

	if err := copyFile(filePath, OUTPUT_PATH); err != nil {
		fmt.Println("Error saving preview:", err)
		os.Exit(1)
	}

	fmt.Printf("preview.png を生成しました (text=%s color=%s -> %s)\n", text, colorInput, hexColor)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
