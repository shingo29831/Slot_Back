# ベースイメージを指定
FROM golang:1.23

# 作業ディレクトリを作成
WORKDIR /app

# 作業ディレクトリにソースコードをコピー
COPY . .

ENV LOG_SERVER="logsystem:logsyspassword@tcp(mysql:3306)/log_server"

ENV ACCOUNT_SERVER="account_system:xM7B)NY-eexsJm@tcp(mysql:3306)/account_server"

# 必要な依存関係をインストール
RUN go mod tidy

# アプリケーションをビルド
RUN go build -o myapp .

# コンテナ起動時に実行されるコマンド
CMD ["/app/myapp"]
