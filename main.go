package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Databaseの初期化メソッド（コンストラクタ風）
func NewDatabase(dsn string) (*sql.DB, error) {
    var db *sql.DB
    var err error
    for i := 0; i < 10; i++ {
        db, err = sql.Open("mysql", dsn)
        if err == nil {
            err = db.Ping()
        }

        if err == nil {
            fmt.Println("Connected to the database!")
            break
        }

        log.Printf("Failed to connect to database (attempt %d/10): %s", i+1, err)
        time.Sleep(3 * time.Second)  // 3秒待機してリトライ
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
    w.WriteHeader(http.StatusOK)
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
    w.WriteHeader(http.StatusOK)
    w.Write(buf)
}

func main() {
    init_account_db()
    init_log_DB()
    Logout_user_Array = *initArray()
    http.HandleFunc("/totals",totals_html)
    http.HandleFunc("/api/totals",totals)
    http.HandleFunc("/api/logout_requests", logout_requests)
    http.HandleFunc("/approve-logout",approve_logout)
    http.HandleFunc("/styles_css", style_css)
    http.HandleFunc("/Logout_req", Logout_page)
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
    http.HandleFunc("/token_exists",Token_exists)
    http.HandleFunc("/update_money", UPDATE_USER_MONEY)
    http.HandleFunc("/get_user_money", GET_USER_MONEY)
    http.HandleFunc("/api/logs",Log_accsess)
    http.HandleFunc("/table_probability",table_probability)
    http.HandleFunc("/update-probability",update_probability)
    http.HandleFunc("/Gettables", GetTables)
    http.HandleFunc("/tables",show_probability)
    http.HandleFunc("/",func (w http.ResponseWriter, r *http.Request)  {
        http.Redirect(w,r, "/create_User",http.StatusMovedPermanently);
    })
    //適当に作った登録完了フォーム（流石に適当がすぎるので、後々治す予定です)<-過去の自分　むりかも
    http.HandleFunc("/Create-success",func (w http.ResponseWriter, r *http.Request)  {
        fmt.Fprintf(w,"登録が完了しました♡")
    })
    fmt.Println("Server is running on port 8443...")
    err := http.ListenAndServeTLS(":8443", "server.crt", "server.key", nil)
    if err != nil {
        panic(err)
    }
}
