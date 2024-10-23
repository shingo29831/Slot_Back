package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

var validUsername = "admin"
var validPassword = "password"

type pay_struct struct{
	TableId string `json:"tableId"`
	DepositAmount json.Number `json:"depositAmount"`
	WithdrawalAmoun json.Number `json:"withdrawalAmoun"`
}

func totals_html(w http.ResponseWriter, r *http.Request){
	session, _ := store.Get(r, "auth-session")

	// 認証されていない場合、ログインページにリダイレクト
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	file, err := os.Open("./web/total.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func loginPage(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodGet {
		file, err := os.Open("./web/login_root.html")
		if err != nil {
			http.Error(w,"サーバーエラー", 500)
			return
		}
		defer file.Close()

		buf, err := io.ReadAll(file)
		if err != nil {
			http.Error(w,"サーバーエラー", 500)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write(buf)
	}else if r.Method == http.MethodPost {

		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == validUsername && password == validPassword {
			// セッションを取得
			session, _ := store.Get(r, "auth-session")

			// 認証情報をセッションに保存
			session.Values["authenticated"] = true
			session.Save(r, w)

			// ダッシュボードにリダイレクト
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		} else {
			// 認証失敗
			fmt.Fprintf(w, "Invalid credentials. Please try again.")
		}
	}
}

func pay_root(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet {
		http.Error(w,"権限がありません", http.StatusForbidden)
		return
	}
	session, _ := store.Get(r, "auth-session")

	// 認証されていない場合、ログインページにリダイレクト
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	file, err := os.Open("./web/pay_root.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func dashboardPage(w http.ResponseWriter, r *http.Request) {
	// セッションを取得
	session, _ := store.Get(r, "auth-session")

	// 認証されていない場合、ログインページにリダイレクト
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	file, err := os.Open("./web/dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func logout(w http.ResponseWriter, r *http.Request) {
	// セッションを取得
	session, _ := store.Get(r, "auth-session")

	// 認証情報をクリア
	session.Values["authenticated"] = false
	session.Save(r, w)

	// ログインページにリダイレクト
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func submit_transaction(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Error(w,"権限がありません", http.StatusForbidden)
		return
	}
	session, _ := store.Get(r, "auth-session")

	// 認証されていない場合、ログインページにリダイレクト
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	var req pay_struct 
	
	if err := json.NewDecoder(r.Body).Decode(&req);err != nil{
		ErrorResponse(err.Error(),nil,w)
		return
	}
	
	var user struct{
		Username string
		Token string
		Table string
	}
	err := account_db.QueryRow("select username, TOKEN, table_id from Account_table where table_id = ?",req.TableId).Scan(&user.Username, &user.Token, &user.Table)
	if err != nil {
		ErrorResponse(err.Error(),nil, w)
		return
	}
	with ,_ := req.WithdrawalAmoun.Int64()
	dep, _  := req.DepositAmount.Int64()
 	update_money := dep - with 
	update_type := 0
	if update_money < 0{
		update_money = 2
	}
	query := `
		select money from Account_table WHERE username = ? AND TOKEN = ?
 	`
 	now := 0
 	if err := account_db.QueryRow(query, user.Username,user.Token).Scan(&now); err != nil{
		http.Error(w, "InternalServerError",http.StatusInternalServerError)
		error_print("クリエエラー１%v",err)
		return
 	}
	query = `
	 	select id from session_tokens
	 	where TOKEN = ? AND username = ? AND table_id = ?
 	`
 	var id int
 	if err := account_db.QueryRow(query, user.Token, user.Username, user.Table).Scan(&id); err != nil{
		http.Error(w, "InternalServerError",http.StatusInternalServerError)
		error_print("クリエエラー2%v",err)
		return
 	}
 	query = `
		 Insert Into slot_result_table(time, money, fluctuation, type, session_id, user, table_id)
		 values(?,?,?,?,?,?,?)
 	`
 	if _, err := account_db.Exec(query, time.Now().Format("2006-01-02 15:04:05"), now+int(update_money), update_money, update_type, id, user.Username,user.Table); err != nil{
		http.Error(w,"InternalSerberError", http.StatusInternalServerError)
 	}
	
	_ ,err = account_db.Exec("update Account_table set money = money + ? where table_id = ?",update_money, req.TableId)
	if err != nil {
		ErrorResponse(err.Error(),nil,w)
		return
	}
	log_print("入出金 USER:%s, TableID:%s 額:%d", user.Username,req.TableId, update_money)
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(Message("success","ALL DONE",nil)); err != nil{
		ErrorResponse("error",nil,w)
	}
}

func show_probability(w http.ResponseWriter, r *http.Request){
	
	// セッションを取得
	session, _ := store.Get(r, "auth-session")

	// 認証されていない場合、ログインページにリダイレクト
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	file, err := os.Open("./web/table_probability.html")
	if err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		error_print("table_probabilityエラー:%v", w)
		return
	}
	buf, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		error_print("table_probabilityエラー:%v", w)
		return
	}
	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(200)
	w.Write(buf)
}