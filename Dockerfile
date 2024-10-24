# ベースイメージを指定
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# 最新のGoバージョンをダウンロード
RUN curl -OL https://golang.org/dl/go1.23.1.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz \
    && rm go1.23.1.linux-amd64.tar.gz

# 環境変数の設定
ENV PATH="/usr/local/go/bin:${PATH}"

# 作業ディレクトリを作成
WORKDIR /app

# 作業ディレクトリにソースコードをコピー
COPY *go .

COPY ./web ./web


RUN openssl req -new -x509 -days 365 -nodes -out server.crt -keyout server.key


ENV LOG_SERVER="logsystem:logsyspassword@tcp(mysql:3306)/log_server"

ENV ACCOUNT_SERVER="account_system:xM7B)NY-eexsJm@tcp(mysql:3306)/account_server"

# 必要な依存関係をインストール
RUN go mod tidy

# アプリケーションをビルド
RUN go build -o myapp .

# コンテナ起動時に実行されるコマンド
CMD ["/app/myapp"]
