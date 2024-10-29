package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
    "net"
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

func MiddlewareIPFilter(next http.Handler)http.Handler{
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
        if err != nil {
            http.Error(w, "Forbidden",http.StatusForbidden)
            return
        }
        log_print("RequestIP:%s",clientIP)
        next.ServeHTTP(w,r)
    })
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
    go func(){
        mux := http.NewServeMux()
        mux.HandleFunc("/create_User_SYS",create_User_Handle)
        mux.HandleFunc("/",func (w http.ResponseWriter, r *http.Request)  {
            http.Redirect(w,r, "/create_User",http.StatusMovedPermanently);
        })
        //適当に作った登録完了フォーム（流石に適当がすぎるので、後々治す予定です)<-過去の自分　むりかも
        mux.HandleFunc("/create_User",Create_User_fromt)
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
        if err := http.ListenAndServeTLS(":443", "server.crt", "server.key", mux); err != nil{
            log.Fatal(err)
        }
    }()
    init_account_db()
    init_log_DB()
    Logout_user_Array = *initArray()
    mux8443 := http.NewServeMux()
    mux8443.HandleFunc("/Bonus_result_append",Best_result_append)
    mux8443.HandleFunc("/users",admins("./web/users.html","ユーザー一覧"))
    mux8443.HandleFunc("/api/show_users",show_users)
    mux8443.HandleFunc("/totals",admins("./web/total.html","total"))
    mux8443.HandleFunc("/api/totals",totals)
    mux8443.HandleFunc("/api/logout_requests", logout_requests)
    mux8443.HandleFunc("/approve-logout",approve_logout)
    mux8443.HandleFunc("/styles_css", style_css)
    mux8443.HandleFunc("/Logout_req", admins("./web/Logout_req.html","ログアウト管理"))
    mux8443.HandleFunc("/script.js" ,fileaccsess)
    mux8443.HandleFunc("/transactions",admins("./web/pay_root.html","入出金処理"))
    mux8443.HandleFunc("/submit-transaction",submit_transaction)
	mux8443.HandleFunc("/login", loginPage)
	mux8443.HandleFunc("/dashboard", admins("./web/dashboard.html","dashboard"))
	mux8443.HandleFunc("/logout", logout)
    mux8443.HandleFunc("/api/add_log", Log_recive)
    mux8443.HandleFunc("/api/add_log_file", Log_ALL_recive)
    mux8443.HandleFunc("/create_guest_user", Create_guest_user)
    mux8443.HandleFunc("/user_Login", User_Login)
    mux8443.HandleFunc("/user_Logout", User_Logout)
    mux8443.HandleFunc("/token_exists",Token_exists)
    mux8443.HandleFunc("/update_money", UPDATE_USER_MONEY)
    mux8443.HandleFunc("/get_user_money", GET_USER_MONEY)
    mux8443.HandleFunc("/api/logs",Log_accsess)
    mux8443.HandleFunc("/table_probability",table_probability)
    mux8443.HandleFunc("/update-probability",update_probability)
    mux8443.HandleFunc("/Gettables", GetTables)
    mux8443.HandleFunc("/tables",admins("./web/table_probability.html","確率管理"))
    mux8443.HandleFunc("/api/result_table",result_table)    
    mux8443.HandleFunc("/bonus_result", admins("./web/Bonus_results.html","ボーナス履歴"))
    

    fmt.Println("Server is running on port 8443...")
    err := http.ListenAndServeTLS(":8443", "server.crt", "server.key", MiddlewareIPFilter(mux8443))
    if err != nil {
        log.Fatal(err)
    }
}
