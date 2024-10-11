package main

import (
	"database/sql"
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
        http.Error(w,"FileNotFountException", 404)
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
    http.HandleFunc("/api/add_log", Log_recive)
    http.HandleFunc("/api/add_log_file", Log_ALL_recive)
    http.HandleFunc("/create_User_SYS",create_User_Handle)
    http.HandleFunc("/create_User",Create_User_fromt)
    http.HandleFunc("/create_guest_user", Create_guest_user)
    http.HandleFunc("/user_Login", User_Login)
    http.HandleFunc("/user_Logout", User_Logout)
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
