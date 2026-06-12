# emojify

テキストからDiscordのカスタム絵文字を作成するBotです。

## コマンド

| コマンド | 説明 |
|---|---|
| `/emojify` | テキストからカスタム絵文字を作成 |
| `/emojilist` | サーバーの絵文字一覧を表示 |
| `/emojidelete` | 絵文字を削除 |
| `/help` | 使い方を表示 |
| `/ping` | 疎通確認 |

### `/emojify` オプション

| オプション | 必須 | 説明 |
|---|---|---|
| `text` | ✓ | 絵文字にするテキスト。`_` で改行（例: `Thank_You`） |
| `name` | ✓ | 絵文字の名前（例: `thx`） |
| `color` | - | 文字色（デフォルト: `red`） |

**color の選択肢:**
`red` `orange` `yellow` `green` `cyan` `blue` `purple` `pink` `white` `gray` `black` `brown`
または HEXコードで直接指定（例: `#ff0000`）

## セットアップ

### 環境変数

`.env` ファイルを作成してください：

```env
DISCORD_BOT_TOKEN=your_bot_token
DISCORD_APP_ID=your_app_id
```

### 起動

```bash
docker compose up -d
```

## 必要な権限

- `applications.commands`
- `bot`
  - `Manage Emojis and Stickers`
