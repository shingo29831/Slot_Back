package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// JSONの構造体を定義
type RequestData struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

func handler(w http.ResponseWriter, r *http.Request) {
    // POSTメソッドのみを許可
    if r.Method != http.MethodPost {
        http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
        return
    }

    // リクエストボディを読み込む
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Failed to read request body", http.StatusInternalServerError)
        return
    }
    defer r.Body.Close()

    // JSONをパースして構造体にマッピング
    var requestData RequestData
    if err := json.Unmarshal(body, &requestData); err != nil {
        http.Error(w, "Invalid JSON format", http.StatusBadRequest)
        return
    }

    // パースしたデータを使用した処理
    fmt.Fprintf(w, "Received JSON: %+v\n", requestData)
}

func main() {
    http.HandleFunc("/submit", handler)
    fmt.Println("Server is running on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
