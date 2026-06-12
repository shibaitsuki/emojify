package discord

import (
	color "emojify/src/domain/Color"
	textemoji "emojify/src/domain/TextEmoji"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

const (
	FONT_PATH = "public/fonts/MPLUSRounded1c-Bold.ttf"
)

// discordgo v0.27.x の Button.Emoji が空でも送信されるバグを回避するカスタム型
type simpleButton struct {
	Label    string
	Style    discordgo.ButtonStyle
	CustomID string
}

func (b simpleButton) Type() discordgo.ComponentType {
	return discordgo.ButtonComponent
}

func (b simpleButton) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type     discordgo.ComponentType `json:"type"`
		Label    string                  `json:"label"`
		Style    discordgo.ButtonStyle   `json:"style"`
		CustomID string                  `json:"custom_id"`
	}{
		Type:     discordgo.ButtonComponent,
		Label:    b.Label,
		Style:    b.Style,
		CustomID: b.CustomID,
	})
}

type pendingEmoji struct {
	imageData []byte
	name      string
}

type pendingDelete struct {
	emojiID   string
	emojiName string
	guildID   string
}

var (
	pending          = map[string]*pendingEmoji{}
	pendingMu        sync.Mutex
	pendingDeletes   = map[string]*pendingDelete{}
	pendingDeletesMu sync.Mutex
)

func commandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "ping":
		responsePing(s, i)
	case "help":
		responseHelp(s, i)
	case "emojify":
		responsePreview(s, i)
	case "emojilist":
		responseList(s, i)
	case "emojidelete":
		responseDelete(s, i)
	}
}

func componentHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	switch {
	case strings.HasPrefix(customID, "emojify_confirm_"):
		handleConfirm(s, i, strings.TrimPrefix(customID, "emojify_confirm_"))
	case strings.HasPrefix(customID, "emojify_cancel_"):
		handleCancel(s, i, strings.TrimPrefix(customID, "emojify_cancel_"))
	case strings.HasPrefix(customID, "emojify_delete_confirm_"):
		handleDeleteConfirm(s, i, strings.TrimPrefix(customID, "emojify_delete_confirm_"))
	case strings.HasPrefix(customID, "emojify_delete_cancel_"):
		handleDeleteCancel(s, i, strings.TrimPrefix(customID, "emojify_delete_cancel_"))
	}
}

func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "oops...: " + message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func responseHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "emojify の使い方",
					Color:       0x1fd1da,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "/emojify",
							Value: "テキストからカスタム絵文字を作成します\n`text` : 絵文字にするテキスト。`_` で改行 (例: `Thank_You`, `気に_なる`)\n`name` : 絵文字の名前 (例: `thx`, `kininaru`)\n`color` : 文字色 (省略時: red)",
						},
						{
							Name:  "color の選択肢",
							Value: "`red` `orange` `yellow` `green` `cyan` `blue` `purple` `pink` `white` `gray` `black` `brown`\nまたは HEXコードで直接指定 (例: `#ff0000`, `ff0000`)",
						},
						{
							Name:  "/emojilist",
							Value: "サーバーの絵文字一覧と残り枠数を表示します",
						},
						{
							Name:  "/emojidelete",
							Value: "絵文字を削除します\n`name` : 削除する絵文字の名前",
						},
					},
				},
			},
		},
	})
}

func responsePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})
}

func responsePreview(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var (
		text      string
		name      string
		colorText = ""
	)
	for _, option := range i.ApplicationCommandData().Options {
		switch option.Name {
		case "text":
			text = option.StringValue()
		case "name":
			name = option.StringValue()
		case "color":
			colorText = option.StringValue()
		}
	}

	var hexColor string
	col := color.NewColorService()
	if colorText == "" {
		hexColor, _ = col.ConvHexColor("red")
	} else {
		hexColor, _ = col.ConvHexColor(colorText)
	}

	te := textemoji.NewTextEmojiService(FONT_PATH)
	filePath, err := te.GenerateTextEmoji(text, hexColor)
	if err != nil {
		fmt.Println("Failed to generate text emoji: ", err)
		respondError(s, i, "Failed to generate text emoji: "+err.Error())
		return
	}

	data, err := os.ReadFile(filePath)
	os.Remove(filePath)
	if err != nil {
		fmt.Println("Failed to read file: ", err)
		respondError(s, i, "Failed to read file")
		return
	}

	pendingMu.Lock()
	pending[i.ID] = &pendingEmoji{imageData: data, name: name}
	pendingMu.Unlock()

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Files: []*discordgo.File{
				{
					Name:        "preview.png",
					ContentType: "image/png",
					Reader:      bytes.NewReader(data),
				},
			},
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "プレビュー : " + name,
					Description: "この絵文字を追加しますか？",
					Color:       0x1fd1da,
					Image:       &discordgo.MessageEmbedImage{URL: "attachment://preview.png"},
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						simpleButton{
							Label:    "追加する",
							Style:    discordgo.SuccessButton,
							CustomID: "emojify_confirm_" + i.ID,
						},
						simpleButton{
							Label:    "キャンセル",
							Style:    discordgo.DangerButton,
							CustomID: "emojify_cancel_" + i.ID,
						},
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Println("プレビューの送信に失敗しました: ", err)
	}
}

func handleConfirm(s *discordgo.Session, i *discordgo.InteractionCreate, originalID string) {
	pendingMu.Lock()
	pe, ok := pending[originalID]
	if ok {
		delete(pending, originalID)
	}
	pendingMu.Unlock()

	if !ok {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "セッションが切れました。もう一度コマンドを実行してください。",
				Components: []discordgo.MessageComponent{},
			},
		})
		return
	}

	emoji, err := s.GuildEmojiCreate(i.GuildID, &discordgo.EmojiParams{
		Name:  pe.name,
		Image: "data:image/png;base64," + base64.StdEncoding.EncodeToString(pe.imageData),
	})
	if err != nil {
		fmt.Println("Failed to add emoji: ", err)
		msg := "絵文字の追加に失敗しました。\nhint: `MANAGE_EMOJIS_AND_STICKERS` 権限が必要です。"
		if strings.Contains(err.Error(), "30008") {
			msg = "絵文字の上限（50個）に達しています。不要な絵文字を削除するか、サーバーをブーストして上限を増やしてください。"
		}
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    msg,
				Components: []discordgo.MessageComponent{},
			},
		})
		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "<:" + emoji.Name + ":" + emoji.ID + "> `:" + emoji.Name + ":` を追加しました！",
			Components: []discordgo.MessageComponent{},
		},
	})
}

func responseList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		respondError(s, i, "サーバー情報の取得に失敗しました")
		return
	}

	emojis, err := s.GuildEmojis(i.GuildID)
	if err != nil {
		respondError(s, i, "絵文字一覧の取得に失敗しました")
		return
	}

	maxEmojis := 50
	switch guild.PremiumTier {
	case 1:
		maxEmojis = 100
	case 2:
		maxEmojis = 150
	case 3:
		maxEmojis = 250
	}

	var sb strings.Builder
	for _, e := range emojis {
		sb.WriteString(fmt.Sprintf("<:%s:%s> `:%s:`\n", e.Name, e.ID, e.Name))
	}
	description := sb.String()
	if description == "" {
		description = "絵文字がまだありません"
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       fmt.Sprintf("絵文字一覧 (%d / %d)", len(emojis), maxEmojis),
					Description: description,
					Color:       0x1fd1da,
				},
			},
		},
	})
}

func responseDelete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Options[0].StringValue()

	emojis, err := s.GuildEmojis(i.GuildID)
	if err != nil {
		respondError(s, i, "絵文字の取得に失敗しました")
		return
	}

	var target *discordgo.Emoji
	for _, e := range emojis {
		if e.Name == name {
			target = e
			break
		}
	}
	if target == nil {
		respondError(s, i, fmt.Sprintf("`:%s:` という絵文字は見つかりませんでした", name))
		return
	}

	pendingDeletesMu.Lock()
	pendingDeletes[i.ID] = &pendingDelete{emojiID: target.ID, emojiName: target.Name, guildID: i.GuildID}
	pendingDeletesMu.Unlock()

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "本当に削除しますか？",
					Description: fmt.Sprintf("<:%s:%s> `:%s:`", target.Name, target.ID, target.Name),
					Color:       0xff4444,
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						simpleButton{
							Label:    "削除する",
							Style:    discordgo.DangerButton,
							CustomID: "emojify_delete_confirm_" + i.ID,
						},
						simpleButton{
							Label:    "キャンセル",
							Style:    discordgo.SecondaryButton,
							CustomID: "emojify_delete_cancel_" + i.ID,
						},
					},
				},
			},
		},
	})
}

func handleDeleteConfirm(s *discordgo.Session, i *discordgo.InteractionCreate, originalID string) {
	pendingDeletesMu.Lock()
	pd, ok := pendingDeletes[originalID]
	if ok {
		delete(pendingDeletes, originalID)
	}
	pendingDeletesMu.Unlock()

	if !ok {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "セッションが切れました。もう一度コマンドを実行してください。",
				Components: []discordgo.MessageComponent{},
			},
		})
		return
	}

	if err := s.GuildEmojiDelete(pd.guildID, pd.emojiID); err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "絵文字の削除に失敗しました。`MANAGE_EMOJIS_AND_STICKERS` 権限が必要です。",
				Components: []discordgo.MessageComponent{},
			},
		})
		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    fmt.Sprintf("`:%s:` を削除しました。", pd.emojiName),
			Components: []discordgo.MessageComponent{},
		},
	})
}

func handleDeleteCancel(s *discordgo.Session, i *discordgo.InteractionCreate, originalID string) {
	pendingDeletesMu.Lock()
	delete(pendingDeletes, originalID)
	pendingDeletesMu.Unlock()

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "キャンセルしました。",
			Components: []discordgo.MessageComponent{},
		},
	})
}

func handleCancel(s *discordgo.Session, i *discordgo.InteractionCreate, originalID string) {
	pendingMu.Lock()
	delete(pending, originalID)
	pendingMu.Unlock()

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "キャンセルしました。",
			Components: []discordgo.MessageComponent{},
		},
	})
}
