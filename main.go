package main

import (
	"database/sql"
	"fmt"
	"html/template"
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

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
    httpsURL := "https://" + r.Host + r.URL.String()
    http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
}

func main() {
    go func() {
        httpMux := http.NewServeMux()
        httpMux.HandleFunc("/", redirectToHTTPS)
        fmt.Println("Starting HTTP server on :80")
        if err := http.ListenAndServe(":80", httpMux); err != nil {
            log.Fatalf("HTTP server error: %v", err)
        }
    }()
    init_account_db()
    init_log_DB()
    Logout_user_Array = *initArray()
    mux := http.NewServeMux()
    mux.HandleFunc("/users",users)
    mux.HandleFunc("/api/show_users",show_users)
    mux.HandleFunc("/totals",totals_html)
    mux.HandleFunc("/api/totals",totals)
    mux.HandleFunc("/api/logout_requests", logout_requests)
    mux.HandleFunc("/approve-logout",approve_logout)
    mux.HandleFunc("/styles_css", style_css)
    mux.HandleFunc("/Logout_req", Logout_page)
    mux.HandleFunc("/script.js" ,fileaccsess)
    mux.HandleFunc("/transactions",pay_root)
    mux.HandleFunc("/submit-transaction",submit_transaction)
	mux.HandleFunc("/login", loginPage)
	mux.HandleFunc("/dashboard", dashboardPage)
	mux.HandleFunc("/logout", logout)
    mux.HandleFunc("/api/add_log", Log_recive)
    mux.HandleFunc("/api/add_log_file", Log_ALL_recive)
    mux.HandleFunc("/create_User_SYS",create_User_Handle)
    mux.HandleFunc("/create_User",Create_User_fromt)
    mux.HandleFunc("/create_guest_user", Create_guest_user)
    mux.HandleFunc("/user_Login", User_Login)
    mux.HandleFunc("/user_Logout", User_Logout)
    mux.HandleFunc("/token_exists",Token_exists)
    mux.HandleFunc("/update_money", UPDATE_USER_MONEY)
    mux.HandleFunc("/get_user_money", GET_USER_MONEY)
    mux.HandleFunc("/api/logs",Log_accsess)
    mux.HandleFunc("/table_probability",table_probability)
    mux.HandleFunc("/update-probability",update_probability)
    mux.HandleFunc("/Gettables", GetTables)
    mux.HandleFunc("/tables",show_probability)
    mux.HandleFunc("/",func (w http.ResponseWriter, r *http.Request)  {
        http.Redirect(w,r, "/create_User",http.StatusMovedPermanently);
    })
    //適当に作った登録完了フォーム（流石に適当がすぎるので、後々治す予定です)<-過去の自分　むりかも
    mux.HandleFunc("/Create-success",func (w http.ResponseWriter, r *http.Request)  {
        fmt.Fprintf(w,"登録が完了しました♡")
    })
    mux.Handle("/custom404",http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tmpl := template.Must(template.New("404エラー").Parse(`
            <!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>404エラー</title>
</head>
<body>
    <h1>ここにはなにもないよーん</h1>
    <p>
        ここにたどり着いたものは、社長に連絡するのだ…<br>
        担当は寝ているので起こさないでね☆
    </p>
</body>
</html>
        `))
        tmpl.Execute(w,nil)
    }))

    fmt.Println("Server is running on port 443...")
    err := http.ListenAndServeTLS(":443", "server.crt", "server.key", mux)
    if err != nil {
        panic(err)
    }
}
