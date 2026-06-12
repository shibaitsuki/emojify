package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session *discordgo.Session
	AppId   string
}

func NewBot(session *discordgo.Session, AppId string) *Bot {
	return &Bot{Session: session, AppId: AppId}
}

func (b *Bot) Setup() {
	// slash-command

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "ping pong",
		},
		{
			Name:        "help",
			Description: "emojify の使い方を表示します",
		},
		{
			Name:        "emojify",
			Description: "テキストからカスタム絵文字を作成します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "text",
					Description: "e.g. Thank_You, 気に_なる",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "e.g. thx, kininaru",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "color",
					Description: "文字色 / 色名 or HEXコード (例: red, #ff0000) (default: red)",
					Required:    false,
				},
			},
		},
		{
			Name:        "emojilist",
			Description: "サーバーの絵文字一覧を表示します",
		},
		{
			Name:        "emojidelete",
			Description: "絵文字を削除します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "削除する絵文字の名前 e.g. thx",
					Required:    true,
				},
			},
		},
	}

	// slash-command の一括登録
	// guildID が空文字列の場合はグローバルコマンドとしてすべての"サーバ"に適用
	_, err := b.Session.ApplicationCommandBulkOverwrite(b.AppId, "", commands)
	if err != nil {
		fmt.Println("error creating slash-commands: ", err)
		panic(err)
	}
	fmt.Println("slash-commands registered")

	// -

	// interaction-handler

	// interaction に対する応答を処理するハンドラを登録
	// handleSlashCommandInteraction をイベントリスナーとして登録
	b.Session.AddHandler(handleSlashCommandInteraction)

}

func handleSlashCommandInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		commandHandler(s, i)
	case discordgo.InteractionMessageComponent:
		componentHandler(s, i)
	}
}
