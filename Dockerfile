# ベースイメージとしてGoを使用
FROM golang:1.20 as builder

# 作業ディレクトリを作成
WORKDIR /app

# 同ディレクトリのファイルをコンテナにコピー
COPY . .

# 依存関係のインストールとコンパイル
RUN go mod tidy
RUN go build -o main .

# マルチステージビルドで軽量化
FROM golang:1.20

# MySQLクライアントのインストール（必要であれば）
RUN apt-get update && apt-get install -y default-mysql-client

# 作業ディレクトリを作成
WORKDIR /app

# ビルド済みのバイナリをコピー
COPY ./web ./web
COPY --from=builder /app/main /app/main
COPY --from=builder /app/config /app/config

# アプリケーションを起動
CMD ["./main"]
