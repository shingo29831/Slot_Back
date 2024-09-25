package main

//環境変数代わり
//テスト用

import (
    "encoding/json"
    "fmt"
    "io"
	"os"
)

// JSONData構造体を定義
type JSONData struct {
    data map[string]interface{}
}

// NewJSONData関数でJSONファイルを読み込んで構造体を初期化
func NewJSONData(filePath string) (*JSONData, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    bytes, err := io.ReadAll(file)
    if err != nil {
        return nil, err
    }

    var data map[string]interface{}
    err = json.Unmarshal(bytes, &data)
    if err != nil {
        return nil, err
    }

    return &JSONData{data: data}, nil
}

// Getメソッドで指定されたキーに対応する値を取得
func (jd *JSONData) Get(key string) (string, error) {
    value_interface, exists := jd.data[key]
    if !exists {
        return "", fmt.Errorf("key '%s' not found in JSON data", key)
    }
	value ,ok := value_interface.(string) 
	if !ok {
        return "", fmt.Errorf("key '%s' not found in JSON data", key)
	}
    return value , nil
}
