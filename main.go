package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Databaseの初期化メソッド（コンストラクタ風）
func NewDatabase(dsn string) (*sql.DB, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }

    // 接続テスト
    err = db.Ping()
    if err != nil {
        return nil, err
    }

    return  db, nil
}


func append_byte(b... []byte)([]byte){
    ans := make([]byte,1024)
    for _, v := range b {
        ans = append(ans, v...)
    }
    return ans
}

func data2json(r *http.Request, v any)(error){
    body, err := io.ReadAll(r.Body)
    fmt.Println(string(body))
    if err != nil {
        return err
    }
    defer r.Body.Close()
	return json.Unmarshal(body,v)
}

// JSONの構造体を定義
type RequestData struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

func handler(w http.ResponseWriter, r *http.Request) {
    if r.Header["Content-Type"][0] != "application/json"{

    }
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

//Create_Userのweb用ハンドラ(唯一まともな使い方をする予定です)
func Create_User_fromt(w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(os.Stderr,"要求:%s",r.Host)
    f,err := os.Open("./web/Create_User.html")
    if err != nil {
        http.Error(w,"サーバーエラー",500)
        return
    }
    defer f.Close()
    data, err := io.ReadAll(f)
    if err != nil{
        http.Error(w,"サーバーエラー",500)
        return
    } 
    w.WriteHeader(200)
    w.Write(data)
}

func fileaccsess(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet || strings.Contains( r.URL.Path, ".."){
        http.Error(w,"server Error",http.StatusBadRequest)
        return
    }
    file, err := os.Open("./web"+r.URL.Path)
    if err != nil{
        http.Error(w,"FileNotFountException <-スペルあってる？", 404)
        return
    } 
    defer file.Close()
    buf, err := io.ReadAll(file)
    if err != nil{
        http.Error(w,"鯖エラー", 500)
        return
    } 
    w.WriteHeader(200)
    w.Write(buf)
}

func main() {
    http.HandleFunc("/script.js" ,fileaccsess)
    http.HandleFunc("/transactions",pay_root)
    http.HandleFunc("/submit-transaction",submit_transaction)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/dashboard", dashboardPage)
	http.HandleFunc("/logout", logout)
    http.HandleFunc("/submit", handler)
    http.HandleFunc("/Log", Log_recive)
    http.HandleFunc("/Log_file", Log_ALL_recive)
    http.HandleFunc("/Create_User_SYS",Create_User_Handle)
    http.HandleFunc("/Create_User",Create_User_fromt)
    http.HandleFunc("/Create_guest_user", Create_guest_user)
    http.HandleFunc("/User_Login", User_Login)
    http.HandleFunc("/User_Logout", User_Logout)
    http.HandleFunc("/update_money", UPDATE_USER_MONEY)
    http.HandleFunc("/get_user_money", GET_USER_MONEY)
    http.HandleFunc("/api/logs",Log_accsess)
    //適当に作った登録完了フォーム（流石に適当がすぎるので、後々治す予定です)
    http.HandleFunc("/Create-success",func (w http.ResponseWriter, r *http.Request)  {
        fmt.Fprintf(w,"登録が完了しました♡")
    })
    fmt.Println("Server is running on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
