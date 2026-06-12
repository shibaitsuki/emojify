# usage:
# - `docker build . -t emojify`
# - `docker run emojify`

# Goのビルド環境
FROM golang:1.26-alpine AS builder

# 作業ディレクトリを設定
WORKDIR /app

# 依存関係ファイルをコピー
COPY go.mod ./
COPY go.sum ./

# 依存関係をインストール
RUN go mod download

# プロジェクトの全ファイルをコピー
COPY . .

# プロジェクトルートのmain.goをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o /discord-bot .

# 実行用の新しいステージ
FROM alpine:3.21
RUN apk add --no-cache ca-certificates

# 必要なファイルやディレクトリを新しいイメージにコピー
COPY --from=builder /discord-bot /discord-bot
COPY --from=builder /app/public /public
# .env はイメージに含めない。実行時に注入する:
# docker run --env-file .env emojify

# アプリケーションがリッスンするポート番号を指定
EXPOSE 8080

# アプリケーションの起動
CMD ["/discord-bot"]