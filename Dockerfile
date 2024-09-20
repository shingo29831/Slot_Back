# ベースイメージとしてGoを使用
FROM ubuntu:22.04 AS builder

# 作業ディレクトリを作成
WORKDIR /app

RUN apt-get update && \
        apt-get install wget -y

RUN wget https://go.dev/dl/go1.23.1.linux-amd64.tar.gz  && \
        tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz

RUN export PATH=$PATH:/usr/local/go/bin

# 必要なファイルをコンテナにコピー
COPY . /app

RUN rm -r go.mod go.sum

RUN /usr/local/go/bin/go mod init app && /usr/local/go/bin/go mod tidy 

# 静的リンクでGoバイナリをビルド
RUN /usr/local/go/bin/go build -o main .

# 最小限のAlpineイメージを使用してコンテナサイズを削減
FROM ubuntu:22.04

# 作業ディレクトリを作成
WORKDIR /app

# ビルド済みバイナリとwebディレクトリをコピー
COPY --from=builder /app/main /app/main
COPY --from=builder /app/web /app/web

# 実行権限を付与
RUN chmod +x /app/main

# アプリケーションを起動
CMD ["./main"]

