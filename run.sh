#!/usr/bin/bash


# systemctl start mysql #管理者権限用

export ACCOUNT_SERVER="account_system:xM7B)NY-eexsJm@tcp(localhost:3306)/account_server"
export LOG_SERVER="logsystem:logsyspassword@tcp(localhost:3306)/log_server"

go mod tidy

go build -o main .

./main